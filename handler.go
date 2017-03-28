package ae

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

	"github.com/chrisolsen/ae/flash"
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

// Handler struct designed to be extended by more specific url handlers
type Handler struct {
	Ctx context.Context
	Req *http.Request
	Res http.ResponseWriter

	config          HandlerConfig
	templates       map[string]*template.Template
	templateHelpers map[string]interface{}
}

// HandlerConfig contains the custom handler configuration settings
type HandlerConfig struct {
	DefaultLayout    string
	LayoutPath       string
	ViewPath         string
	ParentLayoutName string
}

var defaultHandlerConfig = HandlerConfig{
	DefaultLayout:    "application.html",
	LayoutPath:       "layouts",
	ViewPath:         "views",
	ParentLayoutName: "layout",
}

// NewHandler allows one to override the default configuration settings.
//  func NewRootHandler() rootHandler {
//  	return rootHandler{Handler: handler.New(&handler.Config{
//  		LayoutPath: "layouts/admin.html",
//  	})}
//  }
func NewHandler(c *HandlerConfig) Handler {
	b := Handler{config: *c} // copy the passed in pointer
	b.templates = make(map[string]*template.Template)
	return b
}

// DefaultHandler uses the default config settings
//  func NewRootHandler() rootHandler {
//  	return rootHandler{Handler: handler.Default()}
//  }
func DefaultHandler() Handler {
	return NewHandler(&defaultHandlerConfig)
}

// AddHelpers sets the html.template functions for the handler. This method should be
// called once to intialize the handler with a set of common template helpers used
// throughout the app.
func (h *Handler) AddHelpers(helpers map[string]interface{}) {
	dup := make(map[string]interface{})
	for k, v := range helpers {
		dup[k] = v
	}
	h.templateHelpers = dup
}

// AddHelper allows one to add additional helpers to a handler. Use this when a handler
// needs a less common helper.
func (h *Handler) AddHelper(name string, fn interface{}) {
	if h.templateHelpers == nil {
		h.templateHelpers = make(map[string]interface{})
	}
	h.templateHelpers[name] = fn
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
func (h *Handler) ValidateOrigin(allowed []string) {
	origin := h.Req.Header.Get("Origin")
	ok := validateOrigin(origin, allowed)
	if !ok {
		_, cancel := context.WithCancel(h.Ctx)
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
func (h *Handler) ToJSON(data interface{}) {
	h.Res.Header().Add("Content-Type", "application/json")
	err := json.NewEncoder(h.Res).Encode(data)
	if err != nil {
		h.Abort(http.StatusInternalServerError, fmt.Errorf("Decoding JSON: %v", err))
	}
}

// ToJSONWithStatus json encodes an interface into the response writer with a
// custom http status code
func (h *Handler) ToJSONWithStatus(data interface{}, status int) {
	h.Res.Header().Add("Content-Type", "application/json")
	h.Res.WriteHeader(status)
	h.ToJSON(data)
}

// SendStatus writes the passed in status to the response without any data
func (h *Handler) SendStatus(status int) {
	h.Res.WriteHeader(status)
}

// Bind must be called at the beginning of every request to set the required references
func (h *Handler) Bind(c context.Context, w http.ResponseWriter, r *http.Request) {
	h.Ctx, h.Res, h.Req = c, w, r
}

// Header gets the request header value
func (h *Handler) Header(name string) string {
	return h.Req.Header.Get(name)
}

// SetHeader sets a response header value
func (h *Handler) SetHeader(name, value string) {
	h.Res.Header().Set(name, value)
}

// Abort is called when pre-maturally exiting from a handler function due to an
// error. A detailed error is delivered to the client and logged to provide the
// details required to identify the issue.
func (h *Handler) Abort(statusCode int, err error) {
	c, cancel := context.WithCancel(h.Ctx)
	defer cancel()

	// testapp is the name given to all apps when being tested
	var isTest = appengine.AppID(c) == "testapp"

	hErr := &handlerError{
		URL:        h.Req.URL,
		Method:     h.Req.Method,
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

	h.Res.WriteHeader(statusCode)
	if strings.Index(h.Req.Header.Get("Accept"), "application/json") >= 0 {
		json.NewEncoder(h.Res).Encode(hErr)
	}
}

// Redirect is a simple wrapper around the core http method
func (h *Handler) Redirect(str string, args ...interface{}) {
	http.Redirect(h.Res, h.Req, fmt.Sprintf(str, args...), 303)
}

// Render pre-caches and renders template.
func (h *Handler) Render(path string, data interface{}) {
	h.RenderTemplate(path, data, RenderOptions{
		Name:    h.config.ParentLayoutName,
		FuncMap: h.templateHelpers,
		Parents: []string{filepath.Join(h.config.LayoutPath, h.config.DefaultLayout)},
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
func (h *Handler) RenderTemplate(tmplPath string, data interface{}, opts RenderOptions) {
	name := strings.TrimPrefix(tmplPath, "/")
	tmpl := h.templates[name]
	if tmpl == nil {
		t := template.New(name)
		if opts.FuncMap != nil {
			t.Funcs(opts.FuncMap)
		}
		var views []string
		if opts.Parents != nil {
			for _, p := range opts.Parents {
				views = append(views, h.fileNameWithExt(p))
			}
		} else {
			views = make([]string, 0)
		}

		views = append(views, filepath.Join(h.config.ViewPath, h.fileNameWithExt(name)))
		tmpl = template.Must(t.ParseFiles(views...))
		h.templates[name] = tmpl
	}
	if opts.Status != 0 {
		h.Res.WriteHeader(opts.Status)
	} else {
		h.Res.WriteHeader(http.StatusOK)
	}

	var renderErr error
	if opts.Name != "" {
		renderErr = tmpl.ExecuteTemplate(h.Res, opts.Name, data)
	} else {
		renderErr = tmpl.Execute(h.Res, data)
	}
	if renderErr != nil {
		panic(renderErr)
	}
}

// SetLastModified sets the Last-Modified header in the RFC1123 time format
func (h *Handler) SetLastModified(t time.Time) {
	h.Res.Header().Set("Last-Modified", t.Format(time.RFC1123))
}

// SetETag sets the etag with the md5 value
func (h *Handler) SetETag(val interface{}) {
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

	hash := md5.New()
	io.WriteString(hash, str)
	etag := base64.StdEncoding.EncodeToString(hash.Sum(nil))
	h.Res.Header().Set("ETag", etag)
}

// SetExpires sets the Expires response header with a properly formatted time value
func (h *Handler) SetExpires(t time.Time) {
	h.Res.Header().Set("Expires", t.Format(time.RFC1123))
}

// SetExpiresIn is a helper to simplify the calling of SetExpires
func (h *Handler) SetExpiresIn(d time.Duration) {
	h.Res.Header().Set("Expires", time.Now().Add(d).Format(time.RFC1123))
}

func (h *Handler) fileNameWithExt(name string) string {
	var ext string
	if strings.Index(name, ".") > 0 {
		ext = ""
	} else {
		ext = ".html"
	}
	return fmt.Sprintf("%s%s", name, ext)
}

// SetFlash sets a temporary message into a response cookie, that after
// being viewed will be removed, to prevent it from being viewed again.
func (h *Handler) SetFlash(msg string, args ...interface{}) {
	if len(args) > 0 {
		msg = fmt.Sprintf(msg, args...)
	}
	flash.Set(h.Res, msg)
}

// Flash gets the flash value
func (h *Handler) Flash() string {
	return flash.Get(h.Res, h.Req)
}
