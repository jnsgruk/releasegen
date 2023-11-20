package launchpad

import (
	"fmt"
	"log"
	"slices"
	"sync"

	"github.com/jnsgruk/releasegen/internal/repos"
)

// FetchProjectGroupRepos creates a slice of RepoDetails types representing the repos
// associated with a given ProjectGroup in Launchpad.
func FetchProjectGroupRepos(projectGroup string, config Config) ([]repos.RepoDetails, error) {
	projects, err := enumerateProjectGroup(projectGroup)
	if err != nil {
		return nil, fmt.Errorf("error enumerating project group '%s': %w", projectGroup, err)
	}

	var wg sync.WaitGroup

	lpRepos := []*Repository{}
	out := []repos.RepoDetails{}

	for _, project := range projects {
		p := project
		// Check if the name of the repository is in the ignore list for the team
		if slices.Contains(config.IgnoredRepos, p) {
			continue
		}

		// See if we can find a repo in this team with the same name, if yes then skip
		index := slices.IndexFunc(out, func(repo repos.RepoDetails) bool {
			return repo.Name == p
		})
		if index >= 0 {
			continue
		}

		// Create a new repository, add to the list of repos for the project group
		repo := &Repository{
			Details: repos.RepoDetails{
				Name:          p,
				DefaultBranch: "",
				URL:           fmt.Sprintf("https://git.launchpad.net/%s", p),
			},
			projectGroup: projectGroup,
		}
		lpRepos = append(lpRepos, repo)

		wg.Add(1)

		go func() {
			defer wg.Done()

			err := repo.Process()
			if err != nil {
				log.Printf("error populating repo %s from launchpad: %s", repo.Details.Name, err.Error())
			}
		}()
	}

	wg.Wait()

	// Iterate over repos, add only those that have releases to the Team's list of repos
	for _, r := range lpRepos {
		if len(r.Details.Releases) > 0 {
			out = append(out, r.Details)
		}
	}

	return out, nil
}
