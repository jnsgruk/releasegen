package github

import (
	"context"
	"fmt"
	"log"

	gh "github.com/google/go-github/v54/github"
	"github.com/jnsgruk/releasegen/internal/repositories"
	"github.com/jnsgruk/releasegen/internal/stores"
)

// githubRepository represents a single Github repository
type githubRepository struct {
	info repositories.RepositoryInfo
	org  string // The Github Org that owns the repo
	team string // The Github team, within the org, that has rights over the repo
}

// NewGithubRepository fetches and parses release information for a single Github repository
func NewGithubRepository(repo *gh.Repository, team string, org string) *githubRepository {
	// Create a repository to represent the Github Repo
	return &githubRepository{
		info: repositories.RepositoryInfo{
			Name:          *repo.Name,
			DefaultBranch: *repo.DefaultBranch,
			Url:           *repo.HTMLURL,
		},
		org:  org,
		team: team,
	}
}

// Info returns a serialisable representation of the repository
func (r *githubRepository) Info() repositories.RepositoryInfo { return r.info }

// Process populates the Repository with details of its releases, and commits
func (r *githubRepository) Process() error {
	log.Printf("processing github repo: %s/%s/%s\n", r.org, r.team, r.info.Name)

	// Skip archived repositories
	if r.IsArchived() {
		return nil
	}

	client := githubClient()
	ctx := context.Background()
	opts := &gh.ListOptions{PerPage: 3}

	// Get the releases from the repo
	releases, _, err := client.Repositories.ListReleases(ctx, r.org, r.info.Name, opts)
	if err != nil {
		desc := parseApiError(err)
		return fmt.Errorf(
			"error listing releases for repo '%s/%s/%s': %s", r.org, r.team, r.info.Name, desc,
		)
	} else if len(releases) > 0 {
		// Iterate over the releases in the Github repo
		for _, rel := range releases {
			r.info.Releases = append(r.info.Releases, repositories.NewRelease(
				rel.GetID(),
				rel.GetTagName(),
				rel.CreatedAt.Time,
				rel.GetName(),
				rel.GetBody(),
				rel.GetHTMLURL(),
				fmt.Sprintf("%s/compare/%s...%s", r.info.Url, rel.GetTagName(), r.info.DefaultBranch),
			))
		}

		// Add the commit delta between last release and default branch
		comparison, _, err := client.Repositories.CompareCommits(
			ctx, r.org, r.info.Name, r.info.Releases[0].Version, r.info.DefaultBranch, opts,
		)

		if err != nil {
			desc := parseApiError(err)
			return fmt.Errorf(
				"error getting commit comparison for release '%s' in '%s/%s/%s': %s",
				r.info.Releases[0].Version, r.org, r.team, r.info.Name, desc,
			)
		}

		r.info.NewCommits = *comparison.TotalCommits
	} else {
		// If there are no releases, get the latest commit instead
		commits, _, err := client.Repositories.ListCommits(ctx, r.org, r.info.Name, nil)
		// If there is at least one commit, add it as a release
		if err == nil {
			com := commits[0]
			ts := com.GetCommit().GetAuthor().GetDate()
			r.info.Commits = append(r.info.Commits, repositories.NewCommit(
				com.GetSHA(),
				com.GetCommit().GetAuthor().GetName(),
				*ts.GetTime(),
				com.GetCommit().GetMessage(),
				com.GetHTMLURL(),
			))
		}
	}

	// Get contents of the README as a string
	readmeContent, err := r.readme()
	if err != nil {
		// The rest of this method depends on the README content, so if we don't get
		// any README content, we may as well return early
		return err
	}

	// Parse contents of README to identify associated Github Workflows, snaps, charms
	actions, snap, charm := parseReadmeBadges(r.info.Name, readmeContent)
	r.info.CiActions = actions
	r.info.Snap = snap
	r.info.Charm = charm

	return nil
}

// IsArchived indicates whether or not the repository is marked as archived on Github.
func (r *githubRepository) IsArchived() bool {
	client := githubClient()
	// Check if the repository is archived
	repoObject, _, err := client.Repositories.Get(context.Background(), r.org, r.info.Name)
	if err != nil {
		desc := parseApiError(err)
		log.Printf(
			"error while checking archived status for repo '%s/%s/%s': %s",
			r.org, r.team, r.info.Name, desc,
		)
		return false
	}
	return repoObject.GetArchived()
}

// fetchRepoReadme is a helper function to fetch the README from a Github repository and return
// its contents as a string.
func (r *githubRepository) readme() (string, error) {
	client := githubClient()
	readme, _, err := client.Repositories.GetReadme(context.Background(), r.org, r.info.Name, nil)
	if err != nil {
		desc := parseApiError(err)
		return "", fmt.Errorf("error getting README for repo '%s/%s': %s", r.org, r.info.Name, desc)
	}

	content, err := readme.GetContent()
	if err != nil {
		return "", fmt.Errorf("error getting README content for repo '%s/%s'", r.org, r.info.Name)
	}
	return content, nil
}

// parseReadmeBadges inspects a repos README for badges that indicate association with a given
// snap, charm or CI workflow, and returns those appropriately.
func parseReadmeBadges(repo string, readmeContent string) (ciActions []string, snap *stores.StoreArtifact, charm *stores.StoreArtifact) {
	// Extract CI info from the README
	if actions := repositories.GetCiActions(readmeContent, repo); len(actions) > 0 {
		ciActions = actions
	}

	// If the README has a Charmhub Badge, fetch the charm information
	if charmName := stores.GetCharmName(readmeContent); charmName != "" {
		charmInfo, err := stores.FetchCharmInfo(charmName)
		if err != nil {
			log.Printf("failed to fetch charm information for charm: %s", charmName)
		} else {
			charm = stores.NewStoreArtifact(charmName, charmInfo)
		}
	}

	// If the README has a Snapcraft Badge, fetch the snap information
	if snapName := stores.GetSnapName(readmeContent); snapName != "" {
		snapInfo, err := stores.FetchSnapInfo(snapName)
		if err != nil {
			log.Printf("failed to fetch snap package information for snap: %s", snapName)
		} else {
			snap = stores.NewStoreArtifact(snapName, snapInfo)
		}
	}
	return ciActions, snap, charm
}
