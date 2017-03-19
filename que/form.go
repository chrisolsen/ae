package que

import (
	"net/http"
	"strings"

	"golang.org/x/net/context"
)

func SetMethod(c context.Context, w http.ResponseWriter, r *http.Request) context.Context {
	if strings.Contains(r.Header.Get("Content-Type"), "multipart/form-data") {
		r.ParseMultipartForm(10 << 20)
	} else {
		r.ParseForm()
	}
	if method := r.FormValue("_method"); method != "" {
		r.Method = strings.ToUpper(method)
	}
	return c
}
