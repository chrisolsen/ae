package uuid

import gouuid "github.com/satori/go.uuid"

// This package only exists because AppEngine fails to compile if
// the same lib is vendored in multiple packages and used within an app.

// Random returns random generated UUID.
func Random() string {
	return gouuid.NewV4().String()
}
