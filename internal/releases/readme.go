package releases

import (
  "regexp"
  "strings"
)

// ciBadgeRegexp is used to find action in Github CI Badges
var ciBadgeRegexp = regexp.MustCompile(`!\[.*\]\((?P<Action>https://github.com/.+)/badge.svg\)`)
// charmBadgeRegexp is used to find a Charm's name in its CharmHub badge
var charmBadgeRegexp = regexp.MustCompile(`!\[.*\]\(https://charmhub.io/(?P<Name>.+)/badge.svg\)`)

// GetCiStages tries to extract CI stages from the Badges in the README
func GetCiStages(readme string, repoName string) (badges []string) {
  // Parse all the CI stages
  actionIndex := ciBadgeRegexp.SubexpIndex("Action")
  matches := ciBadgeRegexp.FindAllStringSubmatch(readme, -1)
  for _, actionMatch := range matches {
    // Check if the Badge belongs to the repository
    action := actionMatch[actionIndex]
    if strings.Contains(action, repoName) {
      badges = append(badges, action)
    }
  }
  return badges
}

// GetCharmUrl tries to parse the Charm name from a CharmHub badge in the README
func GetCharmName(readme string) (name string) {
  nameIndex := charmBadgeRegexp.SubexpIndex("Name")
  matches := charmBadgeRegexp.FindStringSubmatch(readme)
  if len(matches) > 0 {
    name = matches[nameIndex]
  }
  return name
}

