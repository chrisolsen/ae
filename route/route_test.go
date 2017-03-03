package route

import (
	"fmt"
	"testing"

	"net/http"

	"github.com/chrisolsen/ae/testutils"
	"google.golang.org/appengine/datastore"
)

func TestRouteContains(t *testing.T) {
	type test struct {
		url      string
		val      string
		expected bool
	}

	tests := []test{
		test{url: "/foo/bar", val: "bar", expected: true},
		test{url: "/foo/bar/blah", val: "bar", expected: true},
		test{url: "http://blah.com/foo/bar?blah=bar#blah", val: "blah", expected: false},
	}

	for _, test := range tests {
		req, _ := http.NewRequest("GET", test.url, nil)
		r := Route{req: req}
		if r.Contains(test.val) != test.expected {
			t.Errorf("expected: %v got %v", test.expected, test.val)
		}
	}
}

func TestRouteKey(t *testing.T) {
	type test struct {
		url     string
		pattern string
		key     *datastore.Key
	}

	var T = testutils.T{}
	defer T.Close()
	c := T.GetContext()

	validKey := datastore.NewIncompleteKey(c, "users", nil)

	tests := []test{
		test{
			url:     fmt.Sprintf("/foo/%s/bar", validKey.Encode()),
			pattern: "/foo/:key/bar",
			key:     validKey,
		},
		test{
			url:     fmt.Sprintf("/foo/%s", validKey.Encode()),
			pattern: "/foo/:key",
			key:     validKey,
		},
		test{
			url:     fmt.Sprintf("/%s/bar", validKey.Encode()),
			pattern: "/:key/bar",
			key:     validKey,
		},
		test{
			url:     fmt.Sprintf("/%s", validKey.Encode()),
			pattern: "/:key",
			key:     validKey,
		},
		test{
			url:     "/foo/some_invalid_value/bar",
			pattern: "/foo/:key/bar",
			key:     nil,
		},
	}

	for _, test := range tests {
		req, _ := http.NewRequest("GET", test.url, nil)
		r := Route{req: req}
		if !r.Matches("GET", test.pattern) {
			t.Errorf("failed to extract key from: %v with %v", test.url, test.pattern)
			continue
		}
		key := r.Key("key")
		if !key.Equal(test.key) {
			t.Errorf("keys don't match: %v <=> %v", key, test.key)
			continue
		}
	}
}

func TestRoutesMatch(t *testing.T) {
	type test struct {
		pattern string
		path    string
		matches bool
	}

	tests := []test{
		test{"/foo/:parentKey/bar/:key", "/foo/123/bar/456", true},
		test{"/bar/:parentKey/foo/:key", "/foo/123/bar/456", false},
		test{"/foo/:parentKey", "/foo/123", true},
		test{"/foo/:parentKey/", "/foo/123", true},
		test{"/foo/:parentKey//", "/foo/123", true},
		test{"/:param", "/123", true},
		test{"/", "/", true},
		test{"/", "", true},
		test{"/:foo", "", true}, // blank param
		test{"/foo/:parentKey/bar/", "/foo/123", false},
		test{"/", "/foo", false},
	}

	for _, test := range tests {
		req, _ := http.NewRequest("GET", test.path, nil)
		r := Route{req: req}
		if test.matches != r.Matches("GET", test.pattern) {
			t.Errorf("Fail: %s <=> %s", test.pattern, test.path)
		}
	}
}

func TestRouteParams(t *testing.T) {
	type test struct {
		pattern string
		path    string
		name    string
		value   string
	}

	tests := []test{
		test{"/foo/:parentKey/bar/:key", "/foo/123/bar/456", "parentKey", "123"},
		test{"/foo/:parentKey/bar/:key", "/foo/123/bar/456", "key", "456"},
		test{"/foo/:parentKey", "/foo/123", "parentKey", "123"},
		test{"/foo/:parentKey/", "/foo/123", "parentKey", "123"},
		test{"/foo/:parentKey//", "/foo/123", "parentKey", "123"},
		test{"/:param", "/123", "param", "123"},
		test{"/", "/", "", ""},
		test{"/", "", "", ""},
		test{"/:foo", "", "", ""}, // blank param
	}

	for _, test := range tests {
		req, _ := http.NewRequest("GET", test.path, nil)
		r := Route{req: req}
		if !r.Matches("GET", test.pattern) {
			t.Error("route failed to match")
			continue
		}
		if test.value != r.Get(test.name) {
			t.Errorf("Fail: %s <=> %s", test.pattern, test.path)
		}
	}
}

func TestRouteMethodMatch(t *testing.T) {
	type test struct {
		methodType string
		path       string
		matchType  string
		match      bool
	}

	tests := []test{
		test{"GET", "/foo", "GET", true},
		test{"GET", "/foo", "POST", false},
		test{"GET", "/foo", "PATCH", false},
		test{"GET", "/foo", "PUT", false},
		test{"GET", "/foo", "DELETE", false},
		test{"POST", "/foo", "POST", true},
		test{"POST", "/foo", "PATCH", false},
		test{"POST", "/foo", "PUT", false},
		test{"POST", "/foo", "DELETE", false},
		test{"PUT", "/foo", "PUT", true},
		test{"PUT", "/foo", "PATCH", false},
		test{"PUT", "/foo", "DELETE", false},
		test{"DELETE", "/foo", "DELETE", true},
		test{"DELETE", "/foo", "PATCH", false},
	}

	for _, test := range tests {
		req, _ := http.NewRequest(test.methodType, test.path, nil)
		r := Route{req: req}

		if test.match && !r.Matches(test.matchType, test.path) {
			t.Errorf("Fail: %s <=> %s", test.methodType, test.matchType)
		}
	}
}
