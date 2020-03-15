package handler

import (
	"errors"
	"net/http"

	"github.com/koyuta/manifest-updater/pkg/router"
	"github.com/koyuta/manifest-updater/updater"
)

type Handler struct {
	router router.Router
	queue  chan<- *updater.Entry
}

type PostRequest struct {
	DockerHub string `json:"dockerHub"`
	Filter    string `json:"filer"`
	Git       string `json:"git"`
	Branch    string `json:"branch"`
	Path      string `json:"path"`
}

func (p *PostRequest) Unmarshal() (*updater.Entry, error) {
	// TODO: Add supports for other registry providers.
	if p.DockerHub == "" {
		return nil, errors.New("dockerHub must be specified")
	}
	if p.Git == "" {
		return nil, errors.New("git must be specified")
	}

	entry := &updater.Entry{
		DockerHub: p.DockerHub,
		Filter:    p.Filter,
		Git:       p.Git,
		Branch:    p.Branch,
		Path:      p.Path,
	}
	return entry, nil
}

func NewHandler(r router.Router, q chan<- *updater.Entry) *Handler {
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
