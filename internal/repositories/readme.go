package repositories

import (
	"regexp"
	"strings"

	"github.com/gomarkdown/markdown"
)

// ciBadgeRegexp is used to find action in Github CI Badges
var ciBadgeRegexp = regexp.MustCompile(`(?P<Action>https://github.com/[\w-./]+)/badge.svg`)

// prRegexp is used to find Github PR URLs
var prRegexp = regexp.MustCompile(`(https://github.com/canonical/.+/pull/([0-9]+))`)

// userRegexp is used to find Twitter/Github style mentions such as '@JoeBloggs'
var userRegexp = regexp.MustCompile(`(\A|\s)@([\w-_]+)`)

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

// RenderReleaseBody transforms a Markdown string into HTML
func RenderReleaseBody(body string) string {
	// Preprocess any Pull Request links in the Markdown body
	body = prRegexp.ReplaceAllString(body, `<a href="${1}">#${2}</a>`)
	// Preprocess any user mentions in the Markdown body
	body = userRegexp.ReplaceAllString(body, `${1}<a href="http://github.com/${2}">@${2}</a>`)

	// Render the Markdown to HTML
	md := []byte(body)
	normalised := markdown.NormalizeNewlines(md)
	return string(markdown.ToHTML(normalised, nil, nil))
}
