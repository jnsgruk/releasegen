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
func FetchProjectGroupRepos(projectGroup string, config LaunchpadConfig) (out []repos.RepoDetails, err error) {
	projects, err := enumerateProjectGroup(projectGroup)
	if err != nil {
		return nil, err
	}

	var wg sync.WaitGroup

	lpRepos := []*Repository{}

	for _, project := range projects {
		// Check if the name of the repository is in the ignore list for the team
		if slices.Contains(config.IgnoredRepos, project) {
			continue
		}

		// See if we can find a repo in this team with the same name, if yes then skip
		index := slices.IndexFunc(out, func(repo repos.RepoDetails) bool {
			return repo.Name == project
		})
		if index >= 0 {
			continue
		}

		// Create a new repository, add to the list of repos for the project group
		repo := &Repository{
			Details: repos.RepoDetails{
				Name:          project,
				DefaultBranch: "",
				Url:           fmt.Sprintf("https://git.launchpad.net/%s", project),
			},
			projectGroup: projectGroup,
		}
		lpRepos = append(lpRepos, repo)

		wg.Add(1)

		go func() {
			defer wg.Done()

			err := repo.Process()
			if err != nil {
				log.Printf("error populating repo %s from launchpad: %v", repo.Details.Name, err)
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
