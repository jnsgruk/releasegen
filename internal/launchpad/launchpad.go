package launchpad

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html"
)

// EnumerateProjectGroup lists the projects that are part of the specified project group
func EnumerateProjectGroup(name string) ([]Project, error) {
	url := fmt.Sprintf("https://api.staging.launchpad.net/devel/%s/projects", name)
	log.Printf("processing launchpad project group: %s", url)

	// TODO: Add a retry here?
	client := http.Client{
		Timeout: time.Second * 10, // Timeout after 5 seconds
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	res, getErr := client.Do(req)
	if getErr != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, readErr := ioutil.ReadAll(res.Body)
	if readErr != nil {
		return nil, err
	}

	pg := ProjectGroupProjectsResponse{}
	jsonErr := json.Unmarshal(body, &pg)
	if jsonErr != nil {
		return nil, err
	}

	return pg.Entries, nil
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

type Tag struct {
	Name      string
	Commit    string
	Timestamp *time.Time
}

type Commit struct {
	Commit    string
	Timestamp *time.Time
}

func GetTags(project string, page *goquery.Document) []*Tag {
	tags := []*Tag{}
	// TODO(jnsgruk): Reduce some iteration here - find by Text perhaps?
	page.Find("tr.nohover").Each(func(i int, s *goquery.Selection) {
		// Find the header for the tags table
		if s.Text() == "TagDownloadAuthorAge" {
			s.NextUntil("tr.nohover").EachWithBreak(func(i int, t *goquery.Selection) bool {
				// Only get the first three tags
				if i > 2 {
					return false
				}
				// Find all the 'a' tags on the row
				hrefs := t.Find("a")

				for i := 0; i < hrefs.Length(); i += 2 {
					// The first link is in the first column. Get the tag name from the URL
					tagHref := hrefs.Get(i)
					tagName := strings.Split(getNodeAttr("href", tagHref), "=")[1]

					if strings.HasPrefix(tagName, "rev") {
						t, err := ParseCommitPage(project, tagName)
						if err != nil {
							return false
						}

						// We only consider tags with the name 'rev...'
						tags = append(tags, &Tag{
							Name:      tagName,
							Commit:    t.Commit,
							Timestamp: t.Timestamp,
						})
					}
				}
				return true
			})
		}
	})
	return tags
}

func ParseCommitPage(repo string, tag string) (*Commit, error) {
	url := fmt.Sprintf("https://git.launchpad.net/%s/commit", repo)
	if tag != "" {
		url = fmt.Sprintf("%s/?h=%s", url, tag)
	}

	doc, err := FetchWebDocument(url)
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

func CalculateNewCommits(repo string) int {
	url := fmt.Sprintf("https://git.launchpad.net/%s/log", repo)

	doc, err := FetchWebDocument(url)
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

func getNodeAttr(attr string, node *html.Node) string {
	for _, a := range node.Attr {
		if a.Key == attr {
			return a.Val
		}
	}
	return ""
}

// FetchWebDocument fetches a URL and returns a goquery.Document for scraping
func FetchWebDocument(url string) (*goquery.Document, error) {
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
