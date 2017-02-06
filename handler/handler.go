package handler

// Contains common methods used for writing appengine apps.

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"strings"

	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
)

type handlerError struct {
	AppVersion string   `json:"appVersion"`
	URL        *url.URL `json:"url"`
	Method     string   `json:"method"`
	StatusCode int      `json:"statusCode"`
	InstanceID string   `json:"instanceId"`
	VersionID  string   `json:"versionId"`
	RequestID  string   `json:"requestId"`
	ModuleName string   `json:"moduleName"`
	Err        string   `json:"message"`
}

func (e *handlerError) Error() string {
	b, err := json.MarshalIndent(e, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(b)
}

// Base struct designed to be extended by more specific url handlers
type Base struct {
	Ctx context.Context
	Req *http.Request
	Res http.ResponseWriter
}

// OriginMiddleware returns a middleware function that validates the origin
// header within the request matches the allowed values
func OriginMiddleware(allowed []string) func(context.Context, http.ResponseWriter, *http.Request) context.Context {
	return func(c context.Context, w http.ResponseWriter, r *http.Request) context.Context {
		origin := r.Header.Get("Origin")
		if len(origin) == 0 {
			return c
		}
		ok := validateOrigin(origin, allowed)
		if !ok {
			c2, cancel := context.WithCancel(c)
			cancel()
			return c2
		}

		w.Header().Add("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Add("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS")
		w.Header().Add("Access-Control-Allow-Origin", origin)

		return c
	}
}

// ValidateOrigin is a helper method called within the ServeHTTP method on
// OPTION requests to validate the allowed origins
func (b *Base) ValidateOrigin(allowed []string) {
	origin := b.Req.Header.Get("Origin")
	ok := validateOrigin(origin, allowed)
	if !ok {
		_, cancel := context.WithCancel(b.Ctx)
		cancel()
	}
}

func validateOrigin(origin string, allowed []string) bool {
	if allowed == nil || len(allowed) == 0 {
		return true
	}
	if len(origin) == 0 {
		return false
	}
	for _, allowedOrigin := range allowed {
		if origin == allowedOrigin {
			return true
		}
	}
	return false
}

// ToJSON encodes an interface into the response writer with a default http
// status code of 200
func (b *Base) ToJSON(data interface{}) {
	err := json.NewEncoder(b.Res).Encode(data)
	if err != nil {
		b.Abort(http.StatusInternalServerError, fmt.Errorf("Decoding JSON: %v", err))
	}
}

// ToJSONWithStatus json encodes an interface into the response writer with a
// custom http status code
func (b *Base) ToJSONWithStatus(data interface{}, status int) {
	b.Res.WriteHeader(status)
	b.ToJSON(data)
}

// SendStatus writes the passed in status to the response without any data
func (b *Base) SendStatus(status int) {
	b.Res.WriteHeader(status)
}

// Bind must be called at the beginning of every request to set the required references
func (b *Base) Bind(c context.Context, w http.ResponseWriter, r *http.Request) {
	b.Ctx, b.Res, b.Req = c, w, r
}

// Header gets the request header value
func (b *Base) Header(name string) string {
	return b.Req.Header.Get(name)
}

// SetHeader sets a response header value
func (b *Base) SetHeader(name, value string) {
	b.Res.Header().Set(name, value)
}

// Abort is called when pre-maturally exiting from a handler function due to an
// error. A detailed error is delivered to the client and logged to provide the
// details required to identify the issue.
func (b *Base) Abort(statusCode int, err error) {
	c, cancel := context.WithCancel(b.Ctx)
	defer cancel()

	// testapp is the name given to all apps when being tested
	var isTest = appengine.AppID(c) == "testapp"

	hErr := &handlerError{
		URL:        b.Req.URL,
		Method:     b.Req.Method,
		StatusCode: statusCode,
		AppVersion: appengine.AppID(c),
		RequestID:  appengine.RequestID(c),
	}
	if err != nil {
		hErr.Err = err.Error()
	}

	if !isTest {
		hErr.InstanceID = appengine.InstanceID()
		hErr.VersionID = appengine.VersionID(c)
		hErr.ModuleName = appengine.ModuleName(c)
	}

	// log method to appengine log
	log.Errorf(c, hErr.Error())

	if strings.Index(b.Req.Header.Get("Accept"), "application/json") > 0 {
		b.Res.WriteHeader(statusCode)
		json.NewEncoder(b.Res).Encode(hErr)
	}
}

// Redirect is a simple wrapper around the core http method
func (b *Base) Redirect(url string, perm bool) {
	status := http.StatusTemporaryRedirect
	if perm {
		status = http.StatusMovedPermanently
	}
	http.Redirect(b.Res, b.Req, url, status)
}
