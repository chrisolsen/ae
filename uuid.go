package ae

import gouuid "github.com/satori/go.uuid"

// This method only exists because AppEngine fails to compile if
// the same lib is vendored in multiple packages and used within an app.

// NewV4UUID returns random generated UUID.
func NewV4UUID() string {
	return gouuid.NewV4().String()
}
