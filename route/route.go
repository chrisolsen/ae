package route

import (
	"errors"
	"net/url"
	"strings"

	"google.golang.org/appengine/datastore"
)

// Errors
var (
	ErrNoMatch = errors.New("path and pattern don't match part count")
)

// Route type is a simple wrapper around the public methods to eliminate the need of passing
// the url to each of the methods.
type Route struct {
	URL    *url.URL
	params map[string]string
	parts  map[string]bool
}

// New creates a route
func New(url *url.URL) Route {
	return Route{URL: url}
}

// Matches checks if the request url matches the passed in pattern. Patterns need to
// define the arguments at least one leading `:` character.
// ex.
//  /foo/:var/bar
// This method does not validate pattern argument data formats.
func (r *Route) Matches(pattern string) bool {
	if strings.Index(pattern, ":") == -1 {
		return strings.Trim(r.URL.Path, "/") == strings.Trim(pattern, "/")
	}

	pathParts, patternParts := slice(r.URL.Path), slice(pattern)
	patternPartCount, pathPartCount := len(patternParts), len(pathParts)
	if pathPartCount != patternPartCount {
		return false
	}

	for i := 0; i < patternPartCount; i++ {
		pathPart, patternPart := pathParts[i], patternParts[i]

		if len(patternPart) == 0 || patternPart[0] == ':' {
			continue
		}
		if pathPart != patternPart {
			return false
		}
	}

	// extract pattern params
	params := make(map[string]string)
	for i, part := range patternParts {
		if part[0] == ':' {
			params[part[1:]] = pathParts[i]
		}
	}
	r.params = params

	// save path parts
	parts := make(map[string]bool)
	for _, val := range pathParts {
		parts[val] = true
	}
	r.parts = parts

	return true
}

// Get returns the named param from the url
func (r *Route) Get(name string) string {
	if len(name) > 0 && name[0] == ':' {
		name = name[1:]
	}
	return r.params[name]
}

// Contains indicates if the named param exists within the url
func (r *Route) Contains(val string) bool {
	return strings.Contains(r.URL.Path, val)
}

// Key wraps the public Key() method
func (r *Route) Key(name string) *datastore.Key {
	key, _ := datastore.DecodeKey(r.params[name])
	return key
}

func slice(path string) []string {
	return strings.Split(clean(path), "/")
}

func clean(val string) string {
	return strings.Trim(val, "/")
}
