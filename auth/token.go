package auth

import (
	"time"

	"github.com/chrisolsen/ae/model"
	"github.com/chrisolsen/ae/store"

	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
)

// Token .
type Token struct {
	model.Base
	Expiry time.Time `json:"expiry" datastore:",noindex"`
}

func (t *Token) isExpired() bool {
	return t.Expiry.Before(time.Now())
}

func (t *Token) willExpireIn(duration time.Duration) bool {
	future := time.Now().Add(duration)
	return t.Expiry.Before(future)
}

// Load .
func (t *Token) Load(ps []datastore.Property) error {
	if err := datastore.LoadStruct(t, ps); err != nil {
		return err
	}
	return nil
}

// Save .
func (t *Token) Save() ([]datastore.Property, error) {
	if t.Expiry.IsZero() {
		t.Expiry = time.Now().AddDate(0, 2, 0)
	}
	return datastore.SaveStruct(t)
}

// Value returns the datastore key's string value
func (t *Token) Value() string {
	return t.Key.Encode()
}

// TokenStore .
type TokenStore struct {
	store.Base
}

// NewTokenStore .
func NewTokenStore() TokenStore {
	s := TokenStore{}
	s.TableName = "tokens"
	return s
}

// Create overrides base method since token creation doesn't need any data
// other than the account key
func (s *TokenStore) Create(c context.Context, accountKey *datastore.Key) (*Token, error) {
	var token Token
	key, err := s.Base.Create(c, &token, accountKey)
	if err != nil {
		return nil, err
	}
	token.Key = key
	return &token, nil
}
