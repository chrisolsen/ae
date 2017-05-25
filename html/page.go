package html

import (
	"errors"
	"net/http"

	"github.com/chrisolsen/ae/auth"
)

// Page is the container for data passed to the html form
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
func (p Page) SetUser(user interface{}, _ ...interface{}) {
	p["CurrentUser"] = user
}

// Set sets the key and value
func (p Page) Set(key string, val interface{}) {
	p[key] = val
}
