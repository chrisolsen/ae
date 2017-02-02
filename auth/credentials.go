package auth

import (
	"errors"
	"fmt"

	"github.com/chrisolsen/ae/model"
	"github.com/chrisolsen/ae/store"

	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
)

// Credentials contain authentication details for various providers / methods
type Credentials struct {
	model.Base

	// passed in on initial signup since looking up credentials by non-key cols
	// may result in an empty dataset
	AccountKey *datastore.Key `json:"accountKey" datastore:"-"`

	// oauth
	ProviderID   string `json:"providerId"`
	ProviderName string `json:"providerName"`

	// token is not saved
	ProviderToken string `json:"providerToken" datastore:"-"`

	// username / password
	Username string `json:"username"`
	Password string `json:"password"`
}

// Valid indicates if the credentials are valid for one of the two credential types
func (c *Credentials) Valid() bool {
	p := len(c.ProviderID) > 0 && len(c.ProviderName) > 0 && len(c.ProviderToken) > 0
	l := len(c.Username) > 0 && len(c.Password) > 0
	return p || l
}

type CredentialStore struct {
	store.Base
}

func NewCredentialStore() CredentialStore {
	s := CredentialStore{}
	s.TableName = "credentials"
	return s
}

func (s *CredentialStore) Create(c context.Context, creds *Credentials, accountKey *datastore.Key) (*datastore.Key, error) {
	if !creds.Valid() {
		return nil, errors.New("Invalid credentials")
	}

	q := datastore.NewQuery(s.TableName)
	q.Ancestor(accountKey)
	q.KeysOnly()
	if len(creds.ProviderID) > 0 {
		q.Filter("ProviderID =", creds.ProviderID)
		q.Filter("ProviderName =", creds.ProviderName)
	} else {
		q.Filter("Username =", creds.Username)
	}
	keys, err := q.GetAll(c, nil)
	if err != nil {
		if err != datastore.ErrInvalidEntityType {
			return nil, err
		}
	}
	if len(keys) > 0 {
		return nil, errors.New("account credentials already exists")
	}

	return s.Base.Create(c, creds, accountKey)
}

func (s *CredentialStore) GetAccountKeyByProvider(c context.Context, creds *Credentials) (*datastore.Key, error) {
	keys, err := datastore.NewQuery(s.TableName).
		Filter("ProviderID =", creds.ProviderID).
		Filter("ProviderName =", creds.ProviderName).
		KeysOnly().
		GetAll(c, nil)

	if err != nil {
		return nil, fmt.Errorf("finding account by auth provider: %v", err)
	}

	if len(keys) == 0 {
		return nil, errors.New("no account found matching the auth provider")
	}

	return keys[0].Parent(), nil
}

// GetByUsername .
func (s *CredentialStore) GetByUsername(c context.Context, username string, dst interface{}) ([]*datastore.Key, error) {
	return datastore.NewQuery(s.TableName).Filter("Username =", username).GetAll(c, dst)
}
