package stores

import (
	"fmt"
	"io"
	"net/http"
	"regexp"

	"github.com/tidwall/gjson"
)

// snapBadgeRegexp is used to find a Snap's name in its Snapcraft badge
var snapBadgeRegexp = regexp.MustCompile(`https://snapcraft.io/(?P<Name>[a-zA-Z0-9_-]+)/badge.svg`)

// GetSnapName tries to parse a snap name from a Snapcraft badge in repo's README
func GetSnapName(readme string) (name string) {
	return getArtifactName(readme, snapBadgeRegexp)
}

// FetchSnapInfo fetches the Json representing charm information by querying the Snapcraft API
func FetchSnapInfo(name string) (*artifactInfoResult, error) {
	// Query the Snapcraft API to obtain the charm information
	apiUrl := fmt.Sprintf("http://api.snapcraft.io/v2/snaps/info/%s?fields=channel-map,revision", name)

	client := &http.Client{}
	req, _ := http.NewRequest("GET", apiUrl, nil)
	// According to: https://api.snapcraft.io/docs/refresh.html
	// The only valid 'Snap-Device-Series' to date is '16', and the
	// header must be set in order for the request to be successful.
	req.Header.Set("Snap-Device-Series", "16")
	res, err := client.Do(req)
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
	revisions := gjson.Get(jsonBody, "channel-map.#.revision").Array()

	return &artifactInfoResult{tracks, channels, releaseTimes, revisions}, nil
}