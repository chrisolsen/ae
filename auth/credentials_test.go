package auth

import "testing"

func TestCredentials_Valid(t *testing.T) {
	type data struct {
		name  string
		creds *Credentials
		valid bool
	}

	tests := []data{
		data{name: "missing params", creds: &Credentials{}, valid: false},
		data{name: "missing password", creds: &Credentials{Username: "foo"}, valid: false},
		data{name: "missing username", creds: &Credentials{Password: "foo"}, valid: false},
		data{name: "missing provider name and token", creds: &Credentials{ProviderID: "foo"}, valid: false},
		data{name: "missing provider id and token", creds: &Credentials{ProviderName: "foo"}, valid: false},
		data{name: "missing provider name and id", creds: &Credentials{ProviderToken: "foo"}, valid: false},
		data{name: "missing provider name", creds: &Credentials{ProviderID: "foo", ProviderToken: "foo"}, valid: false},
		data{name: "missing provider id", creds: &Credentials{ProviderName: "foo", ProviderToken: "foo"}, valid: false},
		data{name: "missing provider token", creds: &Credentials{ProviderName: "foo", ProviderID: "foo"}, valid: false},
		data{name: "valid provider", creds: &Credentials{ProviderName: "foo", ProviderID: "foo", ProviderToken: "foo"}, valid: true},
		data{name: "valid username/password", creds: &Credentials{Username: "foo", Password: "foo"}, valid: true},
	}

	for _, test := range tests {
		if test.creds.Valid() != test.valid {
			t.Errorf("failed: %v", test.name)
		}
	}
}

func TestCredentials_Create(t *testing.T) {
	type data struct {
		name  string
		creds *Credentials
		ok    bool
	}

	astore := NewAccountStore()
	cstore := NewCredentialStore()

	// pre-existing data
	ctx := T.GetContext()

	key, _ := astore.Create(ctx, &Credentials{Username: "jim", Password: "foobar"}, &Account{Name: "jim"})
	astore.Create(ctx, &Credentials{ProviderID: "123", ProviderName: "foobar"}, &Account{Name: "jim"})
	var dst Account
	astore.Get(ctx, key, &dst)

	tests := []data{
		data{name: "invalid", creds: &Credentials{}, ok: false},
		data{name: "matching provider id, but different name", creds: &Credentials{ProviderID: "1234", ProviderName: "google", ProviderToken: "8q7wris"}, ok: true},
		data{name: "duplicate provider id", creds: &Credentials{ProviderID: "1234", ProviderName: "foobar", ProviderToken: "34345"}, ok: false},
		data{name: "duplicate username", creds: &Credentials{Username: "jim", Password: "foobar"}, ok: false},
	}

	for _, test := range tests {
		_, err := cstore.Create(ctx, test.creds, key)
		if test.ok && err != nil {
			t.Errorf("%s: %v", test.name, err)
		}
		if !test.ok && err == nil {
			t.Errorf("%s: %v", test.name, err)
		}
	}
}

func TestCredentials_GetAccountKeyByProvider(t *testing.T) {

}
