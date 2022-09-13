package releases

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"sync"

	"github.com/jnsgruk/releasegen/internal/config"
)

type ReleaseReport []*TeamInfo

// GenerateReport takes a given config, and generates the output JSON
func GenerateReport(conf *config.Config) ReleaseReport {
	teams := ReleaseReport{}
	var wg sync.WaitGroup

	// Iterate over the teams specified in the config file
	for _, t := range conf.Teams {
		team := NewTeam(*t)
		teams = append(teams, team.Info())

		wg.Add(1)
		go func() {
			defer wg.Done()
			err := team.Process()
			if err != nil {
				log.Printf("error processing team: %v", err)
			}
		}()
	}
	wg.Wait()

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
