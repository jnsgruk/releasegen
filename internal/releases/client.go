package releases

import (
	"context"
	"log"

	"github.com/google/go-github/v47/github"
	"github.com/spf13/viper"
	"golang.org/x/oauth2"
)

var ghClient *github.Client

func githubClient() *github.Client {
	if ghClient == nil {
		log.Println("creating new Github client")
		ctx := context.Background()
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: viper.Get("token").(string)},
		)
		tc := oauth2.NewClient(ctx, ts)
		ghClient = github.NewClient(tc)
	}
	return ghClient
}
