package auth

import (
	"net/http"

	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/urlfetch"
)

type urlGetter interface {
	Get(url string) (*http.Response, error)
}

type appEngineURLGetter struct {
	Ctx context.Context
}

func (ug appEngineURLGetter) Get(url string) (*http.Response, error) {
	client := &http.Client{Transport: &urlfetch.Transport{Context: ug.Ctx, AllowInvalidServerCertificate: appengine.IsDevAppServer()}}
	return client.Get(url)
}
