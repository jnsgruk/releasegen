package github

import (
	"context"
	"fmt"
	"log"
	"regexp"

	"github.com/gomarkdown/markdown"
	gh "github.com/google/go-github/v54/github"
	"github.com/jnsgruk/releasegen/internal/repos"
)

var (
	// prRegexp is used to find Github PR URLs in blocks of Markdown/HTML.
	prRegexp = regexp.MustCompile(`(https://github.com/canonical/.+/pull/([0-9]+))`)
	// userRegexp is used to find Twitter/Github style mentions such as '@JoeBloggs' in Markdown/HTML.
	userRegexp = regexp.MustCompile(`(\A|\s)@([\w-_]+)`)
)

// Repository represents a single Github Repository.
type Repository struct {
	Details repos.RepoDetails
	org     string // The Github Org that owns the repo.
	team    string // The Github team, within the org, that has rights over the repo.
	readme  *repos.Readme
}

// Process populates the Repository with details of its releases, and commits.
func (r *Repository) Process() error {
	log.Printf("processing github repo: %s/%s/%s\n", r.org, r.team, r.Details.Name)

	// Skip archived repositories.
	if r.IsArchived() {
		return nil
	}

	client := githubClient()
	ctx := context.Background()
	opts := &gh.ListOptions{PerPage: 3}

	// Get the releases from the repo.
	releases, _, err := client.Repositories.ListReleases(ctx, r.org, r.Details.Name, opts)
	if err != nil {
		return fmt.Errorf(
	}

	if len(releases) > 0 {
		// Iterate over the releases in the Github repo.
		for _, rel := range releases {
			r.Details.Releases = append(r.Details.Releases, &repos.Release{
				Id:         rel.GetID(),
				Version:    rel.GetTagName(),
				Timestamp:  rel.CreatedAt.Time.Unix(),
				Title:      rel.GetName(),
				Body:       renderReleaseBody(rel.GetBody()),
				Url:        rel.GetHTMLURL(),
				CompareUrl: fmt.Sprintf("%s/compare/%s...%s", r.Details.Url, rel.GetTagName(), r.Details.DefaultBranch),
			})
		}

		// Add the commit delta between last release and default branch.
		comparison, _, err := client.Repositories.CompareCommits(
			ctx, r.org, r.Details.Name, r.Details.Releases[0].Version, r.Details.DefaultBranch, opts,
		)

		if err != nil {
			return fmt.Errorf(
				"error getting commit comparison for release '%s' in '%s/%s/%s': %s",
				r.Details.Releases[0].Version, r.org, r.team, r.Details.Name, err.Error(),
			)
		}

		r.Details.NewCommits = *comparison.TotalCommits
	} else {
		// If there are no releases, get the latest commit instead.
		commits, _, err := client.Repositories.ListCommits(ctx, r.org, r.Details.Name, nil)
		// If there is at least one commit, add it as a release.
		if err == nil {
			com := commits[0]
			ts := com.GetCommit().GetAuthor().GetDate()
			r.Details.Commits = append(r.Details.Commits, &repos.Commit{
				Sha:       com.GetSHA(),
				Author:    com.GetCommit().GetAuthor().GetName(),
				Timestamp: ts.GetTime().Unix(),
				Message:   com.GetCommit().GetMessage(),
				Url:       com.GetHTMLURL(),
			})
		}
	}

	// Get contents of the README as a string.
	readmeContent, err := r.fetchReadmeContent()
	if err != nil {
		// The rest of this method depends on the README content, so if we don't get
		// any README content, we may as well return early.
		return err
	}

	// Parse contents of README to identify associated Github Workflows, snaps, charms.
	r.readme = &repos.Readme{Body: readmeContent}
	r.Details.CiActions = r.readme.GithubActions()
	r.Details.Snap = r.readme.LinkedSnap()
	r.Details.Charm = r.readme.LinkedCharm()

	return nil
}

// IsArchived indicates whether or not the repository is marked as archived on Github.
func (r *Repository) IsArchived() bool {
	client := githubClient()
	// Check if the repository is archived
	repoObject, _, err := client.Repositories.Get(context.Background(), r.org, r.Details.Name)
	if err != nil {
		log.Printf(
			"error while checking archived status for repo '%s/%s/%s': %s",
			r.org, r.team, r.Details.Name, err.Error(),
		)

		return false
	}

	return repoObject.GetArchived()
}

// fetchReadmeContent is a helper function to fetch the README from a Github repository and return
// its contents as a string.
func (r *Repository) fetchReadmeContent() (string, error) {
	client := githubClient()

	readme, _, err := client.Repositories.GetReadme(context.Background(), r.org, r.Details.Name, nil)
	if err != nil {
		return "", fmt.Errorf("error getting README for repo '%s/%s': %s", r.org, r.Details.Name, err.Error())
	}

	content, err := readme.GetContent()
	if err != nil {
		return "", fmt.Errorf("error getting README content for repo '%s/%s'", r.org, r.Details.Name)
	}

	return content, nil
}

// renderReleaseBody transforms a Markdown string from a Github Release into HTML.
func renderReleaseBody(body string) string {
	// Preprocess any Pull Request links in the Markdown body.
	body = prRegexp.ReplaceAllString(body, `<a href="${1}">#${2}</a>`)
	// Preprocess any user mentions in the Markdown body.
	body = userRegexp.ReplaceAllString(body, `${1}<a href="http://github.com/${2}">@${2}</a>`)

	// Render the Markdown to HTML.
	md := []byte(body)
	normalised := markdown.NormalizeNewlines(md)

	return string(markdown.ToHTML(normalised, nil, nil))
}
