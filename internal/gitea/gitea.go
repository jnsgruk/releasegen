// Package gitea implements support for Gitea repositories using the Gitea API client.
package gitea

// OrgConfig contains fields used in releasegen's config.yaml file to configure
// its behaviour when generating reports about Gitea repositories.
type OrgConfig struct {
	Org      string                `mapstructure:"org"`
	URL      string                `mapstructure:"url"`
	Includes map[string]RepoConfig `mapstructure:"includes"`
	Ignores  []string              `mapstructure:"ignores"`
}

type RepoConfig struct {
	MonorepoFolders []string `mapstructure:"monorepo-folders"`
}
