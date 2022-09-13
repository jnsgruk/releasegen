package launchpad

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

// ProjectGroupProjectsResponse is a representation of the response given when querying the
// projects associated with a Project Group on Launchpad
type ProjectGroupProjectsResponse struct {
	Start            int       `json:"start"`
	TotalSize        int       `json:"total_size"`
	Entries          []Project `json:"entries"`
	ResourceTypeLink string    `json:"resource_type_link"`
}

// Projects lists the projects that are part of the specified project group
func EnumerateProjectGroup(projectGroup string) ([]Project, error) {
	url := fmt.Sprintf("https://api.launchpad.net/devel/%s/projects", projectGroup)

	// TODO: Add a retry here?
	client := http.Client{Timeout: time.Second * 5}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	res, getErr := client.Do(req)
	if getErr != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, readErr := ioutil.ReadAll(res.Body)
	if readErr != nil {
		return nil, err
	}

	pg := ProjectGroupProjectsResponse{}
	jsonErr := json.Unmarshal(body, &pg)
	if jsonErr != nil {
		return nil, err
	}

	return pg.Entries, nil
}
