package auth

import (
	"errors"
	"fmt"

	"github.com/chrisolsen/ae/attachment"
	"github.com/chrisolsen/ae/model"
	"github.com/chrisolsen/ae/store"
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
)

// Account model
type Account struct {
	model.Base

	// Allows for manually making a user an owner via the datastore web interface.
	// Any additional control should exist within the app with a Roles table
	IsOwner   bool   `json:"isOwner,omitempty" datastore:",noindex"`
	FirstName string `json:"firstName" datastore:",noindex"`
	LastName  string `json:"lastName" datastore:",noindex"`
	Gender    string `json:"gender" datastore:",noindex"`
	Locale    string `json:"locale" datastore:",noindex"`
	Location  string `json:"location" datastore:",noindex"`
	Name      string `json:"name" datastore:",noindex"`
	Timezone  int    `json:"timezone" datastore:",noindex"`
	Email     string `json:"email"`

	Photo attachment.File `json:"photo"`
}

// AccountStore .
type AccountStore struct {
	store.Base
}

// NewAccountStore returns a setup AccountStore
func NewAccountStore() AccountStore {
	s := AccountStore{}
	s.TableName = "accounts"
	return s
}

// Create creates a new account
func (s *AccountStore) Create(c context.Context, creds *Credentials, account *Account) (*datastore.Key, error) {
	var err error
	var accountKey *datastore.Key
	var cStore = NewCredentialStore()
	err = datastore.RunInTransaction(c, func(tc context.Context) error {
		accountKey, err = s.Base.Create(tc, account, nil)
		if err != nil {
			return fmt.Errorf("failed to create account: %v", err)
		}

		_, err = cStore.Create(tc, creds, accountKey)
		if err != nil {
			return fmt.Errorf("failed to create credentials: %v", err)
		}

		return nil
	}, &datastore.TransactionOptions{XG: true})

	return accountKey, nil
}

// GetAccountKeyByCredentials fetches the account matching the auth provider credentials
func (s *AccountStore) GetAccountKeyByCredentials(c context.Context, creds *Credentials) (*datastore.Key, error) {
	var err error
	cstore := NewCredentialStore()
	// on initial signup the account key will exist within the credentials
	if creds.AccountKey != nil {
		var accountCreds []*Credentials
		_, err = cstore.GetByAccount(c, creds.AccountKey, &accountCreds)
		if err != nil {
			return nil, fmt.Errorf("failed to find credentials by parent account: %v", err)
		}
		// validate credentials
		for _, ac := range accountCreds {
			if ac.ProviderID == creds.ProviderID && ac.ProviderName == creds.ProviderName {
				return creds.AccountKey, nil
			}
		}
		return nil, errors.New("no matching credentials found for account")
	}

	// by provider
	if len(creds.ProviderID) > 0 {
		return cstore.GetAccountKeyByProvider(c, creds)
	}

	// by username
	var userNameCreds []*Credentials
	ckeys, err := cstore.GetByUsername(c, creds.Username, &userNameCreds)
	if err != nil {
		return nil, err
	}

	if len(userNameCreds) != 1 {
		return nil, errors.New("unable to find unique credentials")
	}

	err = checkCrypt(userNameCreds[0].Password, creds.Password)
	if err != nil {
		return nil, err
	}
	return ckeys[0].Parent(), nil
}
