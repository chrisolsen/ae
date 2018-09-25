package image

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/chrisolsen/ae"
)

// Handler handles Google storage image requests
type Handler struct {
	ae.Handler
}

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.Bind(w, r)
	route := ae.NewRoute(r)

	switch {
	case route.Matches("GET", "/images/:name"):
		h.fetch(route.Get("name"))
	default:
		h.Abort(http.StatusNotFound, nil)
	}
}

// GET /images/:name?w={100}&h={100}
func (h *Handler) fetch(name string) {
	url := h.Req.URL
	if name == "" {
		h.Abort(http.StatusBadRequest, errors.New("name value required"))
		return
	}

	width, _ := strconv.Atoi(url.Query().Get("w"))
	height, _ := strconv.Atoi(url.Query().Get("h"))
	if width+height == 0 {
		h.Abort(http.StatusBadRequest, errors.New("width or height is required"))
		return
	}

	// fetch url for required size
	scheme := "https"
	if h.Req.TLS == nil {
		scheme = "http"
	}
	sizedURL, err := SizedURL(h.Ctx(), scheme, name, width, height)
	if err != nil {
		h.Abort(http.StatusInternalServerError, fmt.Errorf("failed to get sized image: %v", err))
		return
	}
	http.Redirect(h.Res, h.Req, sizedURL, http.StatusMovedPermanently)
}
