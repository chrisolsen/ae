package auth

import (
	"errors"
	"fmt"

	"github.com/chrisolsen/ae"
	"github.com/chrisolsen/ae/store"
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
)

// Credentials contain authentication details for various providers / methods
type Credentials struct {
	ae.Model

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

// CredentialStore .
type CredentialStore struct {
	store.Base
}

// NewCredentialStore .
func NewCredentialStore() CredentialStore {
	s := CredentialStore{}
	s.TableName = "credentials"
	return s
}

// Create .
func (s *CredentialStore) Create(c context.Context, creds *Credentials, accountKey *datastore.Key) (*datastore.Key, error) {
	if !creds.Valid() {
		return nil, errors.New("Invalid credentials")
	}

	var isProvider = len(creds.ProviderID) > 0

	// already exists?
	q := datastore.NewQuery(s.TableName)
	q.Ancestor(accountKey)
	q.KeysOnly()
	if isProvider {
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

	if !isProvider {
		// encrypt password
		creds.Password, err = encrypt(creds.Password)
		if err != nil {
			return nil, fmt.Errorf("failed to encrypt password: %v", err)
		}
	}

	return s.Base.Create(c, creds, accountKey)
}

// GetAccountKeyByProvider .
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

// GetByAccount .
func (s *CredentialStore) GetByAccount(c context.Context, accountKey *datastore.Key, dst interface{}) ([]*datastore.Key, error) {
	return datastore.NewQuery(s.TableName).Ancestor(accountKey).GetAll(c, dst)
}

// UpdatePassword .
func (s CredentialStore) UpdatePassword(c context.Context, accountKey *datastore.Key, password string) error {
	var creds []*Credentials
	keys, err := datastore.NewQuery(s.TableName).
		Ancestor(accountKey).
		Filter("ProviderID =", "").
		GetAll(c, &creds)

	if err != nil {
		return err
	}
	if len(keys) == 0 {
		return errors.New("no credentials found")
	}
	if len(keys) > 1 {
		return errors.New("more than one credential found")
	}

	key, cred := keys[0], creds[0]
	cred.Password, err = encrypt(password)
	if err != nil {
		return errors.New("failed to encrypt password")
	}
	return s.Update(c, key, cred)
}
