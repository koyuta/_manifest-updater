package handler

import (
	"net/http"

	"github.com/koyuta/manifest-updater/pkg/router"
)

type Handler struct {
	router *router.ChiRouter
}

func NewHandler(r *router.ChiRouter) *Handler {
	return &Handler{router: r}
}

func (h *Handler) Get(rw http.ResponseWriter, r *http.Request) {
}

func (h *Handler) Healthz(rw http.ResponseWriter, r *http.Request) {
	rw.WriteHeader(http.StatusOK)
}
