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
func NewGithubRepository(repo *gh.Repository, team string, org string) repositories.Repository {
	// Create a repository to represent the Github Repo
	r := &githubRepository{
		info: repositories.RepositoryInfo{
			Name:          *repo.Name,
			DefaultBranch: *repo.DefaultBranch,
			Url:           *repo.HTMLURL,
		},
		org:  org,
		team: team,
	}
	return r
}

// Info returns a serialisable representation of the repository
func (r *githubRepository) Info() repositories.RepositoryInfo { return r.info }

// Process populates the Repository with details of its releases, and commits
func (r *githubRepository) Process() error {
	log.Printf("processing github repo: %s/%s/%s\n", r.org, r.team, r.info.Name)

	client := githubClient()
	ctx := context.Background()
	opts := &gh.ListOptions{PerPage: 3}

	// Check if the repository is archived
	repoObject, _, err := client.Repositories.Get(ctx, r.org, r.info.Name)
	if err != nil {
		desc := parseApiError(err)
		return fmt.Errorf(
			"error while checking archived status for repo '%s/%s/%s': %s",
			r.org, r.team, r.info.Name, desc,
		)
	}
	//Skip archived repositories
	if r.info.IsArchived = repoObject.GetArchived(); r.info.IsArchived {
		return nil
	}

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

	// Scrape the README for eventual Charm and CI information
	readme, _, err := client.Repositories.GetReadme(ctx, r.org, r.info.Name, nil)
	if err != nil {
		desc := parseApiError(err)
		return fmt.Errorf("error getting README for repo '%s/%s/%s': %s", r.org, r.team, r.info.Name, desc)
	}

	readmeContent, err := readme.GetContent()
	if err != nil {
		return fmt.Errorf("error getting README content for repo '%s/%s/%s'", r.org, r.team, r.info.Name)
	}

	// Extract CI info from the README
	if ciActions := repositories.GetCiActions(readmeContent, r.info.Name); len(ciActions) > 0 {
		r.info.CiActions = ciActions
	}

	// If the README has a Charmhub Badge, fetch the charm information
	if charmName := stores.GetCharmName(readmeContent); charmName != "" {
		charmInfo, err := stores.FetchCharmInfo(charmName)
		if err != nil {
			log.Printf("failed to fetch charm information for charm: %s", charmName)
		} else {
			r.info.Charm = stores.NewStoreArtifact(charmName, charmInfo)
		}
	}

	// If the README has a Snapcraft Badge, fetch the snap information
	if snapName := stores.GetSnapName(readmeContent); snapName != "" {
		snapInfo, err := stores.FetchSnapInfo(snapName)
		if err != nil {
			log.Printf("failed to fetch snap package information for snap: %s", snapName)
		} else {
			r.info.Snap = stores.NewStoreArtifact(snapName, snapInfo)
		}
	}
	return nil
}
