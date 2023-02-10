package releases

import (
	"time"

	"github.com/jnsgruk/releasegen/internal/md"
)

// RepositoryInfo represents the serialisable form of a Repository for the Report
type RepositoryInfo struct {
	Name          string          `json:"name"`
	DefaultBranch string          `json:"default_branch"`
	NewCommits    int             `json:"new_commits"`
	Url           string          `json:"url"`
	Releases      []*Release      `json:"releases"`
  IsArchived      bool          `json:"is_archived"`
  Ci            []string        `json:"ci"`
  CharmUrl      string          `json:"charm_url"`
  CharmReleases []*CharmRelease `json:"charm_releases"`
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
		Body:       md.RenderReleaseBody(body),
		Url:        url,
		CompareUrl: compareUrl,
	}
}

// CharmRelease represents a charm release on CharmHub
type CharmRelease struct {
  Track     string `json:"track"`
  Channel   string `json:"release"`
  Revision  int64  `json:"revision"`
  Timestamp int64  `json:"timestamp"`
}

// NewCharmRelease is used for constructing a valid CharmRelease
func NewCharmRelease(track string, channel string, revision int64, ts time.Time) *CharmRelease {
  return &CharmRelease {
    Track:     track,
    Channel:   channel,
    Revision:  revision,
    Timestamp: ts.Unix(),
  }
}
