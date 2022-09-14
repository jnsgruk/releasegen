package releases

import (
	"context"
	"fmt"
	"log"

	"github.com/google/go-github/v47/github"
	"github.com/jnsgruk/releasegen/internal/launchpad"
)

// RepositoryInfo represents the serialisable form of a Repository for the Report
type RepositoryInfo struct {
	Name          string     `json:"name"`
	DefaultBranch string     `json:"default_branch"`
	NewCommits    int        `json:"new_commits"`
	Url           string     `json:"url"`
	Releases      []*Release `json:"releases"`
}

// Repository is an interface that provides common methods for different types of repository
type Repository interface {
	Info() RepositoryInfo
	Process() error
}

// githubRepository represents a single Github repository
type githubRepository struct {
	info RepositoryInfo
	org  string // The Github Org that owns the repo
	team string // The Github team, within the org, that has rights over the repo
}

// launchpadRepository represents a single Launchpad git repository
type launchpadRepository struct {
	info           RepositoryInfo
	lpProjectGroup string // The project group the repo belongs to
}

// NewGithubRepository fetches and parses release information for a single Github repository
func NewGithubRepository(repo *github.Repository, team string, org string) Repository {
	// Create a repository to represent the Github Repo
	r := &githubRepository{
		info: RepositoryInfo{
			Name:          *repo.Name,
			DefaultBranch: *repo.DefaultBranch,
			Url:           *repo.HTMLURL,
		},
		org:  org,
		team: team,
	}
	return r
}

// Info returns a serialisable representation of the repository
func (r *githubRepository) Info() RepositoryInfo { return r.info }

// Process populates the Repository with details of its releases, and commits
func (r *githubRepository) Process() error {
	log.Printf("processing github repo: %s/%s/%s\n", r.org, r.team, r.info.Name)

	client := githubClient()
	ctx := context.Background()
	opts := &github.ListOptions{PerPage: 3}

	releases, _, err := client.Repositories.ListReleases(ctx, r.org, r.info.Name, opts)
	if err != nil {
		return fmt.Errorf("error listing releases for repo: %s/%s/%s: %v", r.org, r.team, r.info.Name, err)
	} else if len(releases) == 0 {
		return nil
	}

	// Iterate over the releases in the Github repo
	for _, rel := range releases {
		r.info.Releases = append(r.info.Releases, NewRelease(
			rel.GetID(),
			rel.GetTagName(),
			rel.CreatedAt.Time,
			rel.GetName(),
			rel.GetBody(),
			rel.GetHTMLURL(),
			fmt.Sprintf("%s/compare/%s...%s", r.info.Url, rel.GetTagName(), r.info.DefaultBranch),
		))
	}

	// Add the commit delta between last release and default branch
	comparison, _, err := client.Repositories.CompareCommits(ctx, r.org, r.info.Name, r.info.Releases[0].Version, r.info.DefaultBranch, opts)
	if err != nil {
		return fmt.Errorf("error getting commit comparison for release %s in %s/%s/%s", r.info.Releases[0].Version, r.org, r.team, r.info.Name)
	}

	r.info.NewCommits = *comparison.TotalCommits

	return nil
}

// NewLaunchpadRepository creates a new representation for a Launchpad Git repo
func NewLaunchpadRepository(project launchpad.ProjectEntry, lpGroup string) Repository {
	// Create a repository to represent the Launchpad project
	r := &launchpadRepository{
		info: RepositoryInfo{
			Name:          project.Name,
			DefaultBranch: "",
			Url:           fmt.Sprintf("https://git.launchpad.net/%s", project.Name),
		},
		lpProjectGroup: lpGroup,
	}
	return r
}

// Info returns a serialisable representation of the repository
func (r *launchpadRepository) Info() RepositoryInfo { return r.info }

// Process populates the Repository with details of its tags, default branch, and commits
func (r *launchpadRepository) Process() error {
	log.Printf("processing launchpad repo: %s/%s\n", r.lpProjectGroup, r.info.Name)

	project := launchpad.NewProject(r.info.Name)

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
		r.info.Releases = append(r.info.Releases, NewRelease(
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
