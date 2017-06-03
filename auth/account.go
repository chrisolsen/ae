package auth

import (
	"errors"
	"fmt"

	"github.com/chrisolsen/ae"
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
)

const (
	AccountStateUnconfirmed = iota
	AccountStateConfirmed
	AccountStateSuspended
	AccountStateTerminated
)

// Account model
type Account struct {
	ae.Model
	State int `json:"-"`
}

// AccountStore .
type accountStore struct {
	ae.Store
}

// NewAccountStore returns a setup AccountStore
func newAccountStore() accountStore {
	s := accountStore{}
	s.TableName = "accounts"
	return s
}

type AccountSvc struct {
	accountStore

	credentialStore CredentialStore
}

func NewAccountSvc() AccountSvc {
	return AccountSvc{
		accountStore:    newAccountStore(),
		credentialStore: NewCredentialStore(),
	}
}

// Create creates a new account
func (s AccountSvc) Create(c context.Context, creds *Credentials) (*datastore.Key, error) {
	var err error
	var accountKey *datastore.Key
	var account Account

	if err = creds.Valid(); err != nil {
		return nil, err
	}
	accountKey, err = s.accountStore.Create(c, &account, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create account: %v", err)
	}
	_, err = s.credentialStore.Create(c, creds, accountKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create credentials: %v", err)
	}

	return accountKey, nil
}

// GetAccountKeyByCredentials fetches the account matching the auth provider credentials
func (s AccountSvc) GetAccountKeyByCredentials(c context.Context, creds *Credentials) (*datastore.Key, error) {
	var err error
	// on initial signup the account key will exist within the credentials
	if creds.AccountKey != nil {
		var accountCreds []*Credentials
		_, err = s.credentialStore.GetByAccount(c, creds.AccountKey, &accountCreds)
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
		return s.credentialStore.GetAccountKeyByProvider(c, creds)
	}

	// by username
	var userNameCreds []*Credentials
	ckeys, err := s.credentialStore.GetByUsername(c, creds.Username, &userNameCreds)
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
