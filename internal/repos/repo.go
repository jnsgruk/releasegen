package repos

import (
	"slices"

	"github.com/jnsgruk/releasegen/internal/stores"
)

// RepoDetails represents the serialisable form of a Repository for the Report.
type RepoDetails struct {
	Name       string           `json:"name"`
	NewCommits int              `json:"newCommits"`
	URL        string           `json:"url"`
	Releases   []*Release       `json:"releases"`
	Tags       []*Tag           `json:"tags"`
	Commits    []*Commit        `json:"commits"`
	CiActions  []string         `json:"ciActions"`
	Charm      *stores.Artifact `json:"charm"`
	Snap       *stores.Artifact `json:"snap"`
}

// Repository is an interface that provides common methods for different types of repository.
type Repository interface {
	Process() error
}

// Release refers to either Github Release.
type Release struct {
	ID         int64  `json:"id"`
	Version    string `json:"version"`
	Timestamp  int64  `json:"timestamp"`
	Title      string `json:"title"`
	Body       string `json:"body"`
	URL        string `json:"url"`
	CompareURL string `json:"compareUrl"`
}

// Tag refers to a tag.
type Tag struct {
	Name       string `json:"name"`
	Sha        string `json:"sha"`
	Body       string `json:"body"`
	Timestamp  int64  `json:"timestamp"`
	URL        string `json:"url"`
	CompareURL string `json:"compareUrl"`
}

// Commit represents a Git commit.
type Commit struct {
	Sha       string `json:"sha"`
	Author    string `json:"author"`
	Timestamp int64  `json:"timestamp"`
	Message   string `json:"message"`
	URL       string `json:"url"`
}

// repoInSlice is a helper function to test if a given repo is already in a list of repos.
func RepoInSlice(repositories []RepoDetails, r string) bool {
	index := slices.IndexFunc(repositories, func(repo RepoDetails) bool {
		return repo.Name == r
	})

	return index >= 0
}
