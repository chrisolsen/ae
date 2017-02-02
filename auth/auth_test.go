package auth

import (
	"os"
	"testing"

	"github.com/chrisolsen/ae/testutils"
	"google.golang.org/appengine/datastore"
)

var T = testutils.T{}

func Setup(m *testing.M) {
	os.Exit(func() int {
		id := m.Run()
		T.Close()
		return id
	}())
}

func TestEndpoints_Auth(t *testing.T) {
	c := T.GetContext()

	type signupTest struct {
		name        string
		creds       *Credentials
		expectedErr error
	}

	var tests = []signupTest{
		{
			name:        "Invalid account",
			creds:       &Credentials{ProviderName: "facebook", ProviderID: "1234", ProviderToken: "asoiudykaejhes"},
			expectedErr: nil,
		},
	}

	// account to auth with
	a := Account{}
	pkey, _ := datastore.Put(c, datastore.NewIncompleteKey(c, "accounts", nil), &a)
	creds := Credentials{ProviderID: "1234", ProviderName: "facebook", ProviderToken: "asoiudykaejhes"}
	ckey, _ := datastore.Put(c, datastore.NewIncompleteKey(c, "credentials", pkey), &creds)

	for _, ts := range tests {
		ts.creds.Key = ckey
		ts.creds.AccountKey = pkey // prevent the data propogation issue
		func(test signupTest) {
			urlGetter := testutils.MockURLGetter{Err: test.expectedErr, Body: `{"id": "1234"}`}

			token, err := authenticate(c, test.creds, urlGetter)
			if err != nil && test.expectedErr == nil {
				t.Error("Unexpected error", err.Error())
				return
			}

			if err != nil && test.expectedErr != nil {
				return
			}

			if len(token.Value()) == 0 {
				t.Error("No token returned")
				return
			}
		}(ts)
	}
}
