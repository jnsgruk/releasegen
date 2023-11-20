package stores

import (
	"slices"
	"time"

	"github.com/tidwall/gjson"
)

// Artifact holds information about an artifact in a Canonical store (e.g. a snap or charm).
type Artifact struct {
	Name     string     `json:"name"`
	Url      string     `json:"url"`
	Releases []*Release `json:"releases"`
	Channels []string   `json:"channels"`
	Tracks   []string   `json:"tracks"`
}

// NewArtifact returns a representation of an artifact with its releases/tracks/channels populated.
func NewArtifact(name string, r *artifactDetails) *Artifact {
	artifact := &Artifact{
		Name: name,
		Url:  r.StoreURL,
	}

	// Populate the artifact's releases, tracks and channels using the info from the store.
	for index := range r.Tracks {
		parsedTime, _ := time.Parse("2006-01-02T15:04:05.99-07:00", r.ReleaseTimes[index].String())
		track := r.Tracks[index].String()
		channel := r.Channels[index].String()

		artifact.Releases = append(artifact.Releases, &Release{
			Track:     track,
			Channel:   channel,
			Revision:  r.Revisions[index].Int(),
			Timestamp: parsedTime.Unix(),
		})

		if !slices.Contains(artifact.Tracks, track) {
			artifact.Tracks = append(artifact.Tracks, track)
		}

		if !slices.Contains(artifact.Channels, channel) {
			artifact.Channels = append(artifact.Channels, channel)
		}
	}

	return artifact
}

// Release represents a given Release of an artifact in a Canonical Store.
type Release struct {
	Track     string `json:"track"`
	Channel   string `json:"channel"`
	Revision  int64  `json:"revision"`
	Timestamp int64  `json:"timestamp"`
}

// ArtifactDetails is used for storing the raw info fetched about an artifact from the store.
type artifactDetails struct {
	StoreURL     string
	Tracks       []gjson.Result
	Channels     []gjson.Result
	ReleaseTimes []gjson.Result
	Revisions    []gjson.Result
}
