package auth

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"golang.org/x/net/context"
	"google.golang.org/appengine/log"
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
	if r.Method == http.MethodGet {
		return c
	}

	var csrf string
	c2, cancel := context.WithCancel(c)

	err := func() error {
		if strings.Contains(r.Header.Get("Content-Type"), "multipart/form-data") {
			if err := r.ParseMultipartForm(10 << 20); err != nil {
				return fmt.Errorf("error parsing multipart data: %v", err)
			}
			csrf = r.MultipartForm.Value["csrfToken"][0]
		} else {
			if err := r.ParseForm(); err != nil {
				return fmt.Errorf("error pareing data: %v", err)
			}
			csrf = r.FormValue("csrfToken")
		}

		// var auth Token
		uuid, err := getUUIDFromCookie(r)
		if err == ErrNoCookie {
			uuid = anonUUID
		} else if err != nil {
			return err
		}

		if csrf == "" {
			return errors.New("No CSRF token is present in request body")
		}

		if err := checkCrypt(csrf, csrfSecret+uuid); err != nil {
			return fmt.Errorf("failed checkCrypt: %v", err)
		}
		return nil
	}()

	if err != nil {
		log.Errorf(c, err.Error())
		cancel()
	}

	return c2
}
