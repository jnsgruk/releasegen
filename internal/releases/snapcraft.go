package releases

import (
  "fmt"
  "io/ioutil"
  "net/http"
  "time"

  "github.com/tidwall/gjson"
)

// FetchCharmInfo fetches the Json representing charm information by quering the Snapcraft API
func FetchCharmInfo(charmName string) (releases []*CharmRelease, err error) {
  // Query the Snapcraft API to obtain the charm information
  apiUrl := fmt.Sprintf("http://api.snapcraft.io/v2/charms/info/%s?fields=channel-map", charmName)
  res, err := http.Get(apiUrl)
  if err != nil {
    return nil, err
  }
  defer res.Body.Close()
  
  if res.StatusCode != 200 {
    return nil, fmt.Errorf("unexpected status code %d fetching %s", res.StatusCode, apiUrl)
  }
  
  // Pase the useful information from the response
  resBody, err := ioutil.ReadAll(res.Body)
  if err != nil {
    return nil, fmt.Errorf("cannot read response body while fetching %s", apiUrl)
  }

  jsonBody := string(resBody)
  tracks := gjson.Get(jsonBody, "channel-map.#.channel.track").Array()
  channels := gjson.Get(jsonBody, "channel-map.#.channel.risk").Array()
  releaseTime := gjson.Get(jsonBody, "channel-map.#.channel.released-at").Array()
  revision := gjson.Get(jsonBody, "channel-map.#.revision.revision").Array()
  
  // Create a CharmRelease array with the obtained information
  releases = []*CharmRelease{}
  for index, _ := range tracks {
    parsedTime, _ := time.Parse("2006-01-02T15:04:05.99-07:00", releaseTime[index].String())
    releases = append(releases, NewCharmRelease(
      tracks[index].String(),
      channels[index].String(),
      revision[index].Int(),
      parsedTime,
    ))
  }

  return releases, nil
}

