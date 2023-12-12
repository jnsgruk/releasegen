package gitea

type RepoConfig struct {
	MonorepoSources []string `mapstructure:"monorepo-folders"`
}

// OrgConfig contains fields used in releasegen's config.yaml file to configure
// its behaviour when generating reports about Gitea repositories.
type OrgConfig struct {
	Org          string                `mapstructure:"org"`
	URL          string                `mapstructure:"url"`
	IncludeRepos map[string]RepoConfig `mapstructure:"includes"`
	IgnoredRepos []string              `mapstructure:"ignores"`
}
