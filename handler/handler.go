package handler

// Contains common methods used for writing appengine apps.

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"time"

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

	config          Config
	templates       map[string]*template.Template
	templateHelpers map[string]interface{}
}

// Config contains the custom handler configuration settings
type Config struct {
	DefaultLayout    string
	LayoutPath       string
	ViewPath         string
	ParentLayoutName string
}

var defaultConfig = Config{
	DefaultLayout:    "application.html",
	LayoutPath:       "layouts",
	ViewPath:         "views",
	ParentLayoutName: "layout",
}

// New allows one to override the default configuration settings.
//  func NewRootHandler() rootHandler {
//  	return rootHandler{Base: handler.New(&handler.Config{
//  		LayoutPath: "layouts/admin.html",
//  	})}
//  }
func New(c *Config) Base {
	b := Base{config: *c} // copy the passed in pointer
	b.templates = make(map[string]*template.Template)
	return b
}

// Default uses the default config settings
//  func NewRootHandler() rootHandler {
//  	return rootHandler{Base: handler.Default()}
//  }
func Default() Base {
	return New(&defaultConfig)
}

// AddHelpers sets the html.template functions for the handler. This method should be
// called once to intialize the handler with a set of common template helpers used
// throughout the app.
func (b *Base) AddHelpers(helpers map[string]interface{}) {
	dup := make(map[string]interface{})
	for k, v := range helpers {
		dup[k] = v
	}
	b.templateHelpers = dup
}

// AddHelper allows one to add additional helpers to a handler. Use this when a handler
// needs a less common helper.
func (b *Base) AddHelper(name string, fn interface{}) {
	if b.templateHelpers == nil {
		b.templateHelpers = make(map[string]interface{})
	}
	b.templateHelpers[name] = fn
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
	b.Res.Header().Add("Content-Type", "application/json")
	err := json.NewEncoder(b.Res).Encode(data)
	if err != nil {
		b.Abort(http.StatusInternalServerError, fmt.Errorf("Decoding JSON: %v", err))
	}
}

// ToJSONWithStatus json encodes an interface into the response writer with a
// custom http status code
func (b *Base) ToJSONWithStatus(data interface{}, status int) {
	b.Res.Header().Add("Content-Type", "application/json")
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

	b.Res.WriteHeader(statusCode)
	if strings.Index(b.Req.Header.Get("Accept"), "application/json") >= 0 {
		json.NewEncoder(b.Res).Encode(hErr)
	}
}

// Redirect is a simple wrapper around the core http method
func (b *Base) Redirect(str string, args ...interface{}) {
	http.Redirect(b.Res, b.Req, fmt.Sprintf(str, args...), 303)
}

// Render pre-caches and renders template.
func (b *Base) Render(path string, data interface{}) {
	b.RenderTemplate(path, data, RenderOptions{
		Name:    b.config.ParentLayoutName,
		FuncMap: b.templateHelpers,
		Parents: []string{filepath.Join(b.config.LayoutPath, b.config.DefaultLayout)},
	})
}

// RenderOptions contain the optional data items for rendering
type RenderOptions struct {
	// http status to return in the response
	Status int

	// template functions
	FuncMap template.FuncMap

	// parent layout paths to render the defined view within
	Parents []string

	// the defined *name* to render
	// 	{{define "layout"}}...{{end}}
	Name string
}

// RenderTemplate renders the template without any layout
func (b *Base) RenderTemplate(tmplPath string, data interface{}, opts RenderOptions) {
	name := strings.TrimPrefix(tmplPath, "/")
	tmpl := b.templates[name]
	if tmpl == nil {
		t := template.New(name)
		if opts.FuncMap != nil {
			t.Funcs(opts.FuncMap)
		}
		var views []string
		if opts.Parents != nil {
			for _, p := range opts.Parents {
				views = append(views, b.fileNameWithExt(p))
			}
		} else {
			views = make([]string, 0)
		}

		views = append(views, filepath.Join(b.config.ViewPath, b.fileNameWithExt(name)))
		tmpl = template.Must(t.ParseFiles(views...))
		b.templates[name] = tmpl
	}
	if opts.Status != 0 {
		b.Res.WriteHeader(opts.Status)
	} else {
		b.Res.WriteHeader(http.StatusOK)
	}

	var renderErr error
	if opts.Name != "" {
		renderErr = tmpl.ExecuteTemplate(b.Res, opts.Name, data)
	} else {
		renderErr = tmpl.Execute(b.Res, data)
	}
	if renderErr != nil {
		panic(renderErr)
	}
}

// SetLastModified sets the Last-Modified header in the RFC1123 time format
func (b *Base) SetLastModified(t time.Time) {
	b.Res.Header().Set("Last-Modified", t.Format(time.RFC1123))
}

// SetETag sets the etag with the md5 value
func (b *Base) SetETag(val interface{}) {
	var str string
	switch val.(type) {
	case string:
		str = val.(string)
	case time.Time:
		str = val.(time.Time).Format(time.RFC1123)
	case fmt.Stringer:
		str = val.(fmt.Stringer).String()
	default:
		str = fmt.Sprintf("%v", val)
	}

	h := md5.New()
	io.WriteString(h, str)
	etag := base64.StdEncoding.EncodeToString(h.Sum(nil))
	b.Res.Header().Set("ETag", etag)
}

func (b *Base) SetExpires(t time.Time) {
	b.Res.Header().Set("Expires", t.Format(time.RFC1123))
}

func (b *Base) SetExpiresIn(d time.Duration) {
	b.Res.Header().Set("Expires", time.Now().Add(d).Format(time.RFC1123))
}

func (b *Base) fileNameWithExt(name string) string {
	var ext string
	if strings.Index(name, ".") > 0 {
		ext = ""
	} else {
		ext = ".html"
	}
	return fmt.Sprintf("%s%s", name, ext)
}
