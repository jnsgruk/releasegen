package launchpad

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// Project is a representation of a Launchpad Project
var (
	// errUnexpectedStatusCode is returned when an HTTP status code is not as expected.
	errUnexpectedStatusCode = errors.New("unexpected HTTP status code")
	// errFetchReadme is returned when a README could not be fetched or parsed.
	errFetchReadme = errors.New("error getting README for repo")
)

// Project is a representation of a Launchpad Project.
type Project struct {
	Name          string
	defaultBranch string
	tags          []*Tag

	projectPage *goquery.Document
}

// Tags returns a list of tags for a Launchpad project.
func (p *Project) Tags() (tags []*Tag, err error) {
	// If we've already fetched the tags on this run, then return them.
	if len(p.tags) > 0 {
		return p.tags, nil
	}

	// Otherwise fetch the tags from launchpad.
	tags, err = p.fetchTags()
	if err != nil {
		return nil, err
	}

	return tags, nil
}

// DefaultBranch returns the default VCS branch for a Launchpad project.
func (p *Project) DefaultBranch() (branch string, err error) {
	// If we've already populated the default branch, just return it.
	if p.defaultBranch != "" {
		return p.defaultBranch, nil
	}

	// Make sure the Project page data is present.
	if err = p.fetchProjectPage(); err != nil {
		return "", err
	}

	// Find an option element with the selected attribute.
	// Launchpad Git pages only contain one option element for selecting branch, the selected
	// branch at page load time is the default branch.
	p.projectPage.Find("option[selected=selected]").Each(func(i int, s *goquery.Selection) {
		branch = s.Text()
	})

	if branch == "" {
		return "", errors.New("error parsing default branch for repository")
	}

	p.defaultBranch = branch

	return p.defaultBranch, nil
}

// GetNewCommits parses the git log page for a Launchpad project and returns the number of
// commits that have happened on the default branch since the last tag.
func (p *Project) NewCommits() (int, error) {
	url := fmt.Sprintf("https://git.launchpad.net/%s/log", p.Name)

	doc, err := parseWebpage(url)
	if err != nil {
		return -1, err
	}

	commitTable := doc.Find("table.list")
	branchDecorationRow := commitTable.Find("a.branch-deco").First().Parent().Parent().Parent()
	tagDecorationRow := commitTable.Find("a.tag-deco").First().Parent().Parent().Parent()

	// If the decorations are on the same row, there are no commits between last tag and branch.
	if tagDecorationRow.Text() == branchDecorationRow.Text() {
		return 0, nil
	}

	// Return the number of commits between the latest tag and the default branch.
	return branchDecorationRow.NextUntilSelection(tagDecorationRow).Length() + 1, nil
}

// fetchTags scrapes the launchpad project repo page for a list of git tags.
func (p *Project) fetchTags() (tags []*Tag, err error) {
	// Populate the project with a scrapable version of its VCS page if not already fetched.
	if err = p.fetchProjectPage(); err != nil {
		return nil, err
	}

	// Get the row in the table that represents the header before the tags are listed.
	tagRowHeader := p.projectPage.Find("tr.nohover").FilterFunction(func(_ int, s *goquery.Selection) bool {
		return s.Text() == "TagDownloadAuthorAge"
	})

	tagRowHeader.NextUntil("tr.nohover").EachWithBreak(func(_ int, row *goquery.Selection) bool {
		// Only get the first three tags.
		if len(tags) == 3 {
			return false
		}

		// Iterate over the links, each one is either a link to a tag or a commit.
		row.Find("a").Each(func(i int, l *goquery.Selection) {
			// Skip odd numbers, as these are links to commits.
			if i%2 != 0 {
				return
			}

			// Grab the tag name from the only query parameter in the url.
			href, exists := l.Attr("href")
			if !exists {
				return
			}

			// Assign the tag name to the value of the href param.
			tagName := strings.Split(href, "=")[1]

			// We only consider tags with the name 'rev...' so bail if this isn't one.
			if !strings.HasPrefix(tagName, "rev") {
				return
			}

			tag := &Tag{project: p.Name, Name: tagName}
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

// fetchProjectPage fetches the project page, and assigns the project's projectPage to a parsed
// representation of the page.
func (p *Project) fetchProjectPage() error {
	if p.projectPage == nil {
		projectURL := fmt.Sprintf("https://git.launchpad.net/%s", p.Name)

		page, err := parseWebpage(projectURL)
		if err != nil {
			return errors.New("error fetching project page")
		}

		p.projectPage = page
	}

	return nil
}

// fetchReadmeContent fetches the content of a README.md for a project if it has one.
func (p *Project) fetchReadmeContent() (string, error) {
	url := fmt.Sprintf("https://git.launchpad.net/%s/plain/README.md", p.Name)
	res, err := http.Get(url)
	if err != nil {
		return "", errFetchReadme
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return "", errUnexpectedStatusCode
	}

	// Parse the useful information from the response.
	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return "", errFetchReadme
	}

	return string(resBody), nil
}

// Tag is a representation of a git Tag in Launchpad.
type Tag struct {
	Name      string
	Commit    string
	Timestamp *time.Time

	project string
}

// Process populates the tag with details of the relevant commit.
func (t *Tag) Process() error {
	// Construct a URL for the project commit page, including a tag if specified.
	url := fmt.Sprintf("https://git.launchpad.net/%s/commit/?h=%s", t.project, t.Name)

	doc, err := parseWebpage(url)
	if err != nil {
		return errors.New("error fetching commit page")
	}

	// Find the commit hash for the tag.
	commitTable := doc.Find("table.commit-info")
	commit := commitTable.Find("a").First().Text()
	t.Commit = commit

	// Find the timestamp for the tag/commit in question.
	ts := commitTable.Find("td.right").First().Text()

	// Parse the Launchpad timestamp into a time.Time.
	timestamp, err := time.Parse("2006-01-02 15:04:05 -0700", ts)
	if err != nil {
		return errors.New("error parsing timestamp for Launchpad tag")
	}

	t.Timestamp = &timestamp

	return nil
}

// parseWebpage fetches a URL and returns a goquery.Document for scraping.
func parseWebpage(url string) (*goquery.Document, error) {
	res, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error fetching url %s: %w", url, err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, errUnexpectedStatusCode
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, fmt.Errorf("error creating a valid document from fetched page: %w", err)
	}

	return doc, nil
}
