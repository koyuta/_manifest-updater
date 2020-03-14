package main

import (
	"net/http"

	"github.com/koyuta/manifest-updater/handler"
	"github.com/koyuta/manifest-updater/pkg/router"
	"github.com/koyuta/manifest-updater/updater"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

// BuildRouter builds a http router.
func BuildRouter(queue chan<- *updater.Updater) http.Handler {
	var chiContext = router.NewChiRouter()
	var h = handler.NewHandler(chiContext, queue)

	var router = chi.NewRouter()
	router.Route("/app", func(r chi.Router) {
		r.Use(middleware.Recoverer)
		r.Get("/", h.Get)
		r.Post("/", h.Post)
	})
	router.Get("/healthz", h.Healthz)

	return router
}
