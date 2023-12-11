package gitea

import (
	"context"
	"errors"
	"fmt"
	"path"

	"code.gitea.io/sdk/gitea"
	"github.com/gomarkdown/markdown"
	"github.com/jnsgruk/releasegen/internal/repos"
)

const releasesPerRepo = 3

// Repository represents a single repository. Note that this might be a Gitea
// Repository or it might be a folder in a Gitea Repository if the Repository
// is a monorepo (`folder` will be non-empty in that case).
type Repository struct {
	Details       repos.RepoDetails
	org           string // The Gitea Org that owns the repo.
	client        *gitea.Client
	defaultBranch string
	folder        string
}

// Process populates the Repository with details of its releases, and commits.
func (r *Repository) Process() error {
	// Iterate over the releases in the Gitea repo and add them to our
	// repository's details.
	err := r.processReleases()
	if err != nil {
		return err
	}

	if len(r.Details.Releases) > 0 {
		// Calculate the number of commits since the latest release.
		err = r.processCommitsSinceRelease()
	} else {
		// If there are no releases, get the latest commit instead.
		err = r.processCommits()
	}
	if err != nil {
		return err
	}

	// Populate the repository's README from Gitea and parse any linked snaps
	// or charms.
	err = r.parseReadme()
	if err != nil {
		return err
	}

	return nil
}

// parseReadme is a helper function to fetch the README from a Gitea repository
// and populates any linked Charms or Snaps in r.Details.
func (r *Repository) parseReadme() error {
	names := []string{"README.md", "README.rst"}
	var bytes []byte
	for _, name := range names {
		var err error
		var fileName string
		path := name
		if r.folder != "" {
			path = path.Join(r.folder, name)
		}
		bytes, _, err = r.client.GetFile(r.org, r.Details.Name, r.defaultBranch, fileName, false)
		if err == nil {
			break
		}
	}
	if len(bytes) == 0 {
		return errors.New("error getting README for repo")
	}

	content := string(bytes)

	// Parse contents of README to identify associated snaps, charms.
	readme := &repos.Readme{Body: content}
	ctx := context.Background()
	r.Details.Snap = readme.LinkedSnap(ctx)
	r.Details.Charm = readme.LinkedCharm(ctx)

	return nil
}

// processReleases fetches a repository's releases from Gitea, then populates
// r.Details.Releases with the information in the relevant format for releasegen.
func (r *Repository) processReleases() error {
	// TODO: This currently gets all releases across the entire repository, even
	// in a monorepo. It's not clear what happens with releases in a Gitea
	// monorepo - for example, are they not used at all? Do they all start with
	// the name of the 'subrepo'?

	opts := gitea.ListReleasesOptions{
		ListOptions: gitea.ListOptions{PageSize: releasesPerRepo},
		IsDraft: gitea.OptionalBool(false),
		IsPreRelease: gitea.OptionalBool(false)
	}

	releases, _, err := r.client.ListReleases(r.org, r.Details.Name, opts)
	if err != nil {
		return err
	}

	for _, rel := range releases {
		r.Details.Releases = append(r.Details.Releases, &repos.Release{
			ID:         rel.ID,
			Version:    rel.TagName,
			Timestamp:  rel.PublishedAt.Unix(),
			Title:      rel.Title,
			Body:       renderReleaseBody(rel.Note, r),
			URL:        rel.URL,
			CompareURL: fmt.Sprintf(
				"%s/compare/%s...%s", r.Details.URL, rel.TagName, r.defaultBranch
			),
		})
	}

	return nil
}

// processCommitsSinceRelease calculates the number of commits that have
// occurred on the default branch of the repository since the last release, and
// populates the information in r.Details.
func (r *Repository) processCommitsSinceRelease() error {
	// TODO: This does not currently handle monorepos - see note about releases
	// in `processReleases`. If there are releases in Gitea with a monorepo,
	// then this function maybe needs to restrict the commits to ones where the
	// tree overlaps - but maybe we actually want to know how many commits
	// overall, because we want to count things like common code being adjusted.

	if !len(r.Details.Releases) {
		return errors.New("processCommitsSinceRelease cannot be called if there are no releases!")
	}

	// Add the commit delta between last release and default branch.
	latestRelease := r.Details.Releases[len(r.Details.Releases)-1]
	opts := gitea.ListCommitOptions{
		ListOptions: gitea.ListOptions{PageSize: giteaReleasesPerRepo},
		SHA: latestRelease.Version,
		Path: ""
	}
	commits, _, err := r.client.ListRepoCommits(r.org, r.Details.Name, opts)
	if err != nil {
		return err
	}

	r.Details.NewCommits = len(commits)

	return nil
}

// processCommits fetches the latest commits to a repository and populates them
// into the repo struct in the case that there are no releases identified.
func (r *Repository) processCommits() error {
	// TODO: This does not currently handle monorepos. It's not clear which
	// commits should be counted in this case - only ones that have a tree that
	// overlaps with the subfolder? All commits, as now? If there's common code,
	// then all commits is probably truest to the single-repo meaning, but if
	// there are commits that are only in a separate charm, then those probably
	// should not be included.

	opts := gitea.ListCommitOptions{
		ListOptions: gitea.ListOptions{PageSize: giteaReleasesPerRepo},
		SHA: r.defaultBranch,
		Path: ""
	}

	commits, _, err := r.client.ListRepoCommits(r.org, r.Details.Name, opts)
	if err != nil {
		return err
	}

	// Iterate over the commits and append them to r.Details.Commits
	for _, commit := range commits {
		// Some commits don't have an author.
		name := ""
		if commit.Author != nil {
			name = commit.Author.FullName
		}
		r.Details.Commits = append(r.Details.Commits, &repos.Commit{
			Sha:       commit.CommitMeta.SHA,
			Author:    name,
			Timestamp: commit.CommitMeta.Created.Unix(),
			Message:   renderReleaseBody(commit.RepoCommit.Message, r),
			URL:       commit.HTMLURL,
		})
	}

	return nil
}

// renderReleaseBody transforms a Markdown string from a Gitea Release into HTML.
func renderReleaseBody(body string, repo *Repository) string {
	// Render the Markdown to HTML.
	normalised := markdown.NormalizeNewlines([]byte(body))

	return string(markdown.ToHTML(normalised, nil, nil))
}
