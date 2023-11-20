package main

import (
	"errors"
	"fmt"
	"log"
	"runtime"

	"github.com/jnsgruk/releasegen/internal/releasegen"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	version = "dev"
	commit  = "dev"
	date    string
)

const (
	shortDesc = "releasegen - a utility for enumerating Github and Launchpad releases"
	longDesc  = ` releasegen is a utility for enumerating Github and Launchpad releases/tags
from specified Github Organisations or Launchpad project groups.

This tool is configured using a single file in one of the three following locations:

	- ./releasegen.yaml
	- $HOME/.config/releasegen.yaml
	- /etc/releasegen/releasegen.yaml

For more details on the configuration format, see the homepage below.

Prior to launching, you must also set an environment variable named RELEASEGEN_TOKEN whose
contents is a Github Personal Access token with sufficient rights over any org you wish to
query. 

For example:

	export RELEASEGEN_TOKEN=ghp_aBcDeFgHiJkLmNoPqRsTuVwXyZ

You can create a Personal Access Token at: https://github.com/settings/tokens

Homepage: https://github.com/jnsgruk/releasegen
`
)

// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:          "releasegen",
	Version:      buildVersion(version, commit, date),
	Short:        shortDesc,
	Long:         longDesc,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		err := viper.ReadInConfig()
		if err != nil {
			if errors.As(err, &viper.ConfigFileNotFoundError{}) {
				return errors.New("no config file found, see 'releasegen --help' for details")
			}
			return errors.New("error parsing releasegen config file")
		}

		conf := &releasegen.Config{}

		err = viper.Unmarshal(conf)
		if err != nil {
			return errors.New("error parsing releasegen config file")
		}

		if viper.Get("token") == nil {
			return errors.New("environment variable RELEASEGEN_TOKEN not set")
		}

		teams := releasegen.GenerateReport(conf)
		teams.Dump()

		return nil
	},
}

// buildVersion writes a multiline version string from the specified version variables.
func buildVersion(version, commit, date string) string {
	result := version

	if commit != "" {
		result = fmt.Sprintf("%s\ncommit: %s", result, commit)
	}

	if date != "" {
		result = fmt.Sprintf("%s\nbuilt at: %s", result, date)
	}

	result = fmt.Sprintf("%s\ngoos: %s\ngoarch: %s", result, runtime.GOOS, runtime.GOARCH)

	return result
}

func main() {
	// Set the default config file name/type.
	viper.SetConfigName("releasegen")
	viper.SetConfigType("yaml")

	// Add some default config paths.
	viper.AddConfigPath(".")
	viper.AddConfigPath("$HOME/.config")
	viper.AddConfigPath("/etc/releasegen")

	// Setup environment variable parsing.
	viper.SetEnvPrefix("releasegen")
	viper.MustBindEnv("token")

	if err := rootCmd.Execute(); err != nil {
		log.Fatalln(err.Error())
	}
}
