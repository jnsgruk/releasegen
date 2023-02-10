package releases

import (
	"context"
	"fmt"
	"log"

	"github.com/google/go-github/v47/github"
)

// githubRepository represents a single Github repository
type githubRepository struct {
	info RepositoryInfo
	org  string // The Github Org that owns the repo
	team string // The Github team, within the org, that has rights over the repo
}

// NewGithubRepository fetches and parses release information for a single Github repository
func NewGithubRepository(repo *github.Repository, team string, org string) Repository {
	// Create a repository to represent the Github Repo
	r := &githubRepository{
		info: RepositoryInfo{
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
func (r *githubRepository) Info() RepositoryInfo { return r.info }

// Process populates the Repository with details of its releases, and commits
func (r *githubRepository) Process() error {
	log.Printf("processing github repo: %s/%s/%s\n", r.org, r.team, r.info.Name)

	client := githubClient()
	ctx := context.Background()
	opts := &github.ListOptions{PerPage: 3}

  // Check if the repository is archived
  repoObject, _, err := client.Repositories.Get(ctx, r.org, r.info.Name)
  if err != nil {
    return fmt.Errorf(
      "error while checking archived status for repo: %s/%s/%s: %v", r.org, r.team, r.info.Name, err,
    )
  }
  r.info.IsArchived = repoObject.GetArchived()

  // Get the releases from the repo
	releases, _, err := client.Repositories.ListReleases(ctx, r.org, r.info.Name, opts)
	if err != nil {
		return fmt.Errorf(
			"error listing releases for repo: %s/%s/%s: %v", r.org, r.team, r.info.Name, err,
		)
	} else if len(releases) == 0 {
		return nil
	}

	// Iterate over the releases in the Github repo
	for _, rel := range releases {
		r.info.Releases = append(r.info.Releases, NewRelease(
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
		return fmt.Errorf(
			"error getting commit comparison for release %s in %s/%s/%s",
			r.info.Releases[0].Version, r.org, r.team, r.info.Name,
		)
	}

	r.info.NewCommits = *comparison.TotalCommits


  // Scrape the README for eventual Charm and CI information
  readme, _ , err := client.Repositories.GetReadme(ctx, r.org, r.info.Name, nil)
  if err != nil {
    return fmt.Errorf(
      "error getting README for repo: %s/%s/%s: %v", r.org, r.team, r.info.Name, err,
    )
  }

  readmeContent, err := readme.GetContent()
  if err != nil {
    return fmt.Errorf(
      "error reading README for repo: %s/%s/%s: %v", r.org, r.team, r.info.Name, err,
    )
  }

  // Extract CI and Charm info from the README
  ciStages := GetCiStages(readmeContent, r.info.Name)
  if len(ciStages) > 0 {
    r.info.Ci = ciStages
  }
  charmName := GetCharmName(readmeContent)
  if charmName != "" {
    r.info.CharmUrl = fmt.Sprintf("https://charmhub.io/%s", charmName)
    r.info.CharmReleases, err = FetchCharmInfo(charmName)
  }

	return nil
}
