package releases

import (
	"regexp"
	"strings"
)

// ciBadgeRegexp is used to find action in Github CI Badges
var ciBadgeRegexp = regexp.MustCompile(`!\[.*\]\((?P<Action>https://github.com/.+)/badge.svg\)`)

// charmBadgeRegexp is used to find a Charm's name in its CharmHub badge
var charmBadgeRegexp = regexp.MustCompile(`!\[.*\]\(https://charmhub.io/(?P<Name>.+)/badge.svg\)`)

// snapBadgeRegExp is used to find a Charm's name in its CharmHub badge
var snapBadgeRegExp = regexp.MustCompile(`!\[.*\]\(https://snapcraft.io/(?P<Name>.+)/badge.svg\)`)

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

// GetCharmName tries to parse the Charm name from a CharmHub badge in the README
func GetCharmName(readme string) (name string) {
	nameIndex := charmBadgeRegexp.SubexpIndex("Name")
	matches := charmBadgeRegexp.FindStringSubmatch(readme)
	if len(matches) > 0 {
		name = matches[nameIndex]
	}
	return name
}

// GetSnapName tries to parse the Snap name from a Snapcraft badge in the README
func GetSnapName(readme string) (name string) {
	nameIndex := snapBadgeRegExp.SubexpIndex("Name")
	matches := snapBadgeRegExp.FindStringSubmatch(readme)
	if len(matches) > 0 {
		name = matches[nameIndex]
	}
	return name
}
