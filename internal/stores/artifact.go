package stores

import (
	"slices"
	"time"

	"github.com/tidwall/gjson"
)

// Artifact holds information about an artifact in a Canonical store (e.g. a snap or charm).
type Artifact struct {
	Name     string     `json:"name"`
	URL      string     `json:"url"`
	Releases []*Release `json:"releases"`
	Channels []string   `json:"channels"`
	Tracks   []string   `json:"tracks"`
}

// NewArtifact returns a representation of an artifact with its releases/tracks/channels populated.
func NewArtifact(name string, details *ArtifactDetails) *Artifact {
	artifact := &Artifact{
		Name: name,
		URL:  details.StoreURL,
	}

	// Populate the artifact's releases, tracks and channels using the info from the store.
	for index := range details.Tracks {
		parsedTime, _ := time.Parse("2006-01-02T15:04:05.99-07:00", details.ReleaseTimes[index].String())
		track := details.Tracks[index].String()
		channel := details.Channels[index].String()

		base := ""
		if len(details.Bases) > 0 {
			base = details.Bases[index].String()
		}

		artifact.Releases = append(artifact.Releases, &Release{
			Track:     track,
			Channel:   channel,
			Revision:  details.Revisions[index].Int(),
			Timestamp: parsedTime.Unix(),
			Base:      base,
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
	Base      string `json:"base"`
}

// ArtifactDetails is used for storing the raw info fetched about an artifact from the store.
type ArtifactDetails struct {
	StoreURL     string
	Tracks       []gjson.Result
	Channels     []gjson.Result
	ReleaseTimes []gjson.Result
	Revisions    []gjson.Result
	Bases        []gjson.Result
}
