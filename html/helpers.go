package html

import (
	"html/template"
	"time"

	"github.com/chrisolsen/ae/model"
	"github.com/russross/blackfriday"
	"google.golang.org/appengine/datastore"
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

// Add adds the numbers
func Add(a, b int) int {
	return a + b
}

// EncodeKey encodes a datastore key
func EncodeKey(data interface{}) string {
	switch data.(type) {
	case model.Base:
		return data.(model.Base).Key.Encode()
	case *datastore.Key:
		return data.(*datastore.Key).Encode()
	default:
		return ""
	}
}

// Checked returns the checked attribute for positive values.
// 	<input type="checkbox" {{IsAdmin | checked}}> => <input type="checkbox" checked="checked">
func Checked(checked bool) template.HTMLAttr {
	if checked {
		return template.HTMLAttr(`checked="checked"`)
	}
	return ""
}

// Disabled returns the checked attribute for positive values.
// 	<button type="submit" {{HasError | disabled}}>Save</button> => <button type="submit" disabled="">Save</button>
// 	or
// 	<button type="submit" {{ValidationError | disabled}}>Save</button> => <button type="submit" disabled="">Save</button>
func Disabled(err interface{}) template.HTMLAttr {
	d := template.HTMLAttr(`disabled="disabled"`)
	switch err.(type) {
	case string:
		if len(err.(string)) == 0 {
			return ""
		}
		return d
	case error:
		return d
	default:
		return ""
	}
}

// KeyEqual allow for *datastore.Key comparison
func KeyEqual(a, b *datastore.Key) bool {
	return a.Equal(b)
}

// Timestamp formats the time to the RFC3339 layout
func Timestamp(d time.Time) string {
	return d.Format(time.RFC3339)
}
