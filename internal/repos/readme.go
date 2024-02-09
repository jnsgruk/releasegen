package repos

import (
	"context"
	"log"
	"regexp"

	"github.com/jnsgruk/releasegen/internal/stores"
)

var (
	// ciBadgeRegexp is used to find action in Github CI Badges.
	ciBadgeRegexp = regexp.MustCompile(`(?P<Action>https://github.com/[\w-./]+)/badge.svg`)
	// charmBadgeRegexp is used to find a Charm's name in its CharmHub badge.
	charmBadgeRegexp = regexp.MustCompile(`https://charmhub.io/(?P<Name>[\w-]+)/badge.svg`)
	// snapBadgeRegexp is used to find a Snap's name in its Snapcraft badge.
	snapBadgeRegexp = regexp.MustCompile(`https://snapcraft.io/(?P<Name>[\w-]+)/badge.svg`)
)

type Readme struct {
	Body string
}

// GithubActions tries to extract Github Actions Badges from the README.
func (r *Readme) GithubActions() []string {
	// Parse the CI actions.
	actions := []string{}
	actionIndex := ciBadgeRegexp.SubexpIndex("Action")
	matches := ciBadgeRegexp.FindAllStringSubmatch(r.Body, -1)

	for _, actionMatch := range matches {
		// Check if the Action belongs to the repository.
		act := actionMatch[actionIndex]
		actions = append(actions, act)
	}

	return actions
}

// LinkedSnap parses the Readme body, and returns a StoreArtifact representing a snap
// if there is a Snapcraft.io badge in the Readme.
func (r *Readme) LinkedSnap(ctx context.Context) *stores.Artifact {
	// If the README has a Snapcraft Badge, fetch the snap information.
	if snapName := getArtifactName(r.Body, snapBadgeRegexp); snapName != "" {
		snapInfo, err := stores.FetchSnapDetails(ctx, snapName)
		if err != nil {
			log.Printf("failed to fetch snap package information for snap: %s", snapName)
			return nil
		}

		return stores.NewArtifact(snapName, snapInfo)
	}

	return nil
}

// LinkedCharm parses the Readme body, and returns a StoreArtifact representing a charm
// if there is a Charmhub.io badge in the Readme.
func (r *Readme) LinkedCharm(ctx context.Context) *stores.Artifact {
	// If the README has a Charmhub Badge, fetch the charm information
	if charmName := getArtifactName(r.Body, charmBadgeRegexp); charmName != "" {
		charmInfo, err := stores.FetchCharmDetails(ctx, charmName)
		if err != nil {
			log.Printf("failed to fetch charm information for charm: %s", charmName)
			return nil
		}

		return stores.NewArtifact(charmName, charmInfo)
	}

	return nil
}

// getArtifactName tries to parse an artifact name from a store badge in repo's README.
func getArtifactName(readme string, re *regexp.Regexp) string {
	nameIndex := re.SubexpIndex("Name")

	matches := re.FindStringSubmatch(readme)
	if len(matches) > 0 {
		return matches[nameIndex]
	}

	return ""
}
