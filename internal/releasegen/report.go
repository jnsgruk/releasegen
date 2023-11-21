package releasegen

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"

	"github.com/jnsgruk/releasegen/internal/repos"
)

// ReleaseReport is a representation of the output of releasegen.
type ReleaseReport []*TeamDetails

// GenerateReport takes a given config, and generates the output JSON.
func GenerateReport(conf *Config) ReleaseReport {
	teams := ReleaseReport{}

	// Iterate over the teams specified in the config file.
	for _, t := range conf.Teams {
		team := &Team{
			Details: &TeamDetails{
				Name:  t.Name,
				Repos: []repos.RepoDetails{},
			},
			config:      *t,
			githubToken: conf.githubToken,
		}
		teams = append(teams, team.Details)

		err := team.Process()
		if err != nil {
			log.Printf("error processing team '%s': %v", team.Details.Name, err)
		}
	}

	return teams
}

// Dump is used to create a pretty-printed JSON version of a ReleaseReport.
func (r ReleaseReport) Dump() {
	buffer := &bytes.Buffer{}
	encoder := json.NewEncoder(buffer)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "   ")

	if err := encoder.Encode(r); err != nil {
		log.Fatalln("unable to encode report to JSON")
	}

	//nolint:forbidigo
	fmt.Println(buffer.String())
}
