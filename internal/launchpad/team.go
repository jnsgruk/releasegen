package launchpad

import (
	"context"
	"fmt"
	"log"
	"slices"
	"sync"

	"github.com/jnsgruk/releasegen/internal/repos"
)

// FetchProjectGroupRepos creates a slice of RepoDetails types representing the repos
// associated with a given ProjectGroup in Launchpad.
func FetchProjectGroupRepos(projectGroup string, config Config) ([]repos.RepoDetails, error) {
	pgRepos := []repos.RepoDetails{}

	lpRepos, err := getProjectGroupRepos(projectGroup, config, &pgRepos)
	if err != nil {
		return nil, err
	}

	// Iterate over repos, add only those that have tags to the Team's list of repos
	for _, r := range lpRepos {
		if len(r.Details.Tags) > 0 {
			pgRepos = append(pgRepos, r.Details)
		}
	}

	return pgRepos, nil
}

// getProjectGroupRepos fetches information about repositories from Launchpad and returns it as
// a slice of structs in the right format for releasegen.
func getProjectGroupRepos(pg string, config Config, pgRepos *[]repos.RepoDetails) ([]*Repository, error) {
	ctx := context.Background()
	lpRepos := []*Repository{}

	projects, err := enumerateProjectGroup(ctx, pg)
	if err != nil {
		return nil, fmt.Errorf("error enumerating project group '%s': %w", pg, err)
	}

	var wg sync.WaitGroup

	for _, project := range projects {
		p := project
		// Check if the name of the repository is in the ignore list for the team
		if slices.Contains(config.IgnoredRepos, p) || repos.RepoInSlice(*pgRepos, p) {
			continue
		}

		// Create a new repository, add to the list of repos for the project group
		repo := &Repository{
			Details: repos.RepoDetails{
				Name: p,
				URL:  fmt.Sprintf("https://git.launchpad.net/%s", p),
			},
			projectGroup: pg,
		}
		lpRepos = append(lpRepos, repo)

		wg.Add(1)

		go func() {
			defer wg.Done()
			log.Printf("processing launchpad repo: %s/%s\n", repo.projectGroup, repo.Details.Name)

			err := repo.Process(ctx)
			if err != nil {
				log.Printf("error populating repo %s from launchpad: %s", repo.Details.Name, err.Error())
			}
		}()
	}

	wg.Wait()

	return lpRepos, nil
}
