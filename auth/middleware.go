package auth

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/memcache"
)

const (
	tokensTable = "tokens"
	cookieName  = "app-cookie"
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
}

// AuthenticateCookie authenticates the token with a request cookie
func (m *Middleware) AuthenticateCookie(c context.Context, w http.ResponseWriter, r *http.Request) context.Context {
	var err error
	c, cancel := context.WithCancel(c)

	returnURL := fmt.Sprintf("/signin?returnUrl=%s", r.RequestURI)

	cookie, err := r.Cookie(cookieName)
	if err != nil {
		if !m.ContinueWithBadToken {
			http.Redirect(w, r, returnURL, http.StatusTemporaryRedirect)
			cancel()
		}
		return c
	}
	token, err := m.getToken(c, cookie.Value)
	if err != nil {
		if !m.ContinueWithBadToken {
			http.Redirect(w, r, returnURL, http.StatusTemporaryRedirect)
			cancel()
		}
		return c
	}

	// if token has expired return 401
	if token.isExpired() {
		if !m.ContinueWithBadToken {
			http.Redirect(w, r, returnURL, http.StatusTemporaryRedirect)
			cancel()
		}
		return c
	}

	accountKey := token.Key.Parent()

	// if the token's expiry less than a week away, get new token
	if token.willExpireIn(time.Hour * 24 * 7) {
		newToken, err := m.getNewToken(c, accountKey)
		if err != nil {
			if !m.ContinueWithBadToken {
				http.Redirect(w, r, returnURL, http.StatusTemporaryRedirect)
				cancel()
			}
			return c
		}

		// send back the new token values
		http.SetCookie(w, &http.Cookie{
			Name:     cookieName,
			Expires:  time.Now().Add(time.Hour * 24 * 14), // 2 weeks from now
			HttpOnly: true,
			Secure:   !appengine.IsDevAppServer(),
			Value:    newToken.Value(),
		})
	}

	// add accountKey to context
	c = m.Session.SetAccountKey(c, accountKey)

	return c
}

// AuthenticateToken authenticates the Authorization request header token
func (m *Middleware) AuthenticateToken(c context.Context, w http.ResponseWriter, r *http.Request) context.Context {
	// let option requests through
	if r.Method == http.MethodOptions {
		return c
	}

	var err error
	c, cancel := context.WithCancel(c)

	authHeader := r.Header.Get("Authorization")
	if len(authHeader) <= len("token=") {
		log.Errorf(c, "missing token header: %v", err)
		if !m.ContinueWithBadToken {
			w.WriteHeader(http.StatusUnauthorized)
			cancel()
		}
		return c
	}

	// prevent token caching with blank string value
	rawToken := authHeader[len("token="):]
	if len(rawToken) == 0 {
		log.Errorf(c, "missing token value: %v", err)
		if !m.ContinueWithBadToken {
			w.WriteHeader(http.StatusUnauthorized)
			cancel()
		}
		return c
	}

	token, err := m.getToken(c, rawToken)
	if err != nil {
		log.Errorf(c, "failed to get token details: %v", err)
		if !m.ContinueWithBadToken {
			w.WriteHeader(http.StatusUnauthorized)
			cancel()
		}
		return c
	}

	// if token has expired return 401
	if token.isExpired() {
		log.Errorf(c, "expired Token")
		if !m.ContinueWithBadToken {
			w.WriteHeader(http.StatusUnauthorized)
			cancel()
		}
		return c
	}

	accountKey := token.Key.Parent()

	// if the token's expiry less than a week away, get new token
	if token.willExpireIn(time.Hour * 24 * 7) {
		newToken, err := m.getNewToken(c, accountKey)
		if err != nil {
			log.Errorf(c, "failed to create new token: %v", err)
			if !m.ContinueWithBadToken {
				w.WriteHeader(http.StatusUnauthorized)
				cancel()
			}
			return c
		}

		// send back the new token values
		w.Header().Add(newTokenHeader, newToken.Value())
		w.Header().Add(newTokenExpiryHeader, newToken.Expiry.Format(time.RFC3339))
	}

	// add accountKey to context
	c = m.Session.SetAccountKey(c, accountKey)

	return c
}

// Gets the token for the rawToken value
func (m *Middleware) getToken(c context.Context, rawToken string) (*Token, error) {
	var err error

	var token Token
	_, err = memcache.JSON.Get(c, rawToken, &token)
	if err != nil && err != memcache.ErrCacheMiss {
		return nil, err
	}

	if err == memcache.ErrCacheMiss {
		tokenKey, err := datastore.DecodeKey(rawToken)
		if err != nil {
			return nil, fmt.Errorf("decoding token key: %v", err)
		}

		var store = NewTokenStore()
		err = store.Get(c, tokenKey, &token)
		if err != nil {
			return nil, err
		}

		// add the token to memcache
		err = memcache.JSON.Set(c, &memcache.Item{
			Key:        token.Value(),
			Object:     token,
			Expiration: -1 * time.Since(token.Expiry),
		})
	}

	return &token, err
}

// Creates a new token and links it to the account for the old token
func (m *Middleware) getNewToken(c context.Context, accountKey *datastore.Key) (*Token, error) {
	if accountKey == nil {
		return nil, errors.New("account key is required to create a token")
	}

	store := NewTokenStore()
	return store.Create(c, accountKey)
}
