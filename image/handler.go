package image

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/chrisolsen/ae/handler"
	"golang.org/x/net/context"
)

// Handler handles Google storage image requests
type Handler struct {
	handler.Base
}

func (h Handler) ServeHTTP(c context.Context, w http.ResponseWriter, r *http.Request) {
	h.Bind(c, w, r)
	switch r.Method {
	case http.MethodGet:
		h.fetch()
	default:
		h.Abort(http.StatusNotFound, nil)
	}
}

// GET /images/:name?w={100}&h={100}
func (h *Handler) fetch() {
	name := h.PathParam("/images/:name")
	if name == "" {
		h.Abort(http.StatusBadRequest, errors.New("name value required"))
		return
	}

	width, _ := strconv.Atoi(h.QueryParam("w"))
	height, _ := strconv.Atoi(h.QueryParam("h"))
	if width+height == 0 {
		h.Abort(http.StatusBadRequest, errors.New("width or height is required"))
		return
	}

	// fetch url for required size
	scheme := "https"
	if h.Req.TLS == nil {
		scheme = "http"
	}
	url, err := SizedURL(h.Ctx, scheme, name, width, height)
	if err != nil {
		h.Abort(http.StatusInternalServerError, fmt.Errorf("failed to get sized image: %v", err))
		return
	}
	http.Redirect(h.Res, h.Req, url, http.StatusMovedPermanently)
}
