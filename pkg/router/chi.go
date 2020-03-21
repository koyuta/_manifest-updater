package router

import (
	"net/http"

	"github.com/go-chi/chi"
)

type ChiRouter struct {
	context *chi.Context
}

func NewChiRouter() *ChiRouter {
	return &ChiRouter{context: chi.NewRouteContext()}
}

func (c *ChiRouter) URLParam(r *http.Request, key string) string {
	return chi.URLParam(r, key)
}
