package github

import (
	"context"

	gh "github.com/google/go-github/v54/github"
	"golang.org/x/oauth2"
)

// OrgConfig contains fields used in releasegen's config.yaml file to configure
// its behaviour when generating reports about Github repositories.
type OrgConfig struct {
	Org          string   `mapstructure:"org"`
	Teams        []string `mapstructure:"teams"`
	IgnoredRepos []string `mapstructure:"ignores"`

	ghClient *gh.Client
	token    string
}

// GithubClient returns either a new instance of the Github client, or a previously
// initialised client.
func (oc *OrgConfig) GithubClient() *gh.Client {
	if oc.ghClient == nil {
		ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: oc.token})
		tc := oauth2.NewClient(context.Background(), ts)
		oc.ghClient = gh.NewClient(tc)
	}

	return oc.ghClient
}

// SetGithubToken enables the setting of the Github token for the Github org.
func (oc *OrgConfig) SetGithubToken(token string) {
	oc.token = token
}
