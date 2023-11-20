package launchpad

import (
	"context"
	"fmt"

	"github.com/jnsgruk/releasegen/internal/repos"
)

// Repository represents a single Launchpad git Repository.
type Repository struct {
	Details      repos.RepoDetails
	projectGroup string
}

// Process populates the Repository with details of its tags, default branch, and commits.
func (r *Repository) Process(ctx context.Context) error {
	project := &Project{Name: r.Details.Name}

	defaultBranch, err := project.DefaultBranch(ctx)
	if err != nil {
		return err
	}

	r.Details.DefaultBranch = defaultBranch

	newCommits, err := project.NewCommits(ctx)
	if err != nil {
		return err
	}

	r.Details.NewCommits = newCommits

	tags, err := project.Tags(ctx)
	if err != nil {
		return err
	}

	if len(tags) == 0 {
		return nil
	}

	// Iterate over the tags in the launchpad repo.
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

	// Populate the repository's README from Launchpad, parse any linked snaps, charms or CI actions.
	err = r.parseReadme(ctx, project)
	if err != nil {
		return err
	}

	return err
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
