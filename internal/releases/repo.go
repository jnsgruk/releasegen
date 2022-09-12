package releases

import (
	"context"
	"fmt"

	"github.com/google/go-github/v47/github"
	"github.com/jnsgruk/releasegen/internal/config"
	"github.com/jnsgruk/releasegen/internal/launchpad"
)

// Repository refers to a version control Repository on either Github or Launchpad
type Repository struct {
	Name          string    `json:"name"`
	DefaultBranch string    `json:"default_branch"`
	NewCommits    int       `json:"new_commits"`
	Url           string    `json:"url"`
	Releases      []Release `json:"releases"`
}

// NewGithubRepository fetches and parses release information for a single Github repository
func NewGithubRepository(repo *github.Repository, team string, org config.GithubOrg) (*Repository, error) {
	client := githubClient()
	ctx := context.Background()
	opts := &github.ListOptions{PerPage: 3}

	// Create a repository to represent the Github Repo
	r := &Repository{
		Name:          *repo.Name,
		DefaultBranch: *repo.DefaultBranch,
		NewCommits:    0,
		Url:           *repo.HTMLURL,
		Releases:      []Release{},
	}

	releases, _, err := client.Repositories.ListReleases(ctx, org.Name, r.Name, opts)
	if err != nil {
		return nil, fmt.Errorf("error listing releases for repo: %s/%s/%s: %v", org.Name, team, r.Name, err)
	} else if len(releases) == 0 {
		return nil, nil
	}

	// Iterate over the releases in the Github repo
	for _, rel := range releases {
		r.Releases = append(r.Releases, NewRelease(
			rel.GetID(),
			rel.GetTagName(),
			rel.CreatedAt.Time,
			rel.GetName(),
			rel.GetBody(),
			rel.GetHTMLURL(),
			fmt.Sprintf("%s/compare/%s...%s", r.Url, rel.GetTagName(), r.DefaultBranch),
		))
	}

	// Add the commit delta between last release and default branch
	comparison, _, err := client.Repositories.CompareCommits(ctx, org.Name, r.Name, r.Releases[0].Version, r.DefaultBranch, opts)
	if err != nil {
		return nil, fmt.Errorf("error getting commit comparison for release %s in %s/%s/%s", r.Releases[0].Version, org.Name, team, r.Name)
	}

	r.NewCommits = *comparison.TotalCommits
	return r, nil
}

// NewLaunchpadRepository fetches and parses release information for a single Launchpad project
func NewLaunchpadRepository(project launchpad.Project, lpGroup string) (*Repository, error) {
	if project.Vcs != "Git" {
		return nil, nil
	}

	// Create a repository to represent the Launchpad project
	r := &Repository{
		Name:          project.Name,
		DefaultBranch: "",
		NewCommits:    project.NewCommits(),
		Url:           fmt.Sprintf("https://git.launchpad.net/%s", project.Name),
		Releases:      []Release{},
	}

	defaultBranch, err := project.DefaultBranch()
	if err != nil {
		return nil, err
	}
	r.DefaultBranch = defaultBranch

	tags, err := project.Tags()
	if err != nil {
		return nil, err
	}

	if len(tags) == 0 {
		return nil, nil
	}

	//Iterate over the tags in the launchpad repo
	for _, t := range tags {
		r.Releases = append(r.Releases, NewRelease(
			t.Timestamp.Unix(),
			t.Name,
			*t.Timestamp,
			t.Name,
			"",
			fmt.Sprintf("%s/tag/?h=%s", r.Url, t.Name),
			fmt.Sprintf("%s/diff/?id=%s&id2=%s", r.Url, t.Commit, r.DefaultBranch),
		))
	}

	return r, nil
}
