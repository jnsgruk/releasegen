package github

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strconv"

	"github.com/gomarkdown/markdown"
	gh "github.com/google/go-github/v54/github"
	"github.com/jnsgruk/releasegen/internal/repos"
)

const githubReleasesPerRepo = 3

var (
	// prRegexp is used to find Github PR URLs in blocks of Markdown/HTML.
	prRegexp = regexp.MustCompile(`(https://github.com/canonical/.+/pull/([0-9]+))`)
	// prShortRegexp is used to find Github PR's mentioned as '#<PR>' (#34 for example).
	prShortRegexp = regexp.MustCompile(`#([0-9]+)`)
	// userRegexp is used to find Twitter/Github style mentions such as '@JoeBloggs' in Markdown/HTML.
	userRegexp = regexp.MustCompile(`(\A|\s)@([\w-_]+)`)
	// errFetchReadme is returned when a README could not be fetched or parsed.
	errFetchReadme = errors.New("error getting README for repo")
)

// Repository represents a single Github Repository.
type Repository struct {
	Details       repos.RepoDetails
	org           string // The Github Org that owns the repo.
	team          string // The Github team, within the org, that has rights over the repo.
	client        *gh.Client
	defaultBranch string
}

// Process populates the Repository with details of its releases, and commits.
func (r *Repository) Process(ctx context.Context) error {
	// Skip archived repositories.
	if r.IsArchived(ctx) {
		return nil
	}

	// Iterate over the releases in the Github repo and add them to our repository's details.
	err := r.processReleases(ctx)
	if err != nil {
		return err
	}

	if len(r.Details.Releases) > 0 {
		// Calculate the number of commits since the latest release.
		err := r.processCommitsSinceRelease(ctx)
		if err != nil {
			return err
		}
	} else {
		// If there are no releases, get the latest commit instead.
		err := r.processCommits(ctx)
		if err != nil {
			return err
		}
	}

	// Populate the repository's README from Github, parse any linked snaps, charms or CI actions.
	err = r.parseReadme(ctx)
	if err != nil {
		return err
	}

	return nil
}

// IsArchived indicates whether or not the repository is marked as archived on Github.
func (r *Repository) IsArchived(ctx context.Context) bool {
	repoObject, _, err := r.client.Repositories.Get(ctx, r.org, r.Details.Name)
	if err != nil {
		return false
	}

	return repoObject.GetArchived()
}

// parseReadme is a helper function to fetch the README from a Github repository and return
// its contents as a string.
func (r *Repository) parseReadme(ctx context.Context) error {
	githubReadme, _, err := r.client.Repositories.GetReadme(ctx, r.org, r.Details.Name, nil)
	if err != nil {
		return errFetchReadme
	}

	content, err := githubReadme.GetContent()
	if err != nil {
		return errFetchReadme
	}

	// Parse contents of README to identify associated Github Workflows, snaps, charms.
	readme := &repos.Readme{Body: content}
	r.Details.CiActions = readme.GithubActions()
	r.Details.Snap = readme.LinkedSnap(ctx)
	r.Details.Charm = readme.LinkedCharm(ctx)

	return nil
}

// processReleases fetches a repository's releases from Github, then populates r.Details.Releases
// with the information in the relevant format for releasegen.
func (r *Repository) processReleases(ctx context.Context) error {
	opts := &gh.ListOptions{PerPage: githubReleasesPerRepo}

	releases, _, err := r.client.Repositories.ListReleases(ctx, r.org, r.Details.Name, opts)
	if err != nil {
		return errors.New("error listing releases for repo")
	}

	for _, rel := range releases {
		r.Details.Releases = append(r.Details.Releases, &repos.Release{
			ID:         rel.GetID(),
			Version:    rel.GetTagName(),
			Timestamp:  rel.CreatedAt.Time.Unix(),
			Title:      rel.GetName(),
			Body:       renderReleaseBody(rel.GetBody(), r),
			URL:        rel.GetHTMLURL(),
			CompareURL: fmt.Sprintf("%s/compare/%s...%s", r.Details.URL, rel.GetTagName(), r.defaultBranch),
		})
	}

	return nil
}

// processCommitsSinceRelease calculates the number of commits that have occurred on the default
// branch of the repository since the last release, and populates the information in r.Details.
func (r *Repository) processCommitsSinceRelease(ctx context.Context) error {
	opts := &gh.ListOptions{PerPage: githubReleasesPerRepo}
	// Add the commit delta between last release and default branch.
	comparison, _, err := r.client.Repositories.CompareCommits(
		ctx, r.org, r.Details.Name, r.Details.Releases[0].Version, r.defaultBranch, opts,
	)
	if err != nil {
		return errors.New("error getting commit comparison for release")
	}

	r.Details.NewCommits = *comparison.TotalCommits

	return nil
}

// processCommits fetches the latest 3 commits to a repository and populates them into the repo
// struct in the case that there are no releases identified.
func (r *Repository) processCommits(ctx context.Context) error {
	opts := &gh.CommitsListOptions{ListOptions: gh.ListOptions{PerPage: githubReleasesPerRepo}}

	// If there are no releases, get the latest commit instead.
	commits, _, err := r.client.Repositories.ListCommits(ctx, r.org, r.Details.Name, opts)
	if err != nil {
		return errors.New("error listing commits for repository")
	}

	// Iterate over the commits and append them to r.Details.Commits
	for _, commit := range commits {
		ts := commit.GetCommit().GetAuthor().GetDate()
		r.Details.Commits = append(r.Details.Commits, &repos.Commit{
			Sha:       commit.GetSHA(),
			Author:    commit.GetCommit().GetAuthor().GetName(),
			Timestamp: ts.GetTime().Unix(),
			Message:   renderReleaseBody(commit.GetCommit().GetMessage(), r),
			URL:       commit.GetHTMLURL(),
		})
	}

	return nil
}

// renderReleaseBody transforms a Markdown string from a Github Release into HTML.
func renderReleaseBody(body string, repo *Repository) string {
	// Preprocess any Pull Request links in the Markdown body.
	body = prRegexp.ReplaceAllString(body, `<a href="${1}">#${2}</a>`)
	// Preprocess any user mentions in the Markdown body.
	body = userRegexp.ReplaceAllString(body, `${1}<a href="https://github.com/${2}">@${2}</a>`)

	// Preprocess any mentions of PR's using the shorthand notation #<PR>
	body = prShortRegexp.ReplaceAllStringFunc(body, func(s string) string {
		// Grab the number of the PR; if we fail to parse the number just return the original str.
		num, err := strconv.Atoi(s[1:])
		if err != nil {
			return s
		}

		// Construct the possible URL of the PR.
		url := fmt.Sprintf("https://github.com/%s/%s/pull/%d", repo.org, repo.Details.Name, num)

		// Make a HEAD request to the proposed PR URL, if it's not 200 OK, assume the PR reference
		// is a false positive and return the original string.
		res, err := http.Head(url) //nolint
		if err != nil || res.StatusCode != http.StatusOK {
			return s
		}

		// All good! We found the PR so return the link.
		return fmt.Sprintf(`<a href="%s">#%d</a>`, url, num)
	})

	// Render the Markdown to HTML.
	md := []byte(body)
	normalised := markdown.NormalizeNewlines(md)

	return string(markdown.ToHTML(normalised, nil, nil))
}
