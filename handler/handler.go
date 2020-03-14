package handler

import (
	"errors"
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
	DockerHub string `json:"dockerHub"`
	Filter    string `json:"filer"`
	Git       string `json:"git"`
	Branch    string `json:"branch"`
	Path      string `json:"path"`
}

func (p *PostRequest) Unmarshal() (*updater.Updater, error) {
	// TODO: Add supports for other registry providers.
	if p.DockerHub == "" {
		return nil, errors.New("dockerHub must be specified")
	}
	if p.Git == "" {
		return nil, errors.New("git must be specified")
	}

	updater := updater.NewUpdater(
		registry.NewDockerHubRegistry(p.DockerHub, p.Filter),
		repository.NewGitHubRepository(p.Git, p.Branch, p.Path, p.DockerHub),
	)
	return updater, nil
}

func NewHandler(r router.Router, q chan<- *updater.Updater) *Handler {
	return &Handler{router: r, queue: q}
}

func (h *Handler) Get(rw http.ResponseWriter, r *http.Request) {
	WriteOKToHeader(rw)
}

func (h *Handler) Post(rw http.ResponseWriter, r *http.Request) {
	var req = &PostRequest{}
	if err := RetrieveBody(r.Body, req); err != nil {
		WriteBadRequestToHeader(rw, NewErrorResponse("invalid request"))
		return
	}
	updater, err := req.Unmarshal()
	if err != nil {
		WriteBadRequestToHeader(rw, NewErrorResponse(err.Error()))
		return
	}

	h.queue <- updater

	WriteOKToHeader(rw)
}

func (h *Handler) Healthz(rw http.ResponseWriter, r *http.Request) {
	WriteOKToHeader(rw)
}
