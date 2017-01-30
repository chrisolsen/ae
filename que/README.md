# que 

que allows middleware sequenced functionality for AppEngine Go services. Url params are made available, via the `urlparams` package, by using the Go context library that is already used within AppEngine.

Since `context` is not available via the `http.Request` until Go 1.7, and AppEngine usually lags behind in regards to the version of Go made available, this library fills the hole until then.

## Example

```Go
import (
    "github.com/chrisolsen/ae/que"
)

func init() {
    r := httprouter.New()
    q := que.New(middleware1, middleware2)

    r.GET("/accounts/{id}", q.Then(accountHandler))
}

func middleware1(c context.Context, w http.ResponseWriter, r *http.Request) context.Context {
    // get context values
    foo := c.ValueOf("foo") 

    // set context param
    c2 := context.WithValue(c, "key", "some value")

    return c2
}

func accountHandler(c context.Context, w http.ResponseWriter, r *http.Request) {
    // access context values here the same as the middleware    
}
```

## License

MIT License
