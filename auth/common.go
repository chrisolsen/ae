package auth

import (
	"context"
	"net/http"

	"google.golang.org/appengine/urlfetch"
)

type urlGetter interface {
	Get(url string) (*http.Response, error)
}

type appEngineURLGetter struct {
	Ctx context.Context
}

func (ug appEngineURLGetter) Get(url string) (*http.Response, error) {
	client := urlfetch.Client(ug.Ctx)
	return client.Get(url)
}
