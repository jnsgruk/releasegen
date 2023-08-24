package releases

import (
	"context"
	"fmt"
	"log"
	"sort"
	"sync"

	"github.com/google/go-github/v54/github"
	"github.com/jnsgruk/releasegen/internal/config"
	"github.com/jnsgruk/releasegen/internal/launchpad"
	"golang.org/x/exp/slices"
)

// TeamInfo is the serialisable form of a real-life team
type TeamInfo struct {
	Name  string           `json:"team"`
	Repos []RepositoryInfo `json:"repos"`
}

// Team represents a given "real-life Team"
type Team struct {
	info   TeamInfo
	config config.Team
}

// NewTeam creates a new team, and populates it's fields using the Github/Launchpad APIs
func NewTeam(configTeam config.Team) *Team {
	team := &Team{
		info:   TeamInfo{Name: configTeam.Name},
		config: configTeam,
	}
	return team
}

// Info returns a serialisable form of a given team
func (t *Team) Info() *TeamInfo { return &t.info }

// Process is used to populate a given team with the details of its Github Org and Launchpad
// repositories
func (t *Team) Process() error {
	log.Printf("processing team: %s", t.config.Name)

	// Iterate over the Github orgs for a given team
	for _, org := range t.config.Github {
		log.Printf("processing github org: %s\n", org.Name)
		err := t.populateGithubRepos(org)
		if err != nil {
			return fmt.Errorf("error populating github repos: %w", err)
		}
	}

	// Iterate over the Launchpad Project Groups for the team
	for _, group := range t.config.Launchpad.ProjectGroups {
		log.Printf("processing launchpad project group: %s\n", group)
		err := t.populateLaunchpadRepos(group)
		if err != nil {
			return fmt.Errorf("error populating launchpad repos: %w", err)
		}
	}

	// Sort the repos by the last released
	sort.Slice(t.info.Repos, func(i, j int) bool {
		if len(t.info.Repos[i].Releases) == 0 || len(t.info.Repos[j].Releases) == 0 {
			return false
		}
		return t.info.Repos[i].Releases[0].Timestamp > t.info.Repos[j].Releases[0].Timestamp
	})

	return nil
}

// populateLaunchpadRepos creates a slice of Repo types representing the repos owned by
// the specified Launchpad project groups
func (t *Team) populateLaunchpadRepos(projectGroup string) error {
	projects, err := launchpad.EnumerateProjectGroup(projectGroup)
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	repos := []Repository{}

	for _, project := range projects {
		// Check if the name of the repository is in the ignore list for the team, or if it
		// specifies any VCS type other than Git
		if contains(t.config.Launchpad.Ignores, project.Name) || project.Vcs != "Git" {
			continue
		}

		// See if we can find a repo in this team with the same name, if the repository has
		// already been added, skip
		index := slices.IndexFunc(t.info.Repos, func(repo RepositoryInfo) bool {
			return repo.Name == project.Name
		})
		if index >= 0 {
			continue
		}

		repo := NewLaunchpadRepository(project, projectGroup)
		repos = append(repos, repo)

		wg.Add(1)
		go func() {
			defer wg.Done()
			err := repo.Process()
			if err != nil {
				log.Printf("error populating repo %s from launchpad: %v", repo.Info().Name, err)
			}
		}()
	}
	wg.Wait()

	// Iterate over repos, add only those that have releases to the Team's list of repos
	for _, r := range repos {
		if len(r.Info().Releases) > 0 {
			t.info.Repos = append(t.info.Repos, r.Info())
		}
	}
	return nil
}

// populateGithubRepos creates a slice of Repo types representing the repos owned by the specified
// teams in the Github org
func (t *Team) populateGithubRepos(org config.GithubOrg) error {
	client := githubClient()
	ctx := context.Background()
	opts := &github.ListOptions{PerPage: 1000}

	// Iterate over the Github Teams, listing repos for each
	for _, team := range org.Teams {
		// Lists the Github repositories that the 'ghTeam' has access to.
		orgRepos, _, err := client.Teams.ListTeamReposBySlug(ctx, org.Name, team, opts)
		if err != nil {
			desc := parseGithubApiError(err)
			return fmt.Errorf("error listing repositories for github org '%s': %s", org.Name, desc)
		}

		var wg sync.WaitGroup
		repos := []Repository{}

		// Iterate over repositories, populating release info for each
		for _, r := range orgRepos {
			// Check if the name of the repository is in the ignore list or private
			if contains(org.Ignores, *r.Name) || *r.Private {
				continue
			}

			// See if we can find a repo in this team with the same name, if the repository has
			// already been added, skip
			index := slices.IndexFunc(t.info.Repos, func(repo RepositoryInfo) bool {
				return repo.Name == *r.Name
			})
			if index >= 0 {
				continue
			}

			repo := NewGithubRepository(r, team, org.Name)
			repos = append(repos, repo)

			wg.Add(1)
			go func() {
				defer wg.Done()
				err := repo.Process()
				if err != nil {
					desc := parseGithubApiError(err)
					log.Printf("error populating repo '%s' from github: %s", repo.Info().Name, desc)
				}
			}()
		}
		wg.Wait()

		// Iterate over repos and add the unarchived ones that have at least one commit
		for _, r := range repos {
			if !r.Info().IsArchived && (len(r.Info().Releases) > 0 || len(r.Info().Commits) > 0) {
				t.info.Repos = append(t.info.Repos, r.Info())
			}
		}
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
