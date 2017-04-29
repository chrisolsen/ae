package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/chrisolsen/ae"
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/memcache"
)

var ErrInvalidToken = errors.New("invalid token")

// Token .
type Token struct {
	ae.Model
	UUID   string    `json:"uuid"`
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
		t.Expiry = time.Now().AddDate(0, 0, 14)
	}
	return datastore.SaveStruct(t)
}

// TokenStore .
type TokenStore struct {
	ae.Store
}

// NewTokenStore .
func NewTokenStore() TokenStore {
	s := TokenStore{}
	s.TableName = "tokens"
	return s
}

// Get overrides the base get to allow lookup by the uuid rather than a key
func (s *TokenStore) Get(c context.Context, UUID string) (*Token, error) {
	var err error
	var tokens []*Token
	var cachedToken Token

	if len(UUID) == 0 {
		return nil, ErrInvalidToken
	}
	_, err = memcache.JSON.Get(c, UUID, &cachedToken)
	if err == nil {
		return &cachedToken, nil
	}
	if err != memcache.ErrCacheMiss {
		return nil, err
	}

	keys, err := datastore.NewQuery(s.TableName).Filter("UUID =", UUID).GetAll(c, &tokens)
	if err != nil {
		return nil, err
	}
	if len(tokens) == 0 {
		return nil, errors.New("invalid token")
	}
	if len(tokens) > 1 {
		return nil, errors.New("multiple tokens found")
	}
	tokens[0].Key = keys[0]

	memcache.JSON.Set(c, &memcache.Item{
		Key:        UUID,
		Object:     tokens[0],
		Expiration: time.Hour * 24 * 14,
	})

	return tokens[0], nil
}

// Create overrides base method since token creation doesn't need any data
// other than the account key
func (s *TokenStore) Create(c context.Context, accountKey *datastore.Key) (*Token, error) {
	token := Token{UUID: ae.NewV4UUID()}
	_, err := s.Store.Create(c, &token, accountKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create token: %v", err)
	}
	return &token, nil
}

// Delete .
func (s *TokenStore) Delete(c context.Context, uuid string) error {
	token, err := s.Get(c, uuid)
	if err != nil {
		return err
	}
	return s.Store.Delete(c, token.Key)
}
