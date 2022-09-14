package launchpad

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// Project is a representation of a Launchpad project
type Project struct {
	name          string
	defaultBranch string
	tags          []*Tag

	projectPage *goquery.Document
}

// Name reports the name of the Launchpad project
func (p *Project) Name() string { return p.name }

// NewProject returns a new Launchpad Project
func NewProject(name string) *Project {
	return &Project{name: name}
}

// Tags returns a list of tags for a Launchpad project
func (p *Project) Tags() (tags []*Tag, err error) {
	// If we've already fetched the tags on this run, then return them
	if len(p.tags) > 0 {
		return p.tags, nil
	}

	// Otherwise fetch the tags from launchpad
	tags, err = p.fetchTags()
	if err != nil {
		return nil, err
	}

	return tags, nil
}

// DefaultBranch returns the default VCS branch for a Launchpad project
func (p *Project) DefaultBranch() (branch string, err error) {
	// If we've already populated the default branch, just return it
	if p.defaultBranch != "" {
		return p.defaultBranch, nil
	}

	// Make sure the Project page data is present
	if err = p.fetchProjectPage(); err != nil {
		return "", err
	}

	// Find an option element with the selected attribute.
	// Launchpad Git pages only contain one option element for selecting branch, the selected
	// branch at page load time is the default branch
	p.projectPage.Find("option[selected=selected]").Each(func(i int, s *goquery.Selection) {
		branch = s.Text()
	})

	if branch == "" {
		return "", fmt.Errorf("could not parse default branch for: %s", p.Name())
	}

	p.defaultBranch = branch
	return p.defaultBranch, nil
}

// GetNewCommits parses the git log page for a Launchpad project and returns the number of
// commits that have happened on the default branch since the last tag
func (p *Project) NewCommits() (int, error) {
	url := fmt.Sprintf("https://git.launchpad.net/%s/log", p.Name())

	doc, err := parseWebpage(url)
	if err != nil {
		return -1, err
	}

	commitTable := doc.Find("table.list")
	branchDecorationRow := commitTable.Find("a.branch-deco").First().Parent().Parent().Parent()
	tagDecorationRow := commitTable.Find("a.tag-deco").First().Parent().Parent().Parent()

	// If the decorations are on the same row, there are no commits between last tag and branch
	if tagDecorationRow.Text() == branchDecorationRow.Text() {
		return 0, nil
	}
	// Return the number of commits between the latest tag and the default branch
	return branchDecorationRow.NextUntilSelection(tagDecorationRow).Length() + 1, nil
}

// fetchTags scrapes the launchpad project repo page for a list of git tags
func (p *Project) fetchTags() (tags []*Tag, err error) {
	// Populate the project with a scrapable version of its VCS page if not already fetched
	if err = p.fetchProjectPage(); err != nil {
		return nil, err
	}

	// Get the row in the table that represents the header before the tags are listed
	tagRowHeader := p.projectPage.Find("tr.nohover").FilterFunction(func(_ int, s *goquery.Selection) bool {
		return s.Text() == "TagDownloadAuthorAge"
	})

	tagRowHeader.NextUntil("tr.nohover").EachWithBreak(func(_ int, row *goquery.Selection) bool {
		// Only get the first three tags
		if len(tags) == 3 {
			return false
		}

		// Iterate over the links, each one is either a link to a tag or a commit
		row.Find("a").Each(func(i int, l *goquery.Selection) {
			// Skip odd numbers, as these are links to commits
			if i%2 != 0 {
				return
			}

			// Grab the tag name from the only query parameter in the url
			href, exists := l.Attr("href")
			if !exists {
				return
			}

			// Assign the tag name to the value of the href param
			tagName := strings.Split(href, "=")[1]

			// We only consider tags with the name 'rev...' so bail if this isn't one
			if !strings.HasPrefix(tagName, "rev") {
				return
			}

			tag := NewTag(p.name, tagName)
			if err := tag.Process(); err != nil {
				return
			}

			tags = append(tags, tag)
		})
		return true
	})
	p.tags = tags
	return p.tags, nil
}

func (p *Project) fetchProjectPage() error {
	if p.projectPage == nil {
		projectUrl := fmt.Sprintf("https://git.launchpad.net/%s", p.Name())
		page, err := parseWebpage(projectUrl)
		if err != nil {
			return fmt.Errorf("could not fetch launchpad project page %s: %v", p.Name(), err)
		}
		p.projectPage = page
	}
	return nil
}

// parseWebpage fetches a URL and returns a goquery.Document for scraping
func parseWebpage(url string) (*goquery.Document, error) {
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
