package markdown

import (
	"html/template"

	"github.com/russross/blackfriday"
)

// ToHTML converts text in the markdown syntax to html
func ToHTML(input string) interface{} {
	out := string(blackfriday.MarkdownCommon([]byte(input)))
	return template.HTML(out)
}
