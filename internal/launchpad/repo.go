package launchpad

import (
	"fmt"
	"log"

	"github.com/jnsgruk/releasegen/internal/repos"
)

// Repository represents a single Launchpad git Repository
type Repository struct {
	Details      repos.RepoDetails
	projectGroup string
	readme       *repos.Readme
}

// Process populates the Repository with details of its tags, default branch, and commits
func (r *Repository) Process() error {
	log.Printf("processing launchpad repo: %s/%s\n", r.projectGroup, r.Details.Name)

	project := &Project{Name: r.Details.Name}

	defaultBranch, err := project.DefaultBranch()
	if err != nil {
		return err
	}
	r.Details.DefaultBranch = defaultBranch

	newCommits, err := project.NewCommits()
	if err != nil {
		return err
	}
	r.Details.NewCommits = newCommits

	tags, err := project.Tags()
	if err != nil {
		return err
	}

	if len(tags) == 0 {
		return nil
	}

	//Iterate over the tags in the launchpad repo
	for _, t := range tags {
		r.Details.Releases = append(r.Details.Releases, &repos.Release{
			Id:         t.Timestamp.Unix(),
			Version:    t.Name,
			Timestamp:  t.Timestamp.Unix(),
			Title:      t.Name,
			Body:       "",
			Url:        fmt.Sprintf("%s/tag/?h=%s", r.Details.Url, t.Name),
			CompareUrl: fmt.Sprintf("%s/diff/?id=%s&id2=%s", r.Details.Url, t.Commit, r.Details.DefaultBranch),
		})
	}

	// Get contents of the README as a string
	readmeContent, err := project.fetchReadmeContent()
	if err != nil {
		// The rest of this method depends on the README content, so if we don't get
		// any README content, we may as well return early
		return err
	}

	// Parse contents of README to identify associated snaps and charms
	r.readme = &repos.Readme{Body: readmeContent}
	r.Details.Snap = r.readme.LinkedSnap()
	r.Details.Charm = r.readme.LinkedCharm()
	return err
}
