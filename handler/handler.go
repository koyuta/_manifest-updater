package handler

import (
	"net/http"

	"github.com/koyuta/manifest-updater/pkg/registry"
	"github.com/koyuta/manifest-updater/pkg/repository"
	"github.com/koyuta/manifest-updater/pkg/router"
	"github.com/koyuta/manifest-updater/updater"
)

type Handler struct {
	router router.Router
	queue  chan<- *updater.Updater
}

type PostRequest struct {
	DockerHub string
	Filter    string
	Git       string
	Branch    string
	Path      string
	DocerHub  string
}

func NewHandler(r *router.ChiRouter) *Handler {
	return &Handler{router: r}
}

func (h *Handler) Get(rw http.ResponseWriter, r *http.Request) {
	rw.WriteHeader(http.StatusOK)
}

func (h *Handler) Post(rw http.ResponseWriter, r *http.Request) {
	var req = PostRequest{}
	if err := RetrieveBody(r.Body, &req); err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	updater := updater.NewUpdater(
		registry.NewDockerHubRegistry(req.DockerHub, req.Filter),
		repository.NewGitHubRepository(req.Git, req.Branch, req.Path, req.DocerHub),
	)
	h.queue <- updater

	rw.WriteHeader(http.StatusOK)
}

func (h *Handler) Healthz(rw http.ResponseWriter, r *http.Request) {
	rw.WriteHeader(http.StatusOK)
}
