package releases

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/tidwall/gjson"
)

// SnapRelease represents a charm release on Snapcraft.io
type SnapRelease struct {
	Track     string `json:"track"`
	Channel   string `json:"channel"`
	Revision  int64  `json:"revision"`
	Timestamp int64  `json:"timestamp"`
}

// NewSnapRelease is used for constructing a valid SnapRelease
func NewSnapRelease(track string, channel string, revision int64, ts time.Time) *SnapRelease {
	return &SnapRelease{
		Track:     track,
		Channel:   channel,
		Revision:  revision,
		Timestamp: ts.Unix(),
	}
}

// SnapInfo holds all Charm information
type SnapInfo struct {
	Name     string         `json:"name"`
	Url      string         `json:"url"`
	Releases []*SnapRelease `json:"releases"`
	Channels []string       `json:"channels"`
	Tracks   []string       `json:"tracks"`
}

// FetchSnapInfo fetches the Json representing charm information by querying the Snapcraft API
func (c *SnapInfo) FetchSnapInfo() (err error) {
	// Query the Snapcraft API to obtain the charm information
	apiUrl := fmt.Sprintf("http://api.snapcraft.io/v2/snaps/info/%s?fields=channel-map,revision", c.Name)

	client := &http.Client{}
	req, _ := http.NewRequest("GET", apiUrl, nil)
	// According to: https://api.snapcraft.io/docs/refresh.html
	// The only valid 'Snap-Device-Series' to date is '16', and the
	// header must be set in order for the request to be successful.
	req.Header.Set("Snap-Device-Series", "16")
	res, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("cannot query the snapcraft api: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return fmt.Errorf("unexpected status code %d fetching %s", res.StatusCode, apiUrl)
	}

	// Parse the useful information from the response
	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("cannot read response body while fetching %s", apiUrl)
	}

	jsonBody := string(resBody)
	tracks := gjson.Get(jsonBody, "channel-map.#.channel.track").Array()
	channels := gjson.Get(jsonBody, "channel-map.#.channel.risk").Array()
	releaseTime := gjson.Get(jsonBody, "channel-map.#.channel.released-at").Array()
	revision := gjson.Get(jsonBody, "channel-map.#.revision").Array()

	// Create a SnapRelease array with the obtained information
	for index := range tracks {
		parsedTime, _ := time.Parse("2006-01-02T15:04:05.99-07:00", releaseTime[index].String())
		c.Releases = append(c.Releases, NewSnapRelease(
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
