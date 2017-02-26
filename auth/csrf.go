package auth

import (
	"net/http"

	"golang.org/x/net/context"
)

// NewCSRFToken creates token
func NewCSRFToken(r *http.Request) string {
	var uuid string
	cookie, err := r.Cookie(cookieName)
	if err != nil || len(cookie.Value) == 0 {
		uuid = anonUUID
	} else {
		uuid = cookie.Value
	}

	val, err := encrypt(csrfSecret + uuid)
	if err != nil {
		return ""
	}
	return val
}

// VerifyCSRFToken middleware method to check token
func VerifyCSRFToken(c context.Context, w http.ResponseWriter, r *http.Request) context.Context {
	if r.Method != http.MethodPost {
		return c
	}

	c2, cancel := context.WithCancel(c)
	r.ParseForm()
	csrf := r.FormValue("csrfToken")

	// var auth Token
	uuid, err := getUUIDFromCookie(r)
	if err == ErrNoCookie {
		uuid = anonUUID
	} else if err != nil {
		cancel()
		return c2
	}

	if err := checkCrypt(csrf, csrfSecret+uuid); err != nil {
		cancel()
		return c2
	}

	return c2
}
