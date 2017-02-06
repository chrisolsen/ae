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
	URL *url.URL
}

// Matches wraps the putblic Matches() method
func (r *Route) Matches(pattern string) bool {
	return Matches(r.URL, pattern)
}

// Params wraps the public Params() method
func (r *Route) Params(pattern string) (map[string]string, error) {
	return Params(r.URL, pattern)
}

// Param wraps the public Param() method
func (r *Route) Param(pattern string) string {
	return Param(r.URL, pattern)
}

// Key wraps the public Key() method
func (r *Route) Key(pattern string) *datastore.Key {
	return Key(r.URL, pattern)
}

// Matches checks if the request url matches the passed in pattern. Patterns need to
// define the arguments at least one leading `:` character.
// ex.
//  /foo/:var/bar
// This method does not validate pattern argument data formats.
func Matches(url *url.URL, pattern string) bool {
	if strings.Index(pattern, ":") == -1 {
		return strings.Trim(url.Path, "/") == strings.Trim(pattern, "/")
	}

	pathParts, patternParts := slice(url.Path), slice(pattern)
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
	return true
}

// Params returns an array of string values for each of the `:` prefixed pattern parts
// GET /foo/123 - "/foo/:key" => {"key": "123"}
// GET /foo/123/bar/456 - "/foo/:parent/bar/:key" => {"parent": "123", "key": "456"}
func Params(url *url.URL, pattern string) (map[string]string, error) {
	pathParts, patternParts := slice(url.Path), slice(pattern)
	if len(pathParts) != len(patternParts) {
		return nil, ErrNoMatch
	}

	params := make(map[string]string)
	for i := 0; i < len(pathParts); i++ {
		pathPart, patternPart := pathParts[i], patternParts[i]
		if len(patternPart) == 0 {
			continue
		}
		if patternPart[0] == ':' {
			params[patternPart[1:]] = pathPart
		}
	}

	return params, nil
}

func slice(path string) []string {
	return strings.Split(strings.Trim(path, "/"), "/")
}

// Param returns string value for the `:` prefixed pattern part
// GET /foo/123 - "/foo/:key" => "123"
// GET /foo/123/bar/456 - "/foo/parent/bar/:key" => "456"
func Param(url *url.URL, pattern string) string {
	pathParts, patternParts := slice(url.Path), slice(pattern)
	if len(pathParts) != len(patternParts) {
		return ""
	}

	for i := 0; i < len(pathParts); i++ {
		if len(patternParts[i]) == 0 {
			continue
		}
		if patternParts[i][0] == ':' {
			return pathParts[i]
		}
	}
	return ""
}

// Key returns the decoded *datastore.Key value from the url
// GET /foo/agtkZXZ-YXBwbm...CAgICA4NcLDA  - "/foo/:key" => *datastore.Key
func Key(url *url.URL, pattern string) *datastore.Key {
	rawKey := Param(url, pattern)
	if len(rawKey) == 0 {
		return nil
	}
	key, err := datastore.DecodeKey(rawKey)
	if err != nil {
		return nil
	}
	return key
}
