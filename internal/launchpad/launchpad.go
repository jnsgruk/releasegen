package launchpad

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/tidwall/gjson"
)

const launchpadTimeout = 5 * time.Second

// Config contains fields used in releasegen's config.yaml file to configure
// its behaviour when generating reports about Launchpad repositories.
type Config struct {
	ProjectGroups []string `mapstructure:"project-groups"`
	IgnoredRepos  []string `mapstructure:"ignores"`
}

// enumerateProjectGroup lists the projects that are part of the specified project group.
func enumerateProjectGroup(ctx context.Context, projectGroup string) ([]string, error) {
	url := fmt.Sprintf("https://api.launchpad.net/devel/%s/projects", projectGroup)

	client := http.Client{Timeout: launchpadTimeout}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("error enumerating project group '%s': %w", projectGroup, err)
	}

	res, getErr := client.Do(req)
	if getErr != nil {
		return nil, fmt.Errorf("error enumerating project group '%s': %w", projectGroup, err)
	}
	defer res.Body.Close()

	body, readErr := io.ReadAll(res.Body)
	if readErr != nil {
		return nil, fmt.Errorf("error enumerating project group '%s': %w", projectGroup, err)
	}

	// Parse the result as JSON, grab the "entries" key.
	result := gjson.Get(string(body), "entries")
	projects := []string{}

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
