package releases

import (
	"regexp"
	"strings"
)

// ciBadgeRegexp is used to find action in Github CI Badges
var ciBadgeRegexp = regexp.MustCompile(`!\[.*\]\((?P<Action>https://github.com/.+)/badge.svg\)`)

// GetCiActions tries to extract CI stages from the Badges in the README
func GetCiActions(readme string, repoName string) (actions []string) {
	// Parse the CI actions
	actionIndex := ciBadgeRegexp.SubexpIndex("Action")
	matches := ciBadgeRegexp.FindAllStringSubmatch(readme, -1)
	for _, actionMatch := range matches {
		// Check if the Action belongs to the repository
		act := actionMatch[actionIndex]
		if strings.Contains(act, repoName) {
			actions = append(actions, act)
		}
	}
	return actions
}
