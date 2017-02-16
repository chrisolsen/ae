package html

import (
	"html/template"

	"github.com/russross/blackfriday"
)

// Markdown converts text in the markdown syntax to html
func Markdown(input string) interface{} {
	out := string(blackfriday.MarkdownCommon([]byte(input)))
	return template.HTML(out)
}
