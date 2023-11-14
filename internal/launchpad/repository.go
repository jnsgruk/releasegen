package launchpad

import (
	"fmt"
	"log"

	"github.com/jnsgruk/releasegen/internal/repositories"
)

// launchpadRepository represents a single Launchpad git repository
type launchpadRepository struct {
	info           repositories.RepositoryInfo
	lpProjectGroup string // The project group the repo belongs to
}

// NewLaunchpadRepository creates a new representation for a Launchpad Git repo
func NewLaunchpadRepository(project ProjectEntry, lpGroup string) repositories.Repository {
	// Create a repository to represent the Launchpad project
	r := &launchpadRepository{
		info: repositories.RepositoryInfo{
			Name:          project.Name,
			DefaultBranch: "",
			Url:           fmt.Sprintf("https://git.launchpad.net/%s", project.Name),
		},
		lpProjectGroup: lpGroup,
	}
	return r
}

// Info returns a serialisable representation of the repository
func (r *launchpadRepository) Info() repositories.RepositoryInfo { return r.info }

// Process populates the Repository with details of its tags, default branch, and commits
func (r *launchpadRepository) Process() error {
	log.Printf("processing launchpad repo: %s/%s\n", r.lpProjectGroup, r.info.Name)

	project := NewProject(r.info.Name)

	defaultBranch, err := project.DefaultBranch()
	if err != nil {
		return err
	}
	r.info.DefaultBranch = defaultBranch

	newCommits, err := project.NewCommits()
	if err != nil {
		return err
	}
	r.info.NewCommits = newCommits

	tags, err := project.Tags()
	if err != nil {
		return err
	}

	if len(tags) == 0 {
		return nil
	}

	//Iterate over the tags in the launchpad repo
	for _, t := range tags {
		r.info.Releases = append(r.info.Releases, repositories.NewRelease(
			t.Timestamp.Unix(),
			t.Name,
			*t.Timestamp,
			t.Name,
			"",
			fmt.Sprintf("%s/tag/?h=%s", r.info.Url, t.Name),
			fmt.Sprintf("%s/diff/?id=%s&id2=%s", r.info.Url, t.Commit, r.info.DefaultBranch),
		))
	}
	return err
}
