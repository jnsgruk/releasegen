package launchpad

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// Project represents a Launchpad Project
type Project struct {
	SelfLink                          string        `json:"self_link"`
	WebLink                           string        `json:"web_link"`
	ResourceTypeLink                  string        `json:"resource_type_link"`
	TranslationsUsage                 string        `json:"translations_usage"`
	OfficialAnswers                   bool          `json:"official_answers"`
	OfficialBlueprints                bool          `json:"official_blueprints"`
	OfficialCodehosting               bool          `json:"official_codehosting"`
	OfficialBugs                      bool          `json:"official_bugs"`
	InformationType                   string        `json:"information_type"`
	Active                            bool          `json:"active"`
	AllSpecificationsCollectionLink   string        `json:"all_specifications_collection_link"`
	ValidSpecificationsCollectionLink string        `json:"valid_specifications_collection_link"`
	BugReportingGuidelines            interface{}   `json:"bug_reporting_guidelines"`
	BugReportedAcknowledgement        interface{}   `json:"bug_reported_acknowledgement"`
	OfficialBugTags                   []interface{} `json:"official_bug_tags"`
	RecipesCollectionLink             string        `json:"recipes_collection_link"`
	BugSupervisorLink                 string        `json:"bug_supervisor_link"`
	ActiveMilestonesCollectionLink    string        `json:"active_milestones_collection_link"`
	AllMilestonesCollectionLink       string        `json:"all_milestones_collection_link"`
	TranslationgroupLink              interface{}   `json:"translationgroup_link"`
	Translationpermission             string        `json:"translationpermission"`
	QualifiesForFreeHosting           bool          `json:"qualifies_for_free_hosting"`
	ReviewerWhiteboard                string        `json:"reviewer_whiteboard"`
	IsPermitted                       string        `json:"is_permitted"`
	ProjectReviewed                   string        `json:"project_reviewed"`
	LicenseApproved                   string        `json:"license_approved"`
	Private                           bool          `json:"private"`
	DisplayName                       string        `json:"display_name"`
	IconLink                          string        `json:"icon_link"`
	LogoLink                          string        `json:"logo_link"`
	Name                              string        `json:"name"`
	OwnerLink                         string        `json:"owner_link"`
	ProjectGroupLink                  string        `json:"project_group_link"`
	Title                             string        `json:"title"`
	RegistrantLink                    string        `json:"registrant_link"`
	DriverLink                        string        `json:"driver_link"`
	Summary                           string        `json:"summary"`
	Description                       string        `json:"description"`
	DateCreated                       time.Time     `json:"date_created"`
	HomepageURL                       interface{}   `json:"homepage_url"`
	WikiURL                           interface{}   `json:"wiki_url"`
	ScreenshotsURL                    interface{}   `json:"screenshots_url"`
	DownloadURL                       interface{}   `json:"download_url"`
	ProgrammingLanguage               interface{}   `json:"programming_language"`
	SourceforgeProject                interface{}   `json:"sourceforge_project"`
	FreshmeatProject                  interface{}   `json:"freshmeat_project"`
	BrandLink                         string        `json:"brand_link"`
	BranchSharingPolicy               string        `json:"branch_sharing_policy"`
	BugSharingPolicy                  string        `json:"bug_sharing_policy"`
	SpecificationSharingPolicy        string        `json:"specification_sharing_policy"`
	Licenses                          []string      `json:"licenses"`
	LicenseInfo                       interface{}   `json:"license_info"`
	BugTrackerLink                    interface{}   `json:"bug_tracker_link"`
	SeriesCollectionLink              string        `json:"series_collection_link"`
	DevelopmentFocusLink              string        `json:"development_focus_link"`
	ReleasesCollectionLink            string        `json:"releases_collection_link"`
	TranslationFocusLink              interface{}   `json:"translation_focus_link"`
	CommercialSubscriptionLink        interface{}   `json:"commercial_subscription_link"`
	CommercialSubscriptionIsDue       bool          `json:"commercial_subscription_is_due"`
	RemoteProduct                     interface{}   `json:"remote_product"`
	Vcs                               string        `json:"vcs"`
	HTTPEtag                          string        `json:"http_etag"`

	projectPage   *goquery.Document
	defaultBranch string
	tags          []*Tag
}

type Tag struct {
	Name      string
	Commit    string
	Timestamp *time.Time
}

// Tags returns a list of tags for a Launchpad project
func (p *Project) Tags() (tags []*Tag, err error) {
	// If we've already fetched the tags on this run, then return them
	if len(p.tags) > 0 {
		return p.tags, nil
	}

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
			href := l.AttrOr("href", "")
			tagName := strings.Split(href, "=")[1]

			// We only consider tags with the name 'rev...' so bail if this isn't one
			if !strings.HasPrefix(tagName, "rev") {
				return
			}

			tag, err := p.tag(tagName)
			if err != nil {
				return
			}

			tags = append(tags, tag)
		})
		return true
	})
	p.tags = tags
	return p.tags, nil
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
		return "", fmt.Errorf("could not parse default branch for: %s", p.Name)
	}

	p.defaultBranch = branch
	return p.defaultBranch, nil
}

// GetNewCommits parses the git log page for a Launchpad project and returns the number of
// commits that have happened on the default branch since the last tag
func (p *Project) NewCommits() (int, error) {
	url := fmt.Sprintf("https://git.launchpad.net/%s/log", p.Name)

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

// tag returns information about a specified tag in a Launchpad project
func (p *Project) tag(tag string) (*Tag, error) {
	url := fmt.Sprintf("https://git.launchpad.net/%s/commit", p.Name)
	if tag != "" {
		url = fmt.Sprintf("%s/?h=%s", url, tag)
	}

	doc, err := parseWebpage(url)
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

	return &Tag{
		Name:      tag,
		Commit:    commit,
		Timestamp: &timestamp,
	}, nil
}

func (p *Project) fetchProjectPage() error {
	if p.projectPage == nil {
		projectUrl := fmt.Sprintf("https://git.launchpad.net/%s", p.Name)
		page, err := parseWebpage(projectUrl)
		if err != nil {
			return fmt.Errorf("could not fetch launchpad project page %s: %v", p.Name, err)
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
