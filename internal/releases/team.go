package releases

import (
	"fmt"
	"log"
	"sort"

	"github.com/jnsgruk/releasegen/internal/config"
	"github.com/jnsgruk/releasegen/internal/github"
	"github.com/jnsgruk/releasegen/internal/launchpad"
	"github.com/jnsgruk/releasegen/internal/repositories"
)

// TeamInfo is the serialisable form of a real-life team
type TeamInfo struct {
	Name  string                        `json:"team"`
	Repos []repositories.RepositoryInfo `json:"repos"`
}

// Team represents a given "real-life Team"
type Team struct {
	info   TeamInfo
	config config.Team
}

// NewTeam creates a new team, and populates it's fields using the Github/Launchpad APIs
func NewTeam(configTeam config.Team) *Team {
	team := &Team{
		info: TeamInfo{
			Name:  configTeam.Name,
			Repos: []repositories.RepositoryInfo{},
		},
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
		ghRepos, err := github.FetchOrgRepos(org)
		if err != nil {
			return fmt.Errorf("error populating github repos: %w", err)
		}
		t.info.Repos = append(t.info.Repos, ghRepos...)
	}

	// Iterate over the Launchpad Project Groups for the team
	for _, group := range t.config.Launchpad.ProjectGroups {
		log.Printf("processing launchpad project group: %s\n", group)
		lpRepos, err := launchpad.FetchProjectGroupRepos(group, t.config.Launchpad)
		if err != nil {
			return fmt.Errorf("error populating launchpad repos: %w", err)
		}
		t.info.Repos = append(t.info.Repos, lpRepos...)
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
