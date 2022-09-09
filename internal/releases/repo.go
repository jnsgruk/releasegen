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
func NewGithubRepository(ghRepo *github.Repository, ghOrg config.GithubOrg, ghTeam string) (*Repository, error) {
	client := getGithubClient()
	ctx := context.Background()

	// Create a repository to represent the Github Repo
	repo := &Repository{
		Name:          *ghRepo.Name,
		DefaultBranch: *ghRepo.DefaultBranch,
		NewCommits:    0,
		Url:           *ghRepo.HTMLURL,
		Releases:      []Release{},
	}

	releases, err := fetchGithubReleases(*ghRepo.Name, ghOrg.Name, ghTeam)
	if err != nil {
		return nil, err
	}

	// Iterate over the releases in the Github repo
	for _, r := range releases {
		repo.Releases = append(repo.Releases, NewRelease(
			r.GetID(),
			r.GetTagName(),
			r.CreatedAt.Time,
			r.GetName(),
			r.GetBody(),
			r.GetHTMLURL(),
			fmt.Sprintf("%s/compare/%s...%s", repo.Url, r.GetTagName(), repo.DefaultBranch),
		))
	}

	// Add the commit delta between last release and default branch
	comparison, _, err := client.Repositories.CompareCommits(
		ctx,
		ghOrg.Name,
		repo.Name,
		repo.Releases[0].Version,
		repo.DefaultBranch,
		&github.ListOptions{},
	)

	if err != nil {
		return nil, fmt.Errorf(
			"error getting commit comparison for release %s in %s/%s/%s",
			repo.Releases[0].Version, ghOrg.Name, ghTeam, repo.Name,
		)
	}

	repo.NewCommits = *comparison.TotalCommits

	return repo, nil
}

// fetchGithubReleases fetches a list of releases for a given repo, in a given team, in a given org
func fetchGithubReleases(repo string, org string, team string) ([]*github.RepositoryRelease, error) {
	client := getGithubClient()
	ctx := context.Background()
	listOpts := &github.ListOptions{PerPage: 3}

	// Grab the releases for the Github Repo
	releases, _, err := client.Repositories.ListReleases(ctx, org, repo, listOpts)
	if err != nil {
		return nil, fmt.Errorf(
			"error listing releases for repo: %s/%s/%s: %v", org, team, repo, err,
		)
	}

	// If there are no releases, don't include the repo
	if len(releases) == 0 {
		return nil, fmt.Errorf("no releases for repo: %s/%s/%s", org, team, repo)
	}
	return releases, nil
}

// NewLaunchpadRepository fetches and parses release information for a single Launchpad project
func NewLaunchpadRepository(project launchpad.Project, lpGroup string) (*Repository, error) {
	if project.Vcs != "Git" {
		// log.Debugf("launchpad project %s has no git repository", project.Name)
		return nil, nil
	}

	// Create a repository to represent the Launchpad project
	repo := &Repository{
		Name:          project.Name,
		DefaultBranch: "",
		NewCommits:    launchpad.CalculateNewCommits(project.Name),
		Url:           fmt.Sprintf("https://git.launchpad.net/%s", project.Name),
		Releases:      []Release{},
	}

	page, err := launchpad.FetchWebDocument(repo.Url)
	if err != nil {
		return nil, fmt.Errorf("could not fetch launchpad repo page %s/%s: %v", lpGroup, repo.Name, err)
	}

	repo.DefaultBranch = launchpad.GetDefaultBranch(page)
	if repo.DefaultBranch == "" {
		return nil, fmt.Errorf("could not parse default branch for: %s/%s", lpGroup, repo.Name)
	}

	tags := launchpad.GetTags(project.Name, page)
	if len(tags) == 0 {
		return nil, nil
	}

	//Iterate over the tags in the launchpad repo
	for _, t := range tags {
		repo.Releases = append(repo.Releases, NewRelease(
			t.Timestamp.Unix(),
			t.Name,
			*t.Timestamp,
			t.Name,
			"",
			fmt.Sprintf("%s/tag/?h=%s", repo.Url, t.Name),
			fmt.Sprintf("%s/diff/?id=%s&id2=%s", repo.Url, t.Commit, repo.DefaultBranch),
		))
	}

	return repo, nil
}
