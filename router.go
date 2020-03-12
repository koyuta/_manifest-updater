package main

import (
	"net/http"

	"github.com/koyuta/manifest-updater/handler"
	"github.com/koyuta/manifest-updater/pkg/router"

	"github.com/go-chi/chi"
)

func BuildRouter() http.Handler {
	var chiContext = router.NewChiRouter()
	var h = handler.NewHandler(chiContext)

	var router = chi.NewRouter()
	router.Route("/", func(r chi.Router) {
		//r.Use()
		r.Get("/", h.Get)
	})
	router.Get("/healthz", h.Healthz)

	return router
}
