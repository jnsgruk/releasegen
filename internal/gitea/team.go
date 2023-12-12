package gitea

import (
	"fmt"
	"log"
	"slices"
	"strings"

	"code.gitea.io/sdk/gitea"
	"github.com/jnsgruk/releasegen/internal/repos"
)

const reposPerPage = 25 // Get this many repos from the API at once.
const maxPages = 100    // Give up after getting this many pages.

type repoMeta struct {
	repo            *gitea.Repository
	monorepoSources []string
}

// publicOrgRepos finds repositories that are public, not archived, and belong
// to the specified organisation.
func publicOrgRepos(org OrgConfig, client *gitea.Client) ([]repoMeta, error) {
	var orgRepos []repoMeta

	currentPage := 1
	for pageCount := 0; pageCount < maxPages; pageCount += 1 {
		opts := gitea.ListReposOptions{
			ListOptions: gitea.ListOptions{Page: currentPage, PageSize: reposPerPage},
		}
		userRepos, resp, err := client.ListUserRepos(org.Org, opts)
		if err != nil {
			return nil, err
		}

		log.Printf("Processing page %d of %d\n", currentPage, resp.LastPage)

		// Iterate over repositories, populating release info for each.
		for _, repo := range userRepos {
			// Check if the name of the repository is in the ignore list or is
			// private or archived.
			if slices.Contains(org.IgnoredRepos, repo.Name) || repo.Private || repo.Archived {
				continue
			}
			meta := repoMeta{
				repo: repo,
			}
			orgRepos = append(orgRepos, meta)
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

// specifiedRepos gets metadata about the repositories in the the list specified
// in the `include` list in the configuration.
func specifiedRepos(org OrgConfig, client *gitea.Client) ([]repoMeta, error) {
	var orgRepos []repoMeta

	for name, conf := range org.IncludeRepos {
		repo, _, err := client.GetRepo(org.Org, name)
		if err != nil {
			return nil, err
		}

		meta := repoMeta{
			repo:            repo,
			monorepoSources: conf.MonorepoSources,
		}

		orgRepos = append(orgRepos, meta)
	}

	return orgRepos, nil
}

// FetchOrgRepos creates a slice of RepoDetails types representing the repos
// owned by the Gitea org.
func FetchOrgRepos(org OrgConfig) ([]repos.RepoDetails, error) {
	var orgRepos []repos.RepoDetails

	log.Printf("Connecting to gitea at %s\n", org.URL)
	client, err := gitea.NewClient(org.URL)
	if err != nil {
		return nil, fmt.Errorf("error creating gitea client: %w", err)
	}

	var repoList []repoMeta
	if len(org.IncludeRepos) > 0 {
		repoList, err = specifiedRepos(org, client)
	} else {
		repoList, err = publicOrgRepos(org, client)
	}
	if err != nil {
		return nil, err
	}

	// Iterate over repositories, populating release info for each.
	for _, meta := range repoList {
		if len(meta.monorepoSources) > 0 {
			collectedDetails := processFromMonoRepo(
				client,
				org.Org,
				meta.repo,
				meta.monorepoSources,
			)
			for _, details := range collectedDetails {
				if len(details.Releases) > 0 || len(details.Commits) > 0 {
					orgRepos = append(orgRepos, details)
				}
			}
		} else {
			details := processRepo(client, org.Org, meta.repo)
			if len(details.Releases) > 0 || len(details.Commits) > 0 {
				orgRepos = append(orgRepos, details)
			}
		}
	}

	return orgRepos, nil
}

// Process a single (non-mono) repo.
func processRepo(client *gitea.Client, org string, orgRepo *gitea.Repository) repos.RepoDetails {
	repo := &Repository{
		Details: repos.RepoDetails{
			Name: orgRepo.Name,
			URL:  orgRepo.HTMLURL,
		},
		org:           org,
		client:        client,
		defaultBranch: orgRepo.DefaultBranch,
		folder:        "",
	}

	log.Printf("processing gitea repo: %s/%s\n", org, repo.Details.Name)

	err := repo.Process()
	if err != nil {
		log.Printf("error populating repo '%s' from gitea: %s", repo.Details.Name, err.Error())
	}

	return repo.Details
}

// Process multiple 'repositories' from a monorepo.
func processFromMonoRepo(
	client *gitea.Client,
	org string,
	orgRepo *gitea.Repository,
	folders []string,
) []repos.RepoDetails {
	var collectedDetails []repos.RepoDetails

	// It seems like the gitea client does not have the ability to get only the
	// top level of the tree, so we have to get the entire tree for the repo.
	tree, _, err := client.GetTrees(org, orgRepo.Name, orgRepo.DefaultBranch, true)
	if err != nil {
		log.Printf("error listing monorepo '%s': %s", orgRepo.Name, err.Error())
		return collectedDetails
	}

	for _, folder := range folders {
		for _, entry := range tree.Entries {
			parts := strings.Split(entry.Path, "/")
			if len(parts) < 2 || parts[0] != folder {
				continue
			}
			name := parts[1]
			if repos.RepoInSlice(collectedDetails, name) {
				continue
			}

			repo := &Repository{
				Details: repos.RepoDetails{
					Name:     name,
					URL:      entry.URL,
					Monorepo: orgRepo.Name,
				},
				org:           org,
				client:        client,
				defaultBranch: orgRepo.DefaultBranch,
				folder:        entry.Path,
			}

			log.Printf("processing gitea sub-repo: %s %s/%s/%s\n",
				org, orgRepo.Name, folder, repo.Details.Name)

			err := repo.Process()
			if err != nil {
				log.Printf(
					"error populating sub-repo '%s/%s' from gitea: %s",
					repo.Details.Name, folder, err.Error())
			}

			collectedDetails = append(collectedDetails, repo.Details)
		}
	}

	return collectedDetails
}
