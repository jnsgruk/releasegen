package gitea

import (
	"code.gitea.io/sdk/gitea"
)

// OrgConfig contains fields used in releasegen's config.yaml file to configure
// its behaviour when generating reports about Gitea repositories.
type OrgConfig struct {
	Org          string   `mapstructure:"org"`
	URL          string   `mapstructure:"url"`
	IgnoredRepos []string `mapstructure:"ignores"`

	gtClient *gitea.Client
}

// GiteaClient returns either a new instance of the Gitea client, or a
// previously initialised client.
func (oc *OrgConfig) GiteaClient() (*gitea.Client, error) {
	if oc.gtClient == nil {
		var err error
		oc.gtClient, err = gitea.NewClient(oc.URL)
		if err != nil {
			return nil, err
		}
	}

	return oc.gtClient, nil
}
