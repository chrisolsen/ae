# handler

handler contains common methods used for http handler types.

## Common Usage

```
import "github.com/chrisolsen/ae/handler"
import "github.com/chrisolsen/ae/que"

var allowedOrigins []string = {
    "foo.example.com"
}

func init() {
    q := que.New(handler.OriginMiddleware(allowedOrigins))
    http.Handle("/accounts", q.Handle(accountsHandler{})
}

type accountsHandler struct {
    handler.Base
}

func (h accountsHandler) ServeHTTP(c context.Context, w http.ResponseWriter, r *http.Request) {
	h.Bind(c, w, r)

	switch r.Method {
	case http.MethodGet:
        h.handleGet()
	case http.MethodPost:
        // handle post
	case http.MethodPut:
        // handle put
	case http.MethodDelete:
        // handle delete
	case http.MethodOptions:
		h.ValidateOrigin(allowedOrigins)
	default:
		h.Abort(http.StatusNotFound, nil)
	}
}

// GET /accounts?account={key}
func (h *accountsHandler) handleGet() {
    key, ok := h.QueryKey("account")
    if !ok {
        // request is already aborted with status http.StatusBadRequest
        // and error detailing that the account querystring is required
        return
    }

    var a Account
    datastore.Get(h.Ctx, key, &a)

    // return the json encoded account with a 200OK status
    h.ToJSON(a)

    // or return json with custom status
    // h.ToJSONWithStatus(a, http.StatusTeapot)
}
```