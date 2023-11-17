package stores

import (
	"regexp"
	"slices"
	"time"

	"github.com/tidwall/gjson"
)

// StoreArtifact holds information about an artifact in a Canonical store (e.g. a snap or charm)
type StoreArtifact struct {
	Name     string          `json:"name"`
	Url      string          `json:"url"`
	Releases []*storeRelease `json:"releases"`
	Channels []string        `json:"channels"`
	Tracks   []string        `json:"tracks"`
}

func NewStoreArtifact(name string, r *artifactInfoResult) *StoreArtifact {
	artifact := &StoreArtifact{
		Name: name,
		Url:  r.StoreURL,
	}

	// Populate the artifact's releases, tracks and channels using the info from the store
	for index := range r.Tracks {
		parsedTime, _ := time.Parse("2006-01-02T15:04:05.99-07:00", r.ReleaseTimes[index].String())
		track := r.Tracks[index].String()
		channel := r.Channels[index].String()

		artifact.Releases = append(artifact.Releases, newStoreRelease(
			track,
			channel,
			r.Revisions[index].Int(),
			parsedTime,
		))

		if !slices.Contains(artifact.Tracks, track) {
			artifact.Tracks = append(artifact.Tracks, track)
		}

		if !slices.Contains(artifact.Channels, channel) {
			artifact.Channels = append(artifact.Channels, channel)
		}
	}
	return artifact
}

// storeRelease represents a given release of an artifact in a Canonical Store
type storeRelease struct {
	Track     string `json:"track"`
	Channel   string `json:"channel"`
	Revision  int64  `json:"revision"`
	Timestamp int64  `json:"timestamp"`
}

// newStoreRelease is used for constructing a valid StoreRelease
func newStoreRelease(track string, channel string, revision int64, ts time.Time) *storeRelease {
	return &storeRelease{
		Track:     track,
		Channel:   channel,
		Revision:  revision,
		Timestamp: ts.Unix(),
	}
}

// artifactInfoResult is used for storing the raw info fetched about an artifact from the store
type artifactInfoResult struct {
	StoreURL     string
	Tracks       []gjson.Result
	Channels     []gjson.Result
	ReleaseTimes []gjson.Result
	Revisions    []gjson.Result
}

// getArtifactName tries to parse an artifact name from a store badge in repo's README
func getArtifactName(readme string, re *regexp.Regexp) (name string) {
	nameIndex := re.SubexpIndex("Name")
	matches := re.FindStringSubmatch(readme)
	if len(matches) > 0 {
		name = matches[nameIndex]
	}
	return name
}
