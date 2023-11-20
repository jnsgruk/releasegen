package github

import (
	"context"
	"fmt"
	"log"
	"slices"

	gh "github.com/google/go-github/v54/github"
	"github.com/jnsgruk/releasegen/internal/repos"
)

const githubPerPage = 1000

// FetchOrgRepos creates a slice of RepoDetails types representing the repos
// owned by the specified teams in the Github org.
func FetchOrgRepos(org OrgConfig) ([]repos.RepoDetails, error) {
	orgRepos := []repos.RepoDetails{}

	// Iterate over the Github Teams, listing repos for each.
	for _, team := range org.Teams {
		// Process the Github orgs repositories into structs with the info we need.
		teamRepos, err := getTeamRepos(org, team, orgRepos)
		if err != nil {
			return nil, err
		}

		// Iterate over ghRepos and add the unarchived ones that have at least one commit.
		for _, r := range teamRepos {
			if len(r.Details.Releases) > 0 || len(r.Details.Commits) > 0 {
				orgRepos = append(orgRepos, r.Details)
			}
		}
	}

	return orgRepos, nil
}

// getTeamRepos fetches a team's repos from the Github client and processes them into
// the format releasegen needs.
func getTeamRepos(org OrgConfig, team string, orgRepos []repos.RepoDetails) ([]*Repository, error) {
	ghRepos := []*Repository{}
	ctx := context.Background()
	opts := &gh.ListOptions{PerPage: githubPerPage}

	// Lists the Github repositories that the 'ghTeam' has access to.
	teamRepos, _, err := org.GithubClient().Teams.ListTeamReposBySlug(ctx, org.Org, team, opts)
	if err != nil {
		return nil, fmt.Errorf("error listing repositories for github org: %s", org.Org)
	}

	// Iterate over repositories, populating release info for each.
	for _, tRepo := range teamRepos {
		r := tRepo
		// Check if the name of the repository is in the ignore list or private, or already processed.
		if slices.Contains(org.IgnoredRepos, *r.Name) || *r.Private || repoInSlice(orgRepos, *r.Name) {
			continue
		}

		repo := &Repository{
			Details: repos.RepoDetails{
				Name:          *r.Name,
				DefaultBranch: *r.DefaultBranch,
				URL:           *r.HTMLURL,
			},
			org:    org.Org,
			team:   team,
			client: org.GithubClient(),
		}
		ghRepos = append(ghRepos, repo)

		log.Printf("processing github repo: %s/%s/%s\n", repo.org, repo.team, repo.Details.Name)

		err := repo.Process(ctx)
		if err != nil {
			log.Printf("error populating repo '%s' from github: %s", repo.Details.Name, err.Error())
		}
	}

	return ghRepos, nil
}

// repoInSlice is a helper function to test if a given repo is already in a list of repos.
func repoInSlice(repositories []repos.RepoDetails, r string) bool {
	index := slices.IndexFunc(repositories, func(repo repos.RepoDetails) bool {
		return repo.Name == r
	})

	return index >= 0
}
