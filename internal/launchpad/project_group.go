package launchpad

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// ProjectGroupProjectsResponse is a representation of the response given when querying the
// projects associated with a Project Group on Launchpad
type ProjectGroupProjectsResponse struct {
	Start            int            `json:"start"`
	TotalSize        int            `json:"total_size"`
	Entries          []ProjectEntry `json:"entries"`
	ResourceTypeLink string         `json:"resource_type_link"`
}

// ProjectEntry is the Launchpad API representation of a project
type ProjectEntry struct {
	Active                            bool          `json:"active"`
	ActiveMilestonesCollectionLink    string        `json:"active_milestones_collection_link"`
	AllMilestonesCollectionLink       string        `json:"all_milestones_collection_link"`
	AllSpecificationsCollectionLink   string        `json:"all_specifications_collection_link"`
	BranchSharingPolicy               string        `json:"branch_sharing_policy"`
	BrandLink                         string        `json:"brand_link"`
	BugReportedAcknowledgement        interface{}   `json:"bug_reported_acknowledgement"`
	BugReportingGuidelines            interface{}   `json:"bug_reporting_guidelines"`
	BugSharingPolicy                  string        `json:"bug_sharing_policy"`
	BugSupervisorLink                 string        `json:"bug_supervisor_link"`
	BugTrackerLink                    interface{}   `json:"bug_tracker_link"`
	CommercialSubscriptionIsDue       bool          `json:"commercial_subscription_is_due"`
	CommercialSubscriptionLink        interface{}   `json:"commercial_subscription_link"`
	DateCreated                       time.Time     `json:"date_created"`
	Description                       string        `json:"description"`
	DevelopmentFocusLink              string        `json:"development_focus_link"`
	DisplayName                       string        `json:"display_name"`
	DownloadURL                       interface{}   `json:"download_url"`
	DriverLink                        string        `json:"driver_link"`
	FreshmeatProject                  interface{}   `json:"freshmeat_project"`
	HomepageURL                       interface{}   `json:"homepage_url"`
	HTTPEtag                          string        `json:"http_etag"`
	IconLink                          string        `json:"icon_link"`
	InformationType                   string        `json:"information_type"`
	IsPermitted                       string        `json:"is_permitted"`
	LicenseApproved                   string        `json:"license_approved"`
	LicenseInfo                       interface{}   `json:"license_info"`
	Licenses                          []string      `json:"licenses"`
	LogoLink                          string        `json:"logo_link"`
	Name                              string        `json:"name"`
	OfficialAnswers                   bool          `json:"official_answers"`
	OfficialBlueprints                bool          `json:"official_blueprints"`
	OfficialBugs                      bool          `json:"official_bugs"`
	OfficialBugTags                   []interface{} `json:"official_bug_tags"`
	OfficialCodehosting               bool          `json:"official_codehosting"`
	OwnerLink                         string        `json:"owner_link"`
	Private                           bool          `json:"private"`
	ProgrammingLanguage               interface{}   `json:"programming_language"`
	ProjectGroupLink                  string        `json:"project_group_link"`
	ProjectReviewed                   string        `json:"project_reviewed"`
	QualifiesForFreeHosting           bool          `json:"qualifies_for_free_hosting"`
	RecipesCollectionLink             string        `json:"recipes_collection_link"`
	RegistrantLink                    string        `json:"registrant_link"`
	ReleasesCollectionLink            string        `json:"releases_collection_link"`
	RemoteProduct                     interface{}   `json:"remote_product"`
	ResourceTypeLink                  string        `json:"resource_type_link"`
	ReviewerWhiteboard                string        `json:"reviewer_whiteboard"`
	ScreenshotsURL                    interface{}   `json:"screenshots_url"`
	SelfLink                          string        `json:"self_link"`
	SeriesCollectionLink              string        `json:"series_collection_link"`
	SourceforgeProject                interface{}   `json:"sourceforge_project"`
	SpecificationSharingPolicy        string        `json:"specification_sharing_policy"`
	Summary                           string        `json:"summary"`
	Title                             string        `json:"title"`
	TranslationFocusLink              interface{}   `json:"translation_focus_link"`
	TranslationgroupLink              interface{}   `json:"translationgroup_link"`
	Translationpermission             string        `json:"translationpermission"`
	TranslationsUsage                 string        `json:"translations_usage"`
	ValidSpecificationsCollectionLink string        `json:"valid_specifications_collection_link"`
	Vcs                               string        `json:"vcs"`
	WebLink                           string        `json:"web_link"`
	WikiURL                           interface{}   `json:"wiki_url"`
}

// Projects lists the projects that are part of the specified project group
func EnumerateProjectGroup(projectGroup string) ([]ProjectEntry, error) {
	url := fmt.Sprintf("https://api.launchpad.net/devel/%s/projects", projectGroup)

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

	body, readErr := io.ReadAll(res.Body)
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
