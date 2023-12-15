package releasegen

import (
	"github.com/jnsgruk/releasegen/internal/gitea"
	"github.com/jnsgruk/releasegen/internal/github"
	"github.com/jnsgruk/releasegen/internal/launchpad"
)

// Config represents the user provided configuration file.
type Config struct {
	Teams       []*TeamConfig `yaml:"teams"`
	githubToken string
}

// SetGithubToken enables the setting of the Github token from outside.
func (c *Config) SetGithubToken(token string) {
	c.githubToken = token
}

// TeamConfig represents the configuration for a given real-life team.
type TeamConfig struct {
	Name            string             `mapstructure:"name"`
	GithubConfig    []github.OrgConfig `mapstructure:"github"`
	LaunchpadConfig launchpad.Config   `mapstructure:"launchpad"`
	GiteaConfig     []gitea.OrgConfig  `mapstructure:"gitea"`
}
