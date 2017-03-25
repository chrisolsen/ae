package flash

import (
	"net/http"
	"time"

	"google.golang.org/appengine"
)

// Set inserts a cookie into the response that contains the flash message
func Set(w http.ResponseWriter, msg string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "flash",
		Value:    msg,
		Expires:  time.Time{},
		Secure:   !appengine.IsDevAppServer(),
		HttpOnly: true,
		Path:     "/",
	})
}

// Get obtains the value within the flash cookie, then clears the value to prevent
// it from being seen in the next response.
func Get(w http.ResponseWriter, r *http.Request) string {
	c, err := r.Cookie("flash")
	if err != nil {
		return ""
	}
	val := c.Value
	Set(w, "")
	return val
}
