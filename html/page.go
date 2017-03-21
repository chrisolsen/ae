package html

import (
	"errors"
	"math"
	"net/http"

	"github.com/chrisolsen/ae/auth"
)

type Page map[string]interface{}

// NewPage creates a new page
func NewPage() Page {
	return Page(make(map[string]interface{}))
}

// NewPageWithCSRFToken create a page with an initialized CSRF token
func NewPageWithCSRFToken(r *http.Request) Page {
	p := NewPage()
	token := auth.NewCSRFToken(r)
	p["CSRFToken"] = token
	return p
}

// SetError sets any error that needs to be shown
func (p Page) SetError(err interface{}) {
	switch err.(type) {
	case string:
		p["Error"] = errors.New(err.(string))
	case error:
		p["Error"] = err.(error)
	default:
		p["Error"] = nil
	}
}

// SetUser sets the current user
func (p Page) SetUser(user interface{}) {
	p["CurrentUser"] = user
}

// Set sets the key and value
func (p Page) Set(key string, val interface{}) {
	p[key] = val
}

// SetPageOffsets sets an array within the page to allow it to be iterated through
// in the template to create pagination links.
//  // .go file
// 	p := html.NewPage()
// 	items := []string {"foo", "bar", "bits", ...}
// 	p.SetPageOffsets(len(items), 10)
//
//  // template
//  {{range $index, $offset := .Offsets}}
//      <a href="/name?o={{$offset}}">{{add $index 1}}</a>
//  {{end}}
func (p Page) SetPageOffsets(itemCount, pageSize int) {
	offsetCount := int(math.Ceil(float64(itemCount) / float64(pageSize)))
	offsets := make([]int, offsetCount)
	for i := 0; i < offsetCount; i++ {
		offsets[i] = i * 10
	}
	p.Set("Offsets", offsets)
}
