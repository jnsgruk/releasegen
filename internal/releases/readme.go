package releases

import (
  "regexp"
)

// ciBadgeRegex is used to find Github CI Badges
var ciBadgeRegexp = regexp.MustCompile(`!\[.*\]\((?P<Action>https://github.com/.+)/badge.svg\)`)
var charmBadgeRegexp = regexp.MustCompile(`!\[.*\]\(https://charmhub.io/(?P<Name>.+)/badge.svg\)`)

// GetCiStages tries to extract CI stages from the Badges in the README
func GetCiStages(readme string) (badges []string) {
  // Parse all the CI stages
  actionIndex := ciBadgeRegexp.SubexpIndex("Action")
  matches := ciBadgeRegexp.FindAllStringSubmatch(readme, -1)
  for index, _ := range matches {
    badges = append(badges, matches[index][actionIndex])
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

