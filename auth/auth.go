package auth

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/chrisolsen/fbgraphapi"
	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
)

// SignupViaForm creates the acount, credentials and token for a new account and sets the auth cookie in the response
func SignupViaForm(c context.Context, w http.ResponseWriter, account *Account, credentials *Credentials) error {
	var err error

	// ensure that the username/email is not already used
	cstore := NewCredentialStore()
	key, err := cstore.GetByUsername(c, credentials.Username, nil)
	if err != nil && err != datastore.ErrInvalidEntityType {
		return fmt.Errorf("failed to find user by username: %v", err)
	}
	if key != nil {
		return errors.New("Username is already taken")
	}

	// encrypt password
	credentials.Password, err = encrypt(credentials.Password)
	if err != nil {
		return fmt.Errorf("failed to encrypt password: %v", err)
	}

	// create account and credentials
	astore := NewAccountStore()
	accountKey, err := astore.Create(c, credentials, account)
	if err != nil {
		return fmt.Errorf("failed to create account: %v", err)
	}

	// insert auth cookie into response
	tstore := NewTokenStore()
	token, err := tstore.Create(c, accountKey)
	if err != nil {
		return fmt.Errorf("failed to create token: %v", err)
	}

	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Expires:  time.Now().Add(time.Hour * 24 * 14), // 2 weeks from now
		HttpOnly: true,
		Secure:   !appengine.IsDevAppServer(),
		Value:    token.Value(),
	})

	return nil
}

// Authenticate validates that the credentials match an account; if so creates
// and links a new token to the account
// POST /api/auth
//  {
//  	"providerName": "facebook",
//  	"providerId": "users-provider-id",
//  	"providerToken": "provided-token"
//  }
func Authenticate(c context.Context, creds *Credentials) (*Token, error) {
	// Calls the private method with an appengine urlGetter.
	// This allows for internal testing of the authenticate method while
	// stubbing out the external auth service
	return authenticate(c, creds, appEngineURLGetter{Ctx: c})
}

func authenticate(c context.Context, creds *Credentials, urlGetter urlGetter) (*Token, error) {
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
