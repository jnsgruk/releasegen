package stores

import (
	"fmt"
	"io"
	"net/http"
	"regexp"

	"github.com/tidwall/gjson"
)

// charmBadgeRegexp is used to find a Charm's name in its CharmHub badge
var charmBadgeRegexp = regexp.MustCompile(`https://charmhub.io/(?P<Name>[a-zA-Z0-9_-]+)/badge.svg`)

// GetCharmName parses a charm name from a Charmhub badge in repo's README
func GetCharmName(readme string) (name string) {
	return getArtifactName(readme, charmBadgeRegexp)
}

// FetchCharmInfo fetches the Json representing charm information by querying the Charmhub API
func FetchCharmInfo(name string) (*artifactInfoResult, error) {
	apiUrl := fmt.Sprintf("http://api.snapcraft.io/v2/charms/info/%s?fields=channel-map", name)
	res, err := http.Get(apiUrl)
	if err != nil {
		return nil, fmt.Errorf("cannot query the snapcraft api: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return nil, fmt.Errorf("unexpected status code %d fetching %s", res.StatusCode, apiUrl)
	}

	// Parse the useful information from the response
	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("cannot read response body while fetching %s", apiUrl)
	}

	jsonBody := string(resBody)
	tracks := gjson.Get(jsonBody, "channel-map.#.channel.track").Array()
	channels := gjson.Get(jsonBody, "channel-map.#.channel.risk").Array()
	releaseTimes := gjson.Get(jsonBody, "channel-map.#.channel.released-at").Array()
	revisions := gjson.Get(jsonBody, "channel-map.#.revision.revision").Array()

	return &artifactInfoResult{tracks, channels, releaseTimes, revisions}, nil
}
