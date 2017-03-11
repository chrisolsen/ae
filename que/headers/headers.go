package headers

import (
	"net/http"

	"github.com/chrisolsen/ae/que"
	"golang.org/x/net/context"
)

// Set sets the response header to the key and value provided
func Set(key, value string) que.Middleware {
	return func(c context.Context, w http.ResponseWriter, r *http.Request) context.Context {
		w.Header().Set(key, value)
		return c
	}
}

// SetMulti sets the response header to the keys/values provided
func SetMulti(vals map[string]string) que.Middleware {
	return func(c context.Context, w http.ResponseWriter, r *http.Request) context.Context {
		for key, val := range vals {
			w.Header().Set(key, val)
		}
		return c
	}
}
