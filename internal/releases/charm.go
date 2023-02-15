package releases

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/tidwall/gjson"
)

// CharmInfo holds all Charm information
type CharmInfo struct {
	Name     string          `json:"name"`
	Url      string          `json:"url"`
	Releases []*CharmRelease `json:"releases"`
	Channels []string        `json:"channels"`
	Tracks   []string        `json:"tracks"`
}

// FetchCharmInfo fetches the Json representing charm information by quering the Snapcraft API
func (c *CharmInfo) FetchCharmInfo() (err error) {
	// Query the Snapcraft API to obtain the charm information
	apiUrl := fmt.Sprintf("http://api.snapcraft.io/v2/charms/info/%s?fields=channel-map", c.Name)
	res, err := http.Get(apiUrl)
	if err != nil {
		return fmt.Errorf("cannot query the snapcraft api: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return fmt.Errorf("unexpected status code %d fetching %s", res.StatusCode, apiUrl)
	}

	// Pase the useful information from the response
	resBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("cannot read response body while fetching %s", apiUrl)
	}

	jsonBody := string(resBody)
	tracks := gjson.Get(jsonBody, "channel-map.#.channel.track").Array()
	channels := gjson.Get(jsonBody, "channel-map.#.channel.risk").Array()
	releaseTime := gjson.Get(jsonBody, "channel-map.#.channel.released-at").Array()
	revision := gjson.Get(jsonBody, "channel-map.#.revision.revision").Array()

	// Create a CharmRelease array with the obtained information
	for index := range tracks {
		parsedTime, _ := time.Parse("2006-01-02T15:04:05.99-07:00", releaseTime[index].String())
		c.Releases = append(c.Releases, NewCharmRelease(
			tracks[index].String(),
			channels[index].String(),
			revision[index].Int(),
			parsedTime,
		))
	}

	// Populate Channels and Tracks from the releases details
	tracksMap := map[string]bool{}
	for _, track := range tracks {
		if _, exists := tracksMap[track.String()]; !exists {
			tracksMap[track.String()] = true
			c.Tracks = append(c.Tracks, track.String())
		}
	}
	channelsMap := map[string]bool{}
	for _, channel := range channels {
		if _, exists := channelsMap[channel.String()]; !exists {
			channelsMap[channel.String()] = true
			c.Channels = append(c.Channels, channel.String())
		}
	}

	return nil
}
