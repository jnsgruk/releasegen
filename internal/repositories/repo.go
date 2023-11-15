package repositories

import (
	"time"

	"github.com/jnsgruk/releasegen/internal/stores"
)

// RepositoryInfo represents the serialisable form of a Repository for the Report
type RepositoryInfo struct {
	Name          string                `json:"name"`
	DefaultBranch string                `json:"default_branch"`
	NewCommits    int                   `json:"new_commits"`
	Url           string                `json:"url"`
	Releases      []*Release            `json:"releases"`
	Commits       []*Commit             `json:"commits"`
	IsArchived    bool                  `json:"is_archived"`
	CiActions     []string              `json:"ci_actions"`
	Charm         *stores.StoreArtifact `json:"charm"`
	Snap          *stores.StoreArtifact `json:"snap"`
}

// Repository is an interface that provides common methods for different types of repository
type Repository interface {
	Info() RepositoryInfo
	Process() error
}

// Release refers to either Github Release, or a Launchpad Tag
type Release struct {
	Id         int64  `json:"id"`
	Version    string `json:"version"`
	Timestamp  int64  `json:"timestamp"`
	Title      string `json:"title"`
	Body       string `json:"body"`
	Url        string `json:"url"`
	CompareUrl string `json:"compare_url"`
}

// NewRelease is used for constructing a valid Release
func NewRelease(id int64, version string, ts time.Time, title string, body string, url string, compareUrl string) *Release {
	return &Release{
		Id:         id,
		Version:    version,
		Timestamp:  ts.Unix(),
		Title:      title,
		Body:       RenderReleaseBody(body),
		Url:        url,
		CompareUrl: compareUrl,
	}
}

// Commit represents a GitHub commit
type Commit struct {
	Sha       string `json:"sha"`
	Author    string `json:"author"`
	Timestamp int64  `json:"timestamp"`
	Message   string `json:"message"`
	Url       string `json:"url"`
}

// NewCommit constructs a valid Commit
func NewCommit(sha string, author string, ts time.Time, message string, url string) *Commit {
	return &Commit{
		Sha:       sha,
		Author:    author,
		Timestamp: ts.Unix(),
		Message:   message,
		Url:       url,
	}
}
