package launchpad

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type Tag struct {
	Name      string
	Commit    string
	Timestamp *time.Time
}

type Commit struct {
	Commit    string
	Timestamp *time.Time
}

func GetTags(project string, page *goquery.Document) (tags []*Tag) {
	// Get the row in the table that represents the header before the tags are listed
	tagRowHeader := page.Find("tr.nohover").FilterFunction(func(_ int, s *goquery.Selection) bool {
		return s.Text() == "TagDownloadAuthorAge"
	})

	tagRowHeader.NextUntil("tr.nohover").EachWithBreak(func(_ int, row *goquery.Selection) bool {
		// Only get the first three tags
		if len(tags) >= 3 {
			return false
		}

		// Iterate over the links, each one is either a link to a tag or a commit
		row.Find("a").Each(func(i int, l *goquery.Selection) {
			// Skip odd numbers, as these are links to commits
			if i%2 != 0 {
				return
			}

			// Grab the tag name from the only query parameter in the url
			href := l.AttrOr("href", "")
			tagName := strings.Split(href, "=")[1]

			// We only consider tags with the name 'rev...' so bail if this isn't one
			if !strings.HasPrefix(tagName, "rev") {
				return
			}

			commit, err := GetCommit(project, tagName)
			if err != nil {
				return
			}

			tags = append(tags, &Tag{
				Name:      tagName,
				Commit:    commit.Commit,
				Timestamp: commit.Timestamp,
			})
		})
		return true
	})
	return tags
}

func GetCommit(repo string, tag string) (*Commit, error) {
	url := fmt.Sprintf("https://git.launchpad.net/%s/commit", repo)
	if tag != "" {
		url = fmt.Sprintf("%s/?h=%s", url, tag)
	}

	doc, err := ParseWebpage(url)
	if err != nil {
		return nil, fmt.Errorf("could not fetch tag page %s: %v", url, err)
	}

	commitTable := doc.Find("table.commit-info")
	commit := commitTable.Find("a").First().Text()
	ts := commitTable.Find("td.right").First().Text()

	timestamp, err := time.Parse("2006-01-02 15:04:05 -0700", ts)
	if err != nil {
		log.Printf("%v", err)
	}

	return &Commit{
		Commit:    commit,
		Timestamp: &timestamp,
	}, nil
}

// GetDefaultBranch parses a repository's default git branch from its cgit page
func GetDefaultBranch(page *goquery.Document) (branch string) {
	// Find an option element with the selected attribute.
	// Launchpad Git pages only contain one option element for selecting branch, the selected
	// branch at page load time is the default branch
	page.Find("option[selected=selected]").Each(func(i int, s *goquery.Selection) {
		branch = s.Text()
	})
	return branch
}

// GetNewCommits parses the git log page for a Launchpad project and returns the number of
// commits that have happened on the default branch since the last tag
func GetNewCommits(repo string) int {
	url := fmt.Sprintf("https://git.launchpad.net/%s/log", repo)

	doc, err := ParseWebpage(url)
	if err != nil {
		return -1
	}

	commitTable := doc.Find("table.list")
	branchDecorationRow := commitTable.Find("a.branch-deco").First().Parent().Parent().Parent()
	tagDecorationRow := commitTable.Find("a.tag-deco").First().Parent().Parent().Parent()

	if tagDecorationRow.Text() == branchDecorationRow.Text() {
		// If the decorations are on the same row, there are no commits between last tag and branch
		return 0
	} else {
		// Return the number of commits between the latest tag and the default branch
		return branchDecorationRow.NextUntilSelection(tagDecorationRow).Length() + 1
	}
}

// ParseWebpage fetches a URL and returns a goquery.Document for scraping
func ParseWebpage(url string) (*goquery.Document, error) {
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return nil, fmt.Errorf("unexpected status code %d fetching %s", res.StatusCode, url)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	return doc, err
}
