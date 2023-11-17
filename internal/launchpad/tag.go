package launchpad

import (
	"fmt"
	"time"
)

// Tag is a representation of a git tag in Launchpad
type Tag struct {
	Name      string
	Commit    string
	Timestamp *time.Time

	project string
}

// Process populates the tag with details of the relevant commit
func (t *Tag) Process() error {
	// Construct a URL for the project commit page, including a tag if specified
	url := fmt.Sprintf("https://git.launchpad.net/%s/commit/?h=%s", t.project, t.Name)

	doc, err := parseWebpage(url)
	if err != nil {
		return fmt.Errorf("could not fetch tag page %s: %v", url, err)
	}

	// Find the commit hash for the tag
	commitTable := doc.Find("table.commit-info")
	commit := commitTable.Find("a").First().Text()
	t.Commit = commit

	// Find the timestamp for the tag/commit in question
	ts := commitTable.Find("td.right").First().Text()

	// Parse the Launchpad timestamp into a time.Time
	timestamp, err := time.Parse("2006-01-02 15:04:05 -0700", ts)
	if err != nil {
		return err
	}
	t.Timestamp = &timestamp

	return nil
}
