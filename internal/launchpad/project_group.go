package launchpad

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
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
}

// ProjectGroupProjectsResponse is a representation of the response given when querying the
// projects associated with a Project Group on Launchpad
type ProjectGroupProjectsResponse struct {
	Start            int       `json:"start"`
	TotalSize        int       `json:"total_size"`
	Entries          []Project `json:"entries"`
	ResourceTypeLink string    `json:"resource_type_link"`
}

// EnumerateProjectGroup lists the projects that are part of the specified project group
func EnumerateProjectGroup(name string) ([]Project, error) {
	url := fmt.Sprintf("https://api.staging.launchpad.net/devel/%s/projects", name)
	log.Printf("processing launchpad project group: %s", url)

	// TODO: Add a retry here?
	client := http.Client{Timeout: time.Second * 5}

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
