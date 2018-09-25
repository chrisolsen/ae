package html

import (
	"errors"
	"net/http"
	"sync"

	"github.com/chrisolsen/ae/auth"
)

// Page is the container for data passed to the html form
type Page struct {
	data map[string]interface{}
	m    sync.RWMutex
}

// type Page map[string]interface{}

// NewPage creates a new page
func NewPage() *Page {
	return &Page{
		data: make(map[string]interface{}),
		m:    sync.RWMutex{},
	}
}

// NewPageWithCSRFToken create a page with an initialized CSRF token
func NewPageWithCSRFToken(r *http.Request) *Page {
	p := NewPage()
	token := auth.NewCSRFToken(r)
	p.Set("CSRFToken", token)
	return p
}

// SetError sets any error that needs to be shown
func (p *Page) SetError(err interface{}) {
	switch err.(type) {
	case string:
		p.Set("Error", errors.New(err.(string)))
	case error:
		p.Set("Error", err.(error))
	default:
		p.Set("Error", nil)
	}
}

// SetUser sets the current user
func (p *Page) SetUser(user interface{}, _ ...interface{}) {
	p.Set("CurrentUser", user)
}

// Set sets the key and value
func (p *Page) Set(key string, val interface{}) {
	p.m.Lock()
	p.data[key] = val
	p.m.Unlock()
}

func (p *Page) Values() map[string]interface{} {
	return p.data
}
