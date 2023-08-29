package releases

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"

	"github.com/jnsgruk/releasegen/internal/config"
)

type ReleaseReport []*TeamInfo

// GenerateReport takes a given config, and generates the output JSON
func GenerateReport(conf *config.Config) ReleaseReport {
	teams := ReleaseReport{}

	// Iterate over the teams specified in the config file
	for _, t := range conf.Teams {
		team := NewTeam(*t)
		teams = append(teams, team.Info())

		err := team.Process()
		if err != nil {
			log.Printf("error processing team '%s': %v", team.Info().Name, err)
		}
	}

	return teams
}

func (r ReleaseReport) Dump() {
	buffer := &bytes.Buffer{}
	encoder := json.NewEncoder(buffer)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "   ")
	err := encoder.Encode(r)
	if err != nil {
		log.Fatalln("unable to encode report to JSON")
	}
	fmt.Println(buffer.String())
}
