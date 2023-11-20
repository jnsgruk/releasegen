package launchpad

import (
	"context"
	"fmt"

	"github.com/jnsgruk/releasegen/internal/repos"
)

// Repository represents a single Launchpad git Repository.
type Repository struct {
	Details      repos.RepoDetails
	project      *Project
	projectGroup string
}

// Process populates the Repository with details of its tags, default branch, and commits.
func (r *Repository) Process(ctx context.Context) error {
	r.project = &Project{Name: r.Details.Name}

	// Iterate over the releases in the Launchpad repo and add them to our repository's details.
	err := r.processReleases(ctx)
	if err != nil {
		return err
	}

	// Calculate the number of commits since the latest release.
	err = r.processCommitsSinceRelease(ctx)
	if err != nil {
		return err
	}

	// Populate the repository's README from Launchpad, parse any linked snaps, charms or CI actions.
	err = r.parseReadme(ctx, r.project)
	if err != nil {
		return err
	}

	return err
}

// processReleases fetches a repository's tags from Launchpad, then populates r.Details.Releases
// with the information in the relevant format for releasegen.
func (r *Repository) processReleases(ctx context.Context) error {
	tags, err := r.project.Tags(ctx)
	if err != nil {
		return err
	}

	if len(tags) == 0 {
		return nil
	}

	err = r.processDefaultBranch(ctx)
	if err != nil {
		return err
	}

	// Iterate over the tags in the Launchpad repo.
	for _, t := range tags {
		r.Details.Releases = append(r.Details.Releases, &repos.Release{
			ID:         t.Timestamp.Unix(),
			Version:    t.Name,
			Timestamp:  t.Timestamp.Unix(),
			Title:      t.Name,
			Body:       "",
			URL:        fmt.Sprintf("%s/tag/?h=%s", r.Details.URL, t.Name),
			CompareURL: fmt.Sprintf("%s/diff/?id=%s&id2=%s", r.Details.URL, t.Commit, r.Details.DefaultBranch),
		})
	}

	return nil
}

// processDefaultBranch gets the name of the default branch in the Launchpad repo and populates it
// on the repository.
func (r *Repository) processDefaultBranch(ctx context.Context) error {
	defaultBranch, err := r.project.DefaultBranch(ctx)
	if err != nil {
		return err
	}

	r.Details.DefaultBranch = defaultBranch

	return nil
}

// processCommitsSinceRelease calculates the number of commits that have occurred on the default
// branch of the repository since the last release, and populates the information in r.Details.
func (r *Repository) processCommitsSinceRelease(ctx context.Context) error {
	newCommits, err := r.project.NewCommits(ctx)
	if err != nil {
		return err
	}

	r.Details.NewCommits = newCommits

	return nil
}

// parseReadme is a helper function to fetch the README from a Launchpad repository and return
// its contents as a string.
func (r *Repository) parseReadme(ctx context.Context, project *Project) error {
	// Get contents of the README as a string.
	readmeContent, err := project.fetchReadmeContent(ctx)
	if err != nil {
		return err
	}

	// Parse contents of README to identify associated snaps and charms.
	readme := &repos.Readme{Body: readmeContent}
	r.Details.Snap = readme.LinkedSnap(ctx)
	r.Details.Charm = readme.LinkedCharm(ctx)

	return nil
}
