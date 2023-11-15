package github

import (
	"context"
	"fmt"
	"log"
	"slices"

	gh "github.com/google/go-github/v54/github"
	"github.com/jnsgruk/releasegen/internal/config"
	"github.com/jnsgruk/releasegen/internal/repositories"
)

// FetchOrgRepos creates a slice of RepositoryInfo types representing the repos
// owned by the specified teams in the Github org
func FetchOrgRepos(org config.GithubOrg) ([]repositories.RepositoryInfo, error) {
	client := githubClient()
	ctx := context.Background()
	opts := &gh.ListOptions{PerPage: 1000}

	out := []repositories.RepositoryInfo{}

	// Iterate over the Github Teams, listing repos for each
	for _, team := range org.Teams {
		// Lists the Github repositories that the 'ghTeam' has access to.
		orgRepos, _, err := client.Teams.ListTeamReposBySlug(ctx, org.Name, team, opts)
		if err != nil {
			desc := parseApiError(err)
			return nil, fmt.Errorf("error listing repositories for github org '%s': %s", org.Name, desc)
		}

		repos := []*githubRepository{}

		// Iterate over repositories, populating release info for each
		for _, r := range orgRepos {
			// Check if the name of the repository is in the ignore list or private
			if slices.Contains(org.Ignores, *r.Name) || *r.Private {
				continue
			}

			// See if we can find a repo in this team with the same name, if the repository has
			// already been added, skip
			index := slices.IndexFunc(out, func(repo repositories.RepositoryInfo) bool {
				return repo.Name == *r.Name
			})
			if index >= 0 {
				continue
			}

			repo := NewGithubRepository(r, team, org.Name)
			repos = append(repos, repo)

			err := repo.Process()
			if err != nil {
				desc := parseApiError(err)
				log.Printf("error populating repo '%s' from github: %s", repo.Info().Name, desc)
			}
		}

		// Iterate over repos and add the unarchived ones that have at least one commit
		for _, r := range repos {
			if !r.IsArchived() && (len(r.Info().Releases) > 0 || len(r.Info().Commits) > 0) {
				out = append(out, r.Info())
			}
		}
	}
	return out, nil
}
