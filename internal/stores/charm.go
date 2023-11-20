package stores

import (
	"fmt"
	"io"
	"net/http"

	"github.com/tidwall/gjson"
)

// FetchCharmDetails fetches the Json representing charm information by querying the Charmhub API.
func FetchCharmDetails(name string) (*artifactDetails, error) {
	apiUrl := fmt.Sprintf("http://api.snapcraft.io/v2/charms/info/%s?fields=channel-map,result.store-url", name)
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
	storeUrl := gjson.Get(jsonBody, "result.store-url").String()
	tracks := gjson.Get(jsonBody, "channel-map.#.channel.track").Array()
	channels := gjson.Get(jsonBody, "channel-map.#.channel.risk").Array()
	releaseTimes := gjson.Get(jsonBody, "channel-map.#.channel.released-at").Array()
	revisions := gjson.Get(jsonBody, "channel-map.#.revision.revision").Array()

	return &artifactDetails{storeUrl, tracks, channels, releaseTimes, revisions}, nil
}
