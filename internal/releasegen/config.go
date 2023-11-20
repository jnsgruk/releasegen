package releasegen

import (
	"github.com/jnsgruk/releasegen/internal/github"
	"github.com/jnsgruk/releasegen/internal/launchpad"
)

// Config represents the user provided configuration file.
type Config struct {
	Teams []*TeamConfig `yaml:"teams"`
}

// TeamConfig represents the configuration for a given real-life team.
type TeamConfig struct {
	Name            string             `mapstructure:"name"`
	GithubConfig    []github.OrgConfig `mapstructure:"github"`
	LaunchpadConfig launchpad.Config   `mapstructure:"launchpad"`
}
