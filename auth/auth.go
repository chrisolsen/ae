package auth

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/chrisolsen/fbgraphapi"
	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
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
	token.Key, err = tstore.Get(c, key, &token)
	return &token, err
}

func Signup(c context.Context, w http.ResponseWriter, r *http.Request, creds *Credentials, account *Account) error {
	astore := NewAccountStore()
	tstore := NewTokenStore()

	accountKey, err := astore.Create(c, creds, account)
	if err != nil {
		return errors.New("failed to create account")
	}
	token, err := tstore.Create(c, accountKey)

	if accepts(r, "json") {
		setHeaderToken(w, token.Key)
	} else {
		setAuthCookieToken(w, token.Key)
	}

	return nil
}

// Signout deletes the token in the response to the client as well as deletes the token
// in the database to ensure it is no longer usable
func Signout(c context.Context, w http.ResponseWriter, r *http.Request) error {
	var err error
	var key *datastore.Key

	if accepts(r, "json") {
		key, err = getTokenKeyFromHeader(r)
		if err != nil {
			return fmt.Errorf("failed to get token from header: %v", err)
		}
		clearHeader(w)
		err = datastore.Delete(c, key)
	} else {
		key, err = getTokenKeyFromCookie(r)
		if err != nil {
			return fmt.Errorf("failed to get token from cookie: %v", err)
		}
		clearCookie(w)
		err = datastore.Delete(c, key)
	}

	return err
}

// Authorize .
func Authorize(c context.Context, w http.ResponseWriter, r *http.Request, creds *Credentials) (*Token, error) {
	var token *Token
	var err error
	if len(creds.ProviderName) > 0 {
		token, err = authorizeOut(c, creds, appEngineURLGetter{Ctx: c})
	} else {
		token, err = authorizeIn(c, creds)
	}
	if err != nil {
		return nil, err
	}

	if accepts(r, "json") {
		setHeaderToken(w, token.Key)
	} else {
		setAuthCookieToken(w, token.Key)
	}

	return token, nil
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
	})
}

func setAuthCookieToken(w http.ResponseWriter, token *datastore.Key) {
	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Expires:  time.Now().Add(time.Hour * 24 * 14),
		HttpOnly: true,
		Secure:   !appengine.IsDevAppServer(),
		Value:    token.Encode(),
	})
}

func setHeaderToken(w http.ResponseWriter, tokenKey *datastore.Key) {
	w.Header().Set("Authorization", tokenKey.Encode())
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

func accepts(r *http.Request, t string) bool {
	return strings.Index(r.Header.Get("Accept"), t) > 0
}
