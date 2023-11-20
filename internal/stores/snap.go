package stores

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/tidwall/gjson"
)

// FetchSnapDetails fetches the Json representing charm information by querying the Snapcraft API.
func FetchSnapDetails(name string) (*ArtifactDetails, error) {
	// Query the Snapcraft API to obtain the charm information.
	apiURL := fmt.Sprintf("http://api.snapcraft.io/v2/snaps/info/%s?fields=channel-map,revision,store-url", name)

	client := &http.Client{}
	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, apiURL, nil)
	// According to: https://api.snapcraft.io/docs/refresh.html
	// The only valid 'Snap-Device-Series' to date is '16', and the
	// header must be set in order for the request to be successful.
	req.Header.Set("Snap-Device-Series", "16")

	res, err := client.Do(req)
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
	storeURL := gjson.Get(jsonBody, "snap.store-url").String()
	tracks := gjson.Get(jsonBody, "channel-map.#.channel.track").Array()
	channels := gjson.Get(jsonBody, "channel-map.#.channel.risk").Array()
	releaseTimes := gjson.Get(jsonBody, "channel-map.#.channel.released-at").Array()
	revisions := gjson.Get(jsonBody, "channel-map.#.revision").Array()

	return &ArtifactDetails{storeURL, tracks, channels, releaseTimes, revisions}, nil
}
