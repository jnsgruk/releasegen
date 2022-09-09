package config

type Config struct {
	Teams []*Team `yaml:"teams"`
}

type Team struct {
	Name      string          `mapstructure:"name"`
	Github    []GithubOrg     `mapstructure:"github"`
	Launchpad LaunchpadConfig `mapstructure:"launchpad"`
}

type GithubOrg struct {
	Name    string   `mapstructure:"org"`
	Teams   []string `mapstructure:"teams"`
	Ignores []string `mapstructure:"ignores"`
}

type LaunchpadConfig struct {
	ProjectGroups []string `mapstructure:"project-groups"`
	Ignores       []string `mapstructure:"ignores"`
}
