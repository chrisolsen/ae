package route

import "testing"
import "net/url"
import "reflect"

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
		if test.matches != Matches(url, test.pattern) {
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
		if test.param != Param(url, test.pattern) {
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
		args, err := Params(url, test.pattern)
		if err != test.err {
			t.Errorf("Fail: %v => %v", err, test.err)
			continue
		}

		if !reflect.DeepEqual(args, test.args) {
			t.Errorf("Fail: %s => %v", test.pattern, args)
		}
	}

}
