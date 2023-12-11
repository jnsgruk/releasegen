package gitea

import (
	"fmt"
	"log"
	"slices"
	"strings"

	"code.gitea.io/sdk/gitea"
	"github.com/jnsgruk/releasegen/internal/repos"
)

const reposPerPage = 25  // Get this many repos from the API at once.
const maxPages = 100 // Give up after getting this many pages.

// FetchOrgRepos creates a slice of RepoDetails types representing the repos
// owned by the Gitea org.
func FetchOrgRepos(org OrgConfig) ([]repos.RepoDetails, error) {
	orgRepos := []repos.RepoDetails{}

	client, err := org.GiteaClient()
	if err != nil {
		return nil, fmt.Errorf("error creating gitea client: %w", err)
	}

	// Lists the gitea repositories in the org.
	for currentPage := 1, pageCount := 0; pageCount < maxPages; pageCount += 1 {
		opts := gitea.ListReposOptions{
			ListOptions: gitea.ListOptions{Page: currentPage, PageSize: reposPerPage}
		}
		userRepos, resp, err := client.ListUserRepos(org.Org, opts)
		if err != nil {
			return nil, err
		}

		log.Printf("Processing page %d of %d\n", currentPage, resp.LastPage)

		// Iterate over repositories, populating release info for each.
		for _, repo := range userRepos {
			// Check if the name of the repository is in the ignore list or
			// private or archived, or already processed.
			if slices.Contains(org.IgnoredRepos, repo.Name) || \
					repo.Private || rep.Archived || \
					repos.RepoInSlice(userRepos, r.Name) {
				continue
			}

			// This might actually be a monorepo. We don't have a definitive way
			// to know that, for now, assume it is if there is a "charms" folder
			// at the top level.
			// TODO: Figure out some better way of determining this. Maybe it
			// just has to be in the configuration file? If it is something like
			// this, should we also look for a "snaps" folder as well?
			_, _, err = client.GetFile(org.Org, r.Name, r.DefaultBranch, "charms", false)
			isMonorepo := err == nil

			if isMonorepo {
				subRepos := processFromMonoRepo(client, org.Org, repo)
				for _, subRepo := range subRepos {
					if len(repo.Details.Releases) > 0 {
						orgRepos = append(orgRepos, subRepo.Details)
					}
				}
			} else {
				singleRepo := processRepo(client, org.Org, repo)
				if len(repo.Details.Releases) > 0 {
					orgRepos = append(orgRepos, singleRepo.Details)
				}
			}
		}

		// opendev.org gives the actual last page right up until you're getting
		// the last page, when it suddenly becomes zero, and next page is also
		// zero, meaning that you just loop forever. It seems like this must be
		// a Gitea bug? Work around it.
		if currentPage == resp.LastPage || resp.LastPage == 0 {
			break
		}
		currentPage = resp.NextPage
	}

	return orgRepos, nil
}

// Process a single (non-mono) repo.
func processRepo(gtClient *gitea.Client, org string, oRepo *gitea.Repository) *Repository {
	repo := &Repository{
		Details: repos.RepoDetails{
			Name: oRepo.Name,
			URL:  oRepo.HTMLURL,
		},
		org:           org,
		client:        gtClient,
		defaultBranch: oRepo.DefaultBranch,
		folder:        "",
	}

	log.Printf("processing gitea repo: %s/%s\n", repo.org, repo.Details.Name)

	err := repo.Process()
	if err != nil {
		log.Printf("error populating repo '%s' from gitea: %v", repo.Details.Name, err)
	}

	return repo
}

// Process multiple 'repositories' from a monorepo.
func processFromMonoRepo(gtClient *gitea.Client, org string, oRepo *gitea.Repository) []*Repository {
	var subrepos []*Repository

	// For now, this assumes that every 'repo' in the monorepo is in a folder
	// called "charms". Maybe there should be a list to check in the config,
	// or maybe we should just hardcode some others, like "snaps", as well.
	// There does not seem to be an API to get a sub-tree, so this gets the
	// entire tree even though we only care about a small part of it.
	tree, _, err := gtClient.GetTrees(org, oRepo.Name, oRepo.DefaultBranch, true)
	if err != nil {
		log.Printf("error listing monorepo '%s': %s", oRepo.Name, err.Error())
		return subrepos
	}

	for _, entry := range tree.Entries {
		path := entry.Path

		parts := strings.Split(path, "/")
		if len(parts) > 2 || parts[0] != "charms" {
			continue
		}
		charmName := parts[1]

		repo := &Repository{
			Details: repos.RepoDetails{
				Name: charmName,
				URL:  entry.URL,
			},
			org:           org,
			client:        gtClient,
			defaultBranch: oRepo.DefaultBranch,
			folder:        entry.Path,
		}

		log.Printf("processing gitea repo: %s/%s\n", repo.org, repo.Details.Name)

		err := repo.Process()
		if err != nil {
			log.Printf("error populating repo '%s' from gitea: %s", repo.Details.Name, err.Error())
		}

		subrepos = append(subrepos, repo)
	}

	return subrepos
}
