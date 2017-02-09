package route

import (
	"fmt"
	"net/url"
	"reflect"
	"testing"

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
		url, _ := url.Parse(test.url)
		r := Route{URL: url}
		if r.Contains(test.val) != test.expected {
			t.Errorf("expected: %v got %v", test.url, url)
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
		test{
			url:     "/foo/some_invalid_value/bar",
			pattern: "/foo/bar",
			key:     nil,
		},
	}

	for _, test := range tests {
		url, _ := url.Parse(test.url)
		r := Route{URL: url}
		key := r.Key(test.pattern)
		if key == nil && test.key != nil {
			t.Errorf("failed to extract key from: %v with %v", url, test.pattern)
			continue
		}
		if !key.Equal(test.key) {
			t.Errorf("keys don't match: %v <=> %v", url, test.url)
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
		url, _ := url.Parse(test.path)
		r := Route{URL: url}
		if test.matches != r.Matches(test.pattern) {
			t.Errorf("Fail: %s <=> %s", test.pattern, test.path)
		}
	}

}

func TestRoutesParam(t *testing.T) {
	type test struct {
		pattern string
		path    string
		param   string
	}

	tests := []test{
		test{"/foo/:parentKey/bar/key", "/foo/123/bar/456", "123"},
		test{"/bar/parentKey/foo/:key", "/foo/123/bar/456", "456"},
		test{"/foo/:parentKey", "/foo/123", "123"},
		test{"/foo/:parentKey/", "/foo/123", "123"},
		test{"/foo/:parentKey//", "/foo/123", "123"},
		test{"/:param", "/123", "123"},
		test{"/", "/", ""},
		test{"/", "", ""},
		test{"/:foo", "", ""}, // blank param
		test{"/foo/:parentKey/bar/", "/foo/123", ""},
		test{"/", "/foo", ""},
	}

	for _, test := range tests {
		url, _ := url.Parse(test.path)
		r := Route{URL: url}
		if test.param != r.Param(test.pattern) {
			t.Errorf("Fail: %s <=> %s", test.pattern, test.path)
		}
	}
}

func TestRoutesParams(t *testing.T) {

	type test struct {
		pattern string
		path    string
		args    map[string]string
		err     error
	}

	tests := []test{
		test{
			pattern: "/foo/:parentKey/bar/:key",
			path:    "/foo/123/bar/456",
			args: map[string]string{
				"parentKey": "123",
				"key":       "456",
			},
			err: nil,
		},
		test{
			pattern: "/foo/:parentKey/bar/:key",
			path:    "/foo//bar/456",
			args: map[string]string{
				"parentKey": "",
				"key":       "456",
			},
			err: nil,
		},

		test{
			pattern: "/foo/:parentKey/bar/:key",
			path:    "/foo/123/bar/",
			err:     ErrNoMatch,
		},
		test{
			pattern: "/foo/:parentKey/bar",
			path:    "/foo/123/bar/456",
			err:     ErrNoMatch,
		},
		test{
			pattern: "/foo/:parentKey/bar/",
			path:    "/foo/123/bar/456",
			err:     ErrNoMatch,
		},
	}

	for _, test := range tests {
		url, _ := url.Parse(test.path)
		r := Route{URL: url}
		args, err := r.Params(test.pattern)
		if err != test.err {
			t.Errorf("Fail: %v => %v", err, test.err)
			continue
		}

		if !reflect.DeepEqual(args, test.args) {
			t.Errorf("Fail: %s => %v", test.pattern, args)
		}
	}

}
