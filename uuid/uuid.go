package uuid

import suuid "github.com/satori/go.uuid"

// This package only exists becase AppEngine is barfs all over the fuckin' place if
// the same lib is vendored in multiple packages and used within an app.

// New returns random generated UUID.
func New() string {
	return suuid.NewV4().String()
}
