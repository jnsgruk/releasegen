package launchpad

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/tidwall/gjson"
)

var errEnumerateProjectGroup = errors.New("error enumerating project group")

// enumerateProjectGroupError builds a wrapped error that might occur when enumerating a group.
func enumerateProjectGroupError(err error) error {
	return fmt.Errorf("%w: %s", errEnumerateProjectGroup, err.Error())
}

// Config contains fields used in releasegen's config.yaml file to configure
// its behaviour when generating reports about Launchpad repositories.
type Config struct {
	ProjectGroups []string `mapstructure:"project-groups"`
	IgnoredRepos  []string `mapstructure:"ignores"`
}

// enumerateProjectGroup lists the projects that are part of the specified project group.
func enumerateProjectGroup(projectGroup string) (projects []string, err error) {
	url := fmt.Sprintf("https://api.launchpad.net/devel/%s/projects", projectGroup)

	client := http.Client{Timeout: time.Second * 5}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil)
	if err != nil {
		return nil, enumerateProjectGroupError(err)
	}

	res, getErr := client.Do(req)
	if getErr != nil {
		return nil, enumerateProjectGroupError(err)
	}
	defer res.Body.Close()

	body, readErr := io.ReadAll(res.Body)
	if readErr != nil {
		return nil, enumerateProjectGroupError(err)
	}

	// Parse the result as JSON, grab the "entries" key.
	result := gjson.Get(string(body), "entries")

	// Iterate over the entries.
	result.ForEach(func(key, value gjson.Result) bool {
		// If the entry doesn't use Git as it's VCS, move on.
		if vcs := gjson.Get(value.Raw, "vcs").String(); vcs == "Git" {
			// If the project does use Git, add the project name to the output array.
			projects = append(projects, gjson.Get(value.Raw, "name").String())
		}

		return true
	})

	return projects, nil
}
