package stores

import (
	"errors"
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
		return nil, errors.New("failed to contact the store api")
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, errors.New("unexpected status code while fetching store resource")
	}

	// Parse the useful information from the response.
	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, errors.New("failed to read details about artifact from the store")
	}

	jsonBody := string(resBody)
	storeUrl := gjson.Get(jsonBody, "result.store-url").String()
	tracks := gjson.Get(jsonBody, "channel-map.#.channel.track").Array()
	channels := gjson.Get(jsonBody, "channel-map.#.channel.risk").Array()
	releaseTimes := gjson.Get(jsonBody, "channel-map.#.channel.released-at").Array()
	revisions := gjson.Get(jsonBody, "channel-map.#.revision.revision").Array()

	return &artifactDetails{storeUrl, tracks, channels, releaseTimes, revisions}, nil
}
