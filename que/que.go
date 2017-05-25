package que

import (
	"net/http"

	"golang.org/x/net/context"
	"google.golang.org/appengine"
)

// Middleware is a http.HandlerFunc that also includes a context and url params variables
type Middleware func(context.Context, http.ResponseWriter, *http.Request) context.Context

// HandlerFunc much like the standard http.HandlerFunc, but includes the request context
type HandlerFunc func(context.Context, http.ResponseWriter, *http.Request)

// Handler much like the standard http.Handler, but includes the request context
// in the ServeHTTP method
type Handler interface {
	ServeHTTP(context.Context, http.ResponseWriter, *http.Request)
}

// Q allows a list middleware functions to be created and run
type Q struct {
	ops     []Middleware
	handler Handler
}

// New initializes the middleware chain with one or more handler functions.
// The returned pointer allows for additional middleware methods to be added or
// for the chain to be run.
//	q := que.New(foo, bar)
func New(ops ...Middleware) *Q {
	q := Q{}
	q.ops = ops
	return &q
}

// Add allows for one or more middleware handler functions to be added to the
// existing chain
//	q := que.New(cors, format)
//	q.Add(auth)
func (q *Q) Add(ops ...Middleware) {
	q.ops = append(q.ops, ops...)
}

// Run executes the handler chain, which is most useful in tests
//	q := que.New(foo, bar)
// 	q.Add(func(c context.Context, w http.ResponseWriter, r *http.Request) {
// 		// perform tests here
// 	})
//  inst := aetest.NewInstance(nil)
// 	r := inst.NewRequest("GET", "/", nil)
// 	w := httpTest.NewRecorder()
// 	c := appengine.NewContext(r)
// 	q.Run(c, w, r)
func (q *Q) Run(c context.Context, w http.ResponseWriter, r *http.Request) {
	for _, op := range q.ops {
		c = op(c, w, r)
		if c.Err() != nil {
			return
		}
	}
}

// HandleFunc returns the chain of existing middleware that includes the final HandlerFunc argument.
//	q := que.New(foo, bar)
//  router.Get("/", q.HandleFunc(handleRoot))
func (q *Q) HandleFunc(fn HandlerFunc) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		c := appengine.NewContext(r)
		for _, op := range q.ops {
			c = op(c, w, r)
			if c.Err() != nil {
				return
			}
		}
		fn(c, w, r)
	}
}

// Handle accepts a Handler interface and returns the chain of existing middleware
// that includes the final Handler argument.
//	q := que.New(foo, bar)
//  router.Get("/", q.Handle(handleRoot))
func (q *Q) Handle(h Handler) http.Handler {
	return handler{ops: q.ops, handler: h}
}

// handler allows the middleware calls to be wrapped up into a Handler interface
type handler struct {
	ops     []Middleware
	handler Handler
}

func (h handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	for _, op := range h.ops {
		c = op(c, w, r)
		if c.Err() != nil {
			return
		}
	}
	h.handler.ServeHTTP(c, w, r)
}
