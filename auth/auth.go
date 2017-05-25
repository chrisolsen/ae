package auth

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/chrisolsen/fbgraphapi"
	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
)

// Errors
var (
	ErrNoCookie    = errors.New("no cookie found")
	ErrNoAuthToken = errors.New("no header auth token found")
)

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
	if len(uuid) == 0 {
		return nil, ErrNoAuthToken
	}

	tstore := NewTokenStore()
	return tstore.Get(c, uuid)
}

// Signup creates a user account and links up the credentials. Based on the request type an auth cookie
// or header token will be set with an auth token.
func SignupByForm(c context.Context, w http.ResponseWriter, r *http.Request, creds *Credentials, keepCookie bool) (*datastore.Key, error) {
	token, err := signup(c, creds)
	if err != nil {
		return nil, err
	}
	SetAuthCookieToken(w, token.UUID, keepCookie)
	return token.AccountKey(), nil
}

func SignupByAPI(c context.Context, w http.ResponseWriter, r *http.Request, creds *Credentials) (*datastore.Key, error) {
	token, err := signup(c, creds)
	if err != nil {
		return nil, err
	}
	setHeaderToken(w, token.UUID)
	return token.AccountKey(), nil
}

func signup(c context.Context, creds *Credentials) (*Token, error) {
	astore := NewAccountStore()
	tstore := NewTokenStore()
	accountKey, err := astore.Create(c, creds)
	if err != nil {
		return nil, err
	}
	token, err := tstore.Create(c, accountKey)
	if err != nil {
		return nil, err
	}
	return token, nil
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

// Authenticate .
func AuthenticateHeader(c context.Context, w http.ResponseWriter, r *http.Request, creds *Credentials) (*Token, error) {
	var token *Token
	var err error
	token, err = doExternalAuth(c, creds, appEngineURLGetter{Ctx: c})
	if err != nil {
		return nil, err
	}
	setHeaderToken(w, token.UUID)
	return token, nil
}

func AuthenticateForm(c context.Context, w http.ResponseWriter, r *http.Request, creds *Credentials, keepCookie bool) (*Token, error) {
	var token *Token
	var err error
	token, err = doInternalAuth(c, creds)
	if err != nil {
		return nil, err
	}
	SetAuthCookieToken(w, token.UUID, keepCookie)
	return token, nil
}

func doInternalAuth(c context.Context, creds *Credentials) (*Token, error) {
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

func doExternalAuth(c context.Context, creds *Credentials, urlGetter urlGetter) (*Token, error) {
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

func SetAuthCookieToken(w http.ResponseWriter, uuid string, keepCookie bool) {
	c := &http.Cookie{
		Name:     cookieName,
		Path:     "/",
		HttpOnly: true,
		Secure:   !appengine.IsDevAppServer(),
		Value:    uuid,
	}
	if keepCookie {
		c.Expires = time.Now().Add(time.Hour * 24 * 14)
	}
	http.SetCookie(w, c)
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
