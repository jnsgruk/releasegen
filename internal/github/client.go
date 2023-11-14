package github

import (
	"context"
	"log"

	"github.com/gofri/go-github-ratelimit/github_ratelimit"
	gh "github.com/google/go-github/v54/github"
	"github.com/spf13/viper"
	"golang.org/x/oauth2"
)

var ghClient *gh.Client

// githubClient returns either a new instance of the Github client, or a previously
// initialised client.
func githubClient() *gh.Client {
	if ghClient == nil {
		log.Println("creating new Github client")
		ctx := context.Background()
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: viper.Get("token").(string)},
		)
		tc := oauth2.NewClient(ctx, ts)

		rateLimiter, err := github_ratelimit.NewRateLimitWaiterClient(tc.Transport)
		if err != nil {
			panic(err)
		}

		ghClient = gh.NewClient(rateLimiter)
	}
	return ghClient
}

// parseApiError is used to detect rate limiting errors and more
// accurately report them in the logs.
func parseApiError(err error) string {
	if _, ok := err.(*gh.RateLimitError); ok {
		return "rate limit exceeded"
	}
	if _, ok := err.(*gh.AbuseRateLimitError); ok {
		return "secondary rate limit exceeded"
	}
	return err.Error()
}
