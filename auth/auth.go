package auth

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"net/url"

	"github.com/chrisolsen/fbgraphapi"
	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
)

var ErrNoCookie = errors.New("no cookie found")

// GetToken returns the *Token value for the raw token value contained within the auth cookie or auth header
func GetToken(c context.Context, r *http.Request) (*Token, error) {
	uuid, err := getUUIDFromCookie(r)
	if err != nil && err != http.ErrNoCookie {
		return nil, err
	}
	if err == http.ErrNoCookie {
		uuid = getUUIDFromHeader(r)
	}
	if err != nil {
		return nil, err
	}

	tstore := NewTokenStore()
	return tstore.Get(c, uuid)
}

// Signup creates a user account and links up the credentials. Based on the request type an auth cookie
// or header token will be set with an auth token.
func Signup(c context.Context, w http.ResponseWriter, r *http.Request, creds *Credentials, account *Account) error {
	astore := NewAccountStore()
	tstore := NewTokenStore()

	accountKey, err := astore.Create(c, creds, account)
	if err != nil {
		return errors.New("failed to create account")
	}
	token, err := tstore.Create(c, accountKey)

	if accepts(r, "json") {
		setHeaderToken(w, token.UUID)
	} else {
		setAuthCookieToken(w, token.UUID)
	}

	return nil
}

// Signout deletes the token in the response to the client as well as deletes the token
// in the database to ensure it is no longer usable
func Signout(c context.Context, w http.ResponseWriter, r *http.Request) error {
	var err error
	var uuid string

	if accepts(r, "json") {
		uuid = getUUIDFromHeader(r)
		if err != nil {
			return fmt.Errorf("failed to get token from header: %v", err)
		}
		clearHeader(w)
	} else {
		uuid, err = getUUIDFromCookie(r)
		if err != nil {
			return fmt.Errorf("failed to get token from cookie: %v", err)
		}
		clearCookie(w)
	}

	store := NewTokenStore()
	token, err := store.Get(c, uuid)
	if err != nil {
		return err
	}
	err = datastore.Delete(c, token.Key)

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
		setHeaderToken(w, token.UUID)
	} else {
		setAuthCookieToken(w, token.UUID)
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
		Path:     "/",
		Expires:  time.Time{},
		HttpOnly: true,
		Value:    "",
		Secure:   !appengine.IsDevAppServer(),
	})
}

func setAuthCookieToken(w http.ResponseWriter, uuid string) {
	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Path:     "/",
		Expires:  time.Now().Add(time.Hour * 24 * 14),
		HttpOnly: true,
		Secure:   !appengine.IsDevAppServer(),
		Value:    uuid,
	})
}

func setHeaderToken(w http.ResponseWriter, uuid string) {
	w.Header().Set("Authorization", uuid)
}

func getUUIDFromHeader(r *http.Request) string {
	return r.Header.Get("Authorization")[len("token="):]
}

func getUUIDFromCookie(r *http.Request) (string, error) {
	cookie, err := r.Cookie(cookieName)
	if err != nil || len(cookie.Value) == 0 {
		return "", ErrNoCookie
	}

	return cookie.Value, nil
}

func clearHeader(w http.ResponseWriter) {
	w.Header().Del("Authorization")
}

func accepts(r *http.Request, t string) bool {
	return strings.Index(r.Header.Get("Accept"), t) > 0
}

// VerifyReferrer middlware validates the referer header matches the request url's host
func VerifyReferrer(c context.Context, w http.ResponseWriter, r *http.Request) context.Context {
	if r.Method != http.MethodPost {
		return c
	}

	err := func() error {
		referrer, err := url.Parse(r.Header.Get("Referer"))
		if err != nil {
			return err
		}
		if referrer.Host != r.Host {
			return fmt.Errorf("BLOCKED: VerifyReferrer: %v - %v", referrer, r.Host)
		}
		return nil
	}()

	if err != nil {
		cc, cancel := context.WithCancel(c)
		cancel()
		return cc
	}
	return c
}
