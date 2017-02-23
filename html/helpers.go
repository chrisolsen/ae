package html

import (
	"html/template"

	"github.com/russross/blackfriday"
)

// CSS prevents any custom embedded styles from being encoded to html safe values
func CSS(s string) template.CSS {
	return template.CSS(s)
}

// Preview returns a preview of the string
func Preview(size int, val string) string {
	if len(val) <= size {
		return val
	}
	return val[:size-1] + "..."
}

// Markdown converts text in the markdown syntax to html
func Markdown(input string) interface{} {
	out := string(blackfriday.MarkdownCommon([]byte(input)))
	return template.HTML(out)
}
