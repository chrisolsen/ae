package auth

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"

	"github.com/chrisolsen/fbgraphapi"
	"golang.org/x/net/context"
)

// GetToken returns the *Token value for the raw token value contained within the auth cookie or auth header
func GetToken(c context.Context, r *http.Request) (*Token, error) {
	key, err := getTokenKeyFromCookie(r)
	if err != nil && err != http.ErrNoCookie {
		return nil, err
	}
	if err == http.ErrNoCookie {
		key, err = getTokenKeyFromHeader(r)
	}
	if err != nil {
		return nil, err
	}

	tstore := NewTokenStore()
	var token Token
	err = tstore.Get(c, key, &token)
	return &token, err
}

// ClearToken deletes the token in the response to the client as well as deletes the token
// in the database to ensure it is no longer usable
func ClearToken(c context.Context, w http.ResponseWriter, r *http.Request) error {
	var err error
	var key *datastore.Key

	key, err = getTokenKeyFromCookie(r)
	if err != nil {
		clearCookie(w)
		return datastore.Delete(c, key)
	}

	key, err = getTokenKeyFromHeader(r)
	if err != nil {
		clearHeader(w)
		return datastore.Delete(c, key)
	}

	return nil
}

// Authorize .
func Authorize(c context.Context, creds *Credentials) (*Token, error) {
	if len(creds.ProviderName) > 0 {
		// Calls the private method with an appengine urlGetter.
		// This allows for internal testing of the authenticate method while
		// stubbing out the external auth service
		return authorizeOut(c, creds, appEngineURLGetter{Ctx: c})
	}
	return authorizeIn(c, creds)
}

func authorizeIn(c context.Context, creds *Credentials) (*Token, error) {
	accountStore := NewAccountStore()
	tokenStore := NewTokenStore()

	accountKey, err := accountStore.GetAccountKeyByCredentials(c, creds)
	if err != nil {
		return nil, fmt.Errorf("getting account key by credentials: %v", err)
	}

	token, err := tokenStore.Create(c, accountKey)
	if err != nil {
		return nil, err
	}

	return token, nil
}

func authorizeOut(c context.Context, creds *Credentials, urlGetter urlGetter) (*Token, error) {
	var err error
	tokenStore := NewTokenStore()
	accountStore := NewAccountStore()

	switch creds.ProviderName {
	case "facebook":
		err = fbgraphapi.Authenticate(creds.ProviderToken, creds.ProviderID, urlGetter)
	default:
		return nil, errors.New("unknown auth provider")
	}
	if err != nil {
		return nil, fmt.Errorf("authenticate: %v", err)
	}

	accountKey, err := accountStore.GetAccountKeyByCredentials(c, creds)
	if err != nil {
		return nil, fmt.Errorf("getting account key by credentials: %v", err)
	}

	token, err := tokenStore.Create(c, accountKey)
	if err != nil {
		return nil, err
	}

	return token, nil
}

func clearCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Expires:  time.Time{},
		HttpOnly: true,
		Secure:   !appengine.IsDevAppServer(),
		Value:    "o7awyeu;oiqejwuriueysdfia;lkjsd;faseufhsdjhvf", // ensure it is erased
	})
}

func getTokenKeyFromHeader(r *http.Request) (*datastore.Key, error) {
	h := r.Header.Get("Authorization")
	tokenKey, err := datastore.DecodeKey(h[len("token="):])
	if err != nil {
		return nil, fmt.Errorf("failed to decode token key: %v", err)
	}

	return tokenKey, nil
}

func getTokenKeyFromCookie(r *http.Request) (*datastore.Key, error) {
	cookie, err := r.Cookie(cookieName)
	if err != nil {
		return nil, fmt.Errorf("failed to get current cookie: %v", err)
	}

	tokenKey, err := datastore.DecodeKey(cookie.Value)
	if err != nil {
		return nil, fmt.Errorf("failed to decode token key: %v", err)
	}

	return tokenKey, nil
}

func clearHeader(w http.ResponseWriter) {
	w.Header().Del("Authorization")
}
