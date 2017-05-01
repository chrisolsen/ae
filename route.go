package ae

import (
	"errors"
	"net/http"
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
	req    *http.Request
	params map[string]string
	parts  map[string]bool
}

// NewRoute creates a route
func NewRoute(r *http.Request) Route {
	return Route{req: r}
}

// MatchesPath checks for path matches and allows for wildcards
func (r *Route) MatchesPath(pattern string) bool {
	url := r.req.URL
	wildcard := strings.Contains(pattern, "*")
	if strings.Index(pattern, ":") == -1 && !wildcard {
		return strings.Trim(url.Path, "/") == strings.Trim(pattern, "/")
	}

	pathParts, patternParts := slicePath(url.Path), slicePath(pattern)

	var partCount int
	if wildcard {
		partCount = len(patternParts) - 1
		if len(pathParts) < partCount {
			return false
		}
	} else {
		patternPartCount, pathPartCount := len(patternParts), len(pathParts)
		if pathPartCount != patternPartCount {
			return false
		}
		partCount = patternPartCount
	}

	for i := 0; i < partCount; i++ {
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

// Matches checks if the request url matches the passed in pattern. Patterns need to
// define the arguments at least one leading `:` character.
// ex.
//  /foo/:var/bar
// This method does not validate pattern argument data formats.
func (r *Route) Matches(method, pattern string) bool {
	if r.req.Method != strings.ToUpper(method) {
		return false
	}
	return r.MatchesPath(pattern)
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
	return strings.Contains(r.req.URL.Path, val)
}

// Key wraps the public Key() method
func (r *Route) Key(name string) *datastore.Key {
	key, _ := datastore.DecodeKey(r.params[name])
	return key
}

func slicePath(path string) []string {
	return strings.Split(strings.Trim(path, "/"), "/")
}
