package auth

import (
	"os"
)

var csrfSecret string
var anonUUID string

func init() {
	csrfSecret = os.Getenv("CSRF_SECRET")
	anonUUID = os.Getenv("ANON_UUID")
}
