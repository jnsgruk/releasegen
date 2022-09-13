package releases

import (
	"time"

	"github.com/jnsgruk/releasegen/internal/md"
)

// Release refers to either Github Release, or a Launchpad Tag
type Release struct {
	Id         int64  `json:"id"`
	Version    string `json:"version"`
	Timestamp  int64  `json:"timestamp"`
	Title      string `json:"title"`
	Body       string `json:"body_html"`
	Url        string `json:"href"`
	CompareUrl string `json:"compare_href"`
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
