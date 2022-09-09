package md

import (
	"regexp"

	"github.com/gomarkdown/markdown"
)

// prRegexp is used to find Github PR URLs
var prRegexp = regexp.MustCompile(`(https://github.com/canonical/.+/pull/([0-9]+))`)

// userRegexp is used to find Twitter/Github style mentions such as '@JoeBloggs'
var userRegexp = regexp.MustCompile(`(\A|\s)@(\w+)`)

// RenderReleaseBody transforms a Markdown string into HTML
func RenderReleaseBody(body string) string {
	// Preprocess any Pull Request links in the Markdown body
	body = prRegexp.ReplaceAllString(body, `<a href="${1}">#${2}</a>`)
	// Preprocess any user mentions in the Markdown body
	body = userRegexp.ReplaceAllString(body, `${1}<a href="http://github.com/${2}">@${2}</a>`)

	// Render the Markdown to HTML
	md := []byte(body)
	return string(markdown.ToHTML(md, nil, nil))
}
