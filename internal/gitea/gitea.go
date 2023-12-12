package gitea

// OrgConfig contains fields used in releasegen's config.yaml file to configure
// its behaviour when generating reports about Gitea repositories.
type OrgConfig struct {
	Org          string   `mapstructure:"org"`
	URL          string   `mapstructure:"url"`
	IgnoredRepos []string `mapstructure:"ignores"`
}
