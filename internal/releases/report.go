package releases

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"sync"

	"github.com/jnsgruk/releasegen/internal/config"
)

type ReleaseReport []*Team

// NewReleaseReport takes a given config, and generates the output JSON
func NewReleaseReport(conf *config.Config) ReleaseReport {
	teams := ReleaseReport{}
	var wg sync.WaitGroup

	// Iterate over the teams specified in the config file
	for _, t := range conf.Teams {
		wg.Add(1)

		go func(t *config.Team) {
			defer wg.Done()
			teams = append(teams, NewTeam(*t))
		}(t)
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
