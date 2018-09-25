package auth

import (
	"context"
	"errors"
	"fmt"

	"google.golang.org/appengine/datastore"
)

var (
	sessionKey = contextKey("session-key")
)

// Errors
var (
	ErrMissingToken  = errors.New("no auth token found")
	ErrNoSuchAccount = errors.New("failed to find account")
)

// Session provides helper methods to get and set the account key within the request context
type Session struct{}

// SignedIn returns boolean value indicating if the user is signed in or not
func (s *Session) SignedIn(c context.Context) bool {
	key, err := s.AccountKey(c)
	if err != nil {
		return false
	}
	return key != nil
}

// AccountKey return the *datastore.Key value for the account
func (s *Session) AccountKey(c context.Context) (*datastore.Key, error) {
	var err error
	val := c.Value(sessionKey)
	if val == nil {
		return nil, ErrMissingToken
	}

	key, err := datastore.DecodeKey(val.(string))
	if err != nil {
		return nil, fmt.Errorf("decoding the context account key: %v", err)
	}

	return key, nil
}

// SetAccountKey sets the key in the request context to allow for later access
func (s *Session) SetAccountKey(c context.Context, key *datastore.Key) context.Context {
	return context.WithValue(c, sessionKey, key.Encode())
}
