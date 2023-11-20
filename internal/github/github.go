package github

import (
	"context"
	"log"

	ghrl "github.com/gofri/go-github-ratelimit/github_ratelimit"
	gh "github.com/google/go-github/v54/github"
	"github.com/spf13/viper"
	"golang.org/x/oauth2"
)

var ghClient *gh.Client

// OrgConfig contains fields used in releasegen's config.yaml file to configure
// its behaviour when generating reports about Github repositories.
type OrgConfig struct {
	Org          string   `mapstructure:"org"`
	Teams        []string `mapstructure:"teams"`
	IgnoredRepos []string `mapstructure:"ignores"`
}

// githubClient returns either a new instance of the Github client, or a previously
// initialised client.
func githubClient() *gh.Client {
	if ghClient == nil {
		log.Println("creating new Github client")

		ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: viper.GetString("token")})
		tc := oauth2.NewClient(context.Background(), ts)

		rateLimiter, err := ghrl.NewRateLimitWaiterClient(tc.Transport)
		if err != nil {
			panic(err)
		}

		ghClient = gh.NewClient(rateLimiter)
	}

	return ghClient
}
