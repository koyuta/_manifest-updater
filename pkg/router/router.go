package router

import (
	"net/http"
)

type Router interface {
	URLParam(r *http.Request, key string) string
}
