package releasegen

import (
	"github.com/jnsgruk/releasegen/internal/github"
	"github.com/jnsgruk/releasegen/internal/launchpad"
)

// ReleasegenConfig represents the user provided configuration file
type ReleasegenConfig struct {
	Teams []*TeamConfig `yaml:"teams"`
}

// TeamConfig represents the configuration for a given real-life team
type TeamConfig struct {
	Name            string                    `mapstructure:"name"`
	GithubConfig    []github.GithubOrgConfig  `mapstructure:"github"`
	LaunchpadConfig launchpad.LaunchpadConfig `mapstructure:"launchpad"`
}