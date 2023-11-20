package github

import (
	"context"
	"fmt"
	"log"
	"slices"

	gh "github.com/google/go-github/v54/github"
	"github.com/jnsgruk/releasegen/internal/repos"
)

// FetchOrgRepos creates a slice of RepoDetails types representing the repos
// owned by the specified teams in the Github org
func FetchOrgRepos(org GithubOrgConfig) (out []repos.RepoDetails, err error) {
	client := githubClient()
	ctx := context.Background()
	opts := &gh.ListOptions{PerPage: 1000}

	// Iterate over the Github Teams, listing repos for each.
	for _, team := range org.Teams {
		// Lists the Github repositories that the 'ghTeam' has access to.
		orgRepos, _, err := client.Teams.ListTeamReposBySlug(ctx, org.Org, team, opts)
		if err != nil {
			return nil, fmt.Errorf("error listing repositories for github org '%s': %s", org.Org, err.Error())
		}

		ghRepos := []*Repository{}

		// Iterate over repositories, populating release info for each.
		for _, r := range orgRepos {
			// Check if the name of the repository is in the ignore list or private.
			if slices.Contains(org.IgnoredRepos, *r.Name) || *r.Private {
				continue
			}

			// See if we can find a repo in this team with the same name, if the repository has
			// already been added, skip.
			index := slices.IndexFunc(out, func(repo repos.RepoDetails) bool {
				return repo.Name == *r.Name
			})
			if index >= 0 {
				continue
			}

			repo := &Repository{
				Details: repos.RepoDetails{
					Name:          *r.Name,
					DefaultBranch: *r.DefaultBranch,
					Url:           *r.HTMLURL,
				},
				org:  org.Org,
				team: team,
			}
			ghRepos = append(ghRepos, repo)

			err := repo.Process()
			if err != nil {
				log.Printf("error populating repo '%s' from github: %s", repo.Details.Name, err.Error())
			}
		}

		// Iterate over ghRepos and add the unarchived ones that have at least one commit
		for _, r := range ghRepos {
			if len(r.Details.Releases) > 0 || len(r.Details.Commits) > 0 {
				out = append(out, r.Details)
			}
		}
	}

	return out, nil
}
