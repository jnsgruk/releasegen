package releases

import (
	"context"
	"fmt"
	"log"
	"sort"
	"sync"

	"github.com/google/go-github/v47/github"
	"github.com/jnsgruk/releasegen/internal/config"
	"github.com/jnsgruk/releasegen/internal/launchpad"
	"golang.org/x/exp/slices"
)

// Team represents a given "real-life Team"
type Team struct {
	Name  string       `json:"team"`
	Repos []Repository `json:"repos"`
}

// NewTeam creates a new team, and populates it's fields using the Github/Launchpad APIs
func NewTeam(configTeam config.Team) *Team {
	log.Printf("processing team: %s", configTeam.Name)
	// Create a template team
	team := &Team{Name: configTeam.Name, Repos: []Repository{}}

	// Iterate over the Github orgs for a given team
	for _, ghOrg := range configTeam.Github {
		log.Printf("processing org: %s\n", ghOrg.Name)

		err := team.populateGithubRepos(ghOrg)
		if err != nil {
			log.Printf("%v", err)
		}
	}

	// Iterate over the Launchpad Project Groups for the team
	for _, lpGroup := range configTeam.Launchpad.ProjectGroups {
		err := team.populateLaunchpadRepos(lpGroup, configTeam.Launchpad.Ignores)
		if err != nil {
			log.Printf("error populating launchpad repos: %v", err)
		}
	}

	// Sort the repos by the last released
	sort.Slice(team.Repos, func(i, j int) bool {
		return team.Repos[i].Releases[0].Timestamp > team.Repos[j].Releases[0].Timestamp
	})
	return team
}

// populateLaunchpadRepos creates a slice of Repo types representing the repos owned by
// the specified Launchpad project groups
func (t *Team) populateLaunchpadRepos(lpGroup string, ignores []string) error {
	pg := launchpad.ProjectGroup{Name: lpGroup}

	projects, err := pg.Projects()
	if err != nil {
		return err
	}
	var wg sync.WaitGroup

	for _, project := range projects {
		wg.Add(1)
		go func(p launchpad.Project) {
			defer wg.Done()
			log.Printf("processing launchpad repo: %s/%s\n", lpGroup, p.Name)

			// Check if the name of the repository is in the ignore list for the team
			if contains(ignores, p.Name) {
				return
			}

			// See if we can find a repo in this team with the same name
			index := slices.IndexFunc(
				t.Repos, func(repo Repository) bool { return repo.Name == p.Name },
			)

			// TODO: Investigate if this read/append is thread safe
			// We couldn't find the repo, so go ahead and add it
			if index < 0 {
				repo, err := NewLaunchpadRepository(p, lpGroup)
				if repo != nil {
					t.Repos = append(t.Repos, *repo)
				}
				if err != nil {
					log.Printf("error creating launchpad repository: %v", err)
				}
			}
		}(project)
	}
	wg.Wait()

	return nil
}

// populateGithubRepos creates a slice of Repo types representing the repos owned by the specified
// teams in the Github org
func (t *Team) populateGithubRepos(ghOrg config.GithubOrg) error {
	client := githubClient()
	// repos := []Repo{}
	ctx := context.Background()
	// Iterate over the Github Teams, listing repos for each
	for _, ghTeam := range ghOrg.Teams {
		// Lists the Github repositories that the 'ghTeam' has access to.
		ghRepos, _, err := client.Teams.ListTeamReposBySlug(
			ctx,
			ghOrg.Name,
			ghTeam,
			&github.ListOptions{PerPage: 1000},
		)

		if err != nil {
			return fmt.Errorf("error listing repositories for github org: %s", ghOrg.Name)
		}

		var wg sync.WaitGroup

		// Iterate over repositories, populating release info for each
		for _, r := range ghRepos {
			wg.Add(1)
			go func(r *github.Repository) {
				defer wg.Done()
				log.Printf("processing github repo: %s/%s/%s\n", ghOrg.Name, ghTeam, *r.Name)

				// Check if the name of the repository is in the ignore list for the team
				if contains(ghOrg.Ignores, *r.Name) {
					return
				}

				// See if we can find a repo in this team with the same name
				index := slices.IndexFunc(
					t.Repos, func(repo Repository) bool { return repo.Name == *r.Name },
				)

				// TODO: Investigate if this read/append is thread safe
				// We couldn't find the repo, so go ahead and add it
				if index < 0 {
					repo, err := NewGithubRepository(r, ghTeam, ghOrg)
					if err == nil {
						t.Repos = append(t.Repos, *repo)
					}
				}
			}(r)
		}
		wg.Wait()
	}

	return nil
}

// contains checks if a string is present in a slice
func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
}
