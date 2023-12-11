package releasegen

import (
	"fmt"
	"log"
	"sort"

	"github.com/jnsgruk/releasegen/internal/gitea"
	"github.com/jnsgruk/releasegen/internal/github"
	"github.com/jnsgruk/releasegen/internal/launchpad"
	"github.com/jnsgruk/releasegen/internal/repos"
)

// TeamDetails is the serialisable form of a real-life team.
type TeamDetails struct {
	Name  string              `json:"team"`
	Repos []repos.RepoDetails `json:"repos"`
}

// Team represents a given "real-life Team".
type Team struct {
	Details     *TeamDetails
	config      TeamConfig
	githubToken string
}

// Process populates a given team with the details of its Github/Launchpad/Gitea repos.
func (t *Team) Process() error {
	log.Printf("processing team: %s", t.config.Name)

	// Iterate over the Github orgs for a given team.
	for _, org := range t.config.GithubConfig {
		log.Printf("processing github org: %s\n", org.Org)

		// Set the Github token on the org so it can access the API
		org.SetGithubToken(t.githubToken)

		ghRepos, err := github.FetchOrgRepos(org)
		if err != nil {
			return fmt.Errorf("error populating github repos: %w", err)
		}

		t.Details.Repos = append(t.Details.Repos, ghRepos...)
	}

	// Iterate over the Launchpad Project Groups for the team.
	for _, group := range t.config.LaunchpadConfig.ProjectGroups {
		log.Printf("processing launchpad project group: %s\n", group)

		lpRepos, err := launchpad.FetchProjectGroupRepos(group, t.config.LaunchpadConfig)
		if err != nil {
			return fmt.Errorf("error populating launchpad repos: %w", err)
		}

		t.Details.Repos = append(t.Details.Repos, lpRepos...)
	}

	// Iterate over the Gitea orgs.
	for _, org := range t.config.GiteaConfig {
		log.Printf("processing gitea org: %s\n", org.Org)

		odRepos, err := gitea.FetchOrgRepos(org)
		if err != nil {
			return fmt.Errorf("error populating gitea repos: %w", err)
		}

		t.Details.Repos = append(t.Details.Repos, odRepos...)
	}

	// Sort the repos by the last released.
	sort.Slice(t.Details.Repos, func(i, j int) bool {
		if len(t.Details.Repos[i].Releases) == 0 || len(t.Details.Repos[j].Releases) == 0 {
			return false
		}

		return t.Details.Repos[i].Releases[0].Timestamp > t.Details.Repos[j].Releases[0].Timestamp
	})

	return nil
}
