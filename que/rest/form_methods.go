package rest

import (
	"net/http"
	"strings"

	"golang.org/x/net/context"
)

// SetMethod is a middleware function that changes the request method of them form if it includes a
// `_method` value with submitted form values.
func SetMethod(c context.Context, w http.ResponseWriter, r *http.Request) context.Context {
	r.ParseForm()
	action := r.FormValue("_method")

	switch action {
	case http.MethodDelete, http.MethodHead, http.MethodPatch, http.MethodPost, http.MethodPut:
		r.Method = strings.ToUpper(action)
	}

	return c
}
