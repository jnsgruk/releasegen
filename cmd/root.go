package cmd

import (
	"log"
	"os"

	"github.com/jnsgruk/releasegen/internal/config"
	"github.com/jnsgruk/releasegen/internal/releases"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "releasegen",
	Short: "releasegen - a utility for enumerating Github and Launchpad releases",
	Long:  "releasegen - a utility for enumerating Github and Launchpad releases",
	Run: func(cmd *cobra.Command, args []string) {
		err := viper.ReadInConfig()
		if err != nil {
			if _, ok := err.(viper.ConfigFileNotFoundError); ok {
				log.Fatalln("config file 'releasegen.yaml' not found")
			} else {
				log.Fatalln("error parsing config file")
			}
		}

		conf := &config.Config{}
		err = viper.Unmarshal(conf)
		if err != nil {
			log.Fatalf("unable to decode into config struct, %v\n", err)
		}

		if viper.Get("token") == nil {
			log.Fatalln("environment variable RELEASEGEN_TOKEN not set")
		}

		teams := releases.GenerateReport(conf)
		teams.Dump()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute(version string) {
	rootCmd.Version = version
	rootCmd.SetVersionTemplate("releasegen\nversion: {{.Version}}")

	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Set the default config file name/type
	viper.SetConfigName("releasegen")
	viper.SetConfigType("yaml")

	// Add some default config paths
	viper.AddConfigPath(".")
	viper.AddConfigPath("$HOME/.config")
	viper.AddConfigPath("/etc/releasegen")

	// Setup environment variable parsing
	viper.SetEnvPrefix("releasegen")
	viper.MustBindEnv("token")
}
