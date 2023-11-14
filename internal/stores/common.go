package stores

import (
	"regexp"
	"time"
)

// charmBadgeRegexp is used to find a Charm's name in its CharmHub badge
var charmBadgeRegexp = regexp.MustCompile(`!\[.*\]\(https://charmhub.io/(?P<Name>.+)/badge.svg\)`)

// snapBadgeRegExp is used to find a Snap's name in its Snapcraft badge
var snapBadgeRegExp = regexp.MustCompile(`!\[.*\]\(https://snapcraft.io/(?P<Name>.+)/badge.svg\)`)

// StoreArtifact holds information about an artifact in a Canonical store (e.g. a snap or charm)
type StoreArtifact struct {
	Name     string          `json:"name"`
	Url      string          `json:"url"`
	Releases []*StoreRelease `json:"releases"`
	Channels []string        `json:"channels"`
	Tracks   []string        `json:"tracks"`
}

// StoreRelease represents a given release of an artifact in a Canonical Store
type StoreRelease struct {
	Track     string `json:"track"`
	Channel   string `json:"channel"`
	Revision  int64  `json:"revision"`
	Timestamp int64  `json:"timestamp"`
}

// NewStoreRelease is used for constructing a valid StoreRelease
func NewStoreRelease(track string, channel string, revision int64, ts time.Time) *StoreRelease {
	return &StoreRelease{
		Track:     track,
		Channel:   channel,
		Revision:  revision,
		Timestamp: ts.Unix(),
	}
}

// GetArtifactName tries to parse an artifact name from a store badge in repo's README
func GetArtifactName(readme string, artifactType string) (name string) {
	var re *regexp.Regexp
	switch artifactType {
	case "charm":
		re = charmBadgeRegexp
	case "snap":
		re = snapBadgeRegExp
	default:
		return ""
	}

	nameIndex := re.SubexpIndex("Name")
	matches := re.FindStringSubmatch(readme)
	if len(matches) > 0 {
		name = matches[nameIndex]
	}
	return name
}
