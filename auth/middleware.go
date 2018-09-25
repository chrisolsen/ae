package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
)

const (
	tokensTable = "tokens"
	cookieName  = "token"
)

// Token keys
const (
	newTokenHeader       string = "new-auth-token"
	newTokenExpiryHeader string = "new-auth-token-expiry"
)

// Middleware .
type Middleware struct {
	// Reference to the session that is used to set/get account data
	Session Session

	// Whether the request should be allowed to continue if the token has expired, or no token exists.
	// This is useful for pages or endpoints that render/return differently based on whethe the user
	// is authenticated or not
	ContinueWithBadToken bool

	// The signin URL of the web app that the middleware will redirect to on failed cookie auth
	SignInURL string
}

// AuthCookieName returns the name of the auth cook
func AuthCookieName() string {
	return cookieName
}

// AuthenticateCookie authenticates the token with a request cookie
func (m *Middleware) AuthenticateCookie(w http.ResponseWriter, r *http.Request) {
	var accountKey *datastore.Key

	returnURL := fmt.Sprintf("%s?returnUrl=%s", m.SignInURL, r.RequestURI)

	c := r.Context()
	run := func() error {
		cookie, err := r.Cookie(cookieName)
		if err != nil {
			return fmt.Errorf("failed getting cookie: %v", err)
		}

		token, err := m.getToken(c, cookie.Value)
		if err != nil {
			return fmt.Errorf("failed to get token: %v", err)
		}
		if token.isExpired() {
			return errors.New("expired token")
		}

		accountKey = token.Key.Parent()

		// if the token's expiry less than a week away, get new token and kill the current
		if token.willExpireIn(time.Hour * 24 * 7) {
			newToken, err := m.getNewToken(c, token)
			if err != nil {
				return fmt.Errorf("failed to get new token: %v", err)
			}

			http.SetCookie(w, &http.Cookie{
				Name:     cookieName,
				Expires:  time.Now().Add(time.Hour * 24 * 14),
				HttpOnly: true,
				Secure:   !appengine.IsDevAppServer(),
				Value:    newToken.UUID,
			})
		}

		// add accountKey to context
		c = m.Session.SetAccountKey(c, accountKey)
		*r = *r.WithContext(c)
		return nil
	}

	if err := run(); err != nil && !m.ContinueWithBadToken {
		_, cancel := context.WithCancel(c)
		http.Redirect(w, r, returnURL, http.StatusTemporaryRedirect)
		cancel()
	}
}

// AuthenticateToken authenticates the Authorization request header token
func (m *Middleware) AuthenticateToken(w http.ResponseWriter, r *http.Request) {
	// let option requests through
	if r.Method == http.MethodOptions {
		return
	}

	c := r.Context()

	run := func() error {
		authHeader := r.Header.Get("Authorization")
		if len(authHeader) <= len("token=") {
			return errors.New("missing token header")
		}

		// prevent token caching with blank string value
		rawToken := authHeader[len("token="):]
		if len(rawToken) == 0 {
			return errors.New("missing token value")
		}

		token, err := m.getToken(c, rawToken)
		if err != nil {
			return fmt.Errorf("failed to get token details: %v", err)
		}

		// if token has expired return 401
		if token.isExpired() {
			return errors.New("expired token")
		}

		accountKey := token.Key.Parent()

		// if the token's expiry less than a week away, get new token
		if token.willExpireIn(time.Hour * 24 * 7) {
			newToken, err := m.getNewToken(c, token)
			if err != nil {
				return fmt.Errorf("failed to create new token: %v", err)
			}

			// send back the new token values
			w.Header().Add(newTokenHeader, newToken.UUID)
			w.Header().Add(newTokenExpiryHeader, newToken.Expiry.Format(time.RFC3339))
		}

		// add accountKey to context
		c = m.Session.SetAccountKey(c, accountKey)

		return nil
	}

	if err := run(); err != nil && !m.ContinueWithBadToken {
		_, cancel := context.WithCancel(c)
		w.WriteHeader(http.StatusUnauthorized)
		cancel()
	}
}

// Gets the token for the rawToken value
func (m *Middleware) getToken(c context.Context, uuid string) (*Token, error) {
	var store = newTokenStore()
	token, err := store.Get(c, uuid)
	if err != nil {
		return nil, fmt.Errorf("failed to get token from store: %v", err)
	}
	return token, nil
}

// Creates a new token and links it to the account for the old token
func (m *Middleware) getNewToken(c context.Context, oldToken *Token) (*Token, error) {
	store := newTokenStore()
	newToken, err := store.Create(c, oldToken.Key.Parent())
	if err != nil {
		return nil, err
	}

	store.Delete(c, oldToken.UUID)
	return newToken, nil
}
