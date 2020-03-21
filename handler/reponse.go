package handler

import (
	"encoding/json"
	"net/http"
)

type ErrorResponse struct {
	Message string `json:"message"`
}

func NewErrorResponse(m string) *ErrorResponse {
	return &ErrorResponse{Message: m}
}

func WriteOKToHeader(rw http.ResponseWriter) {
	rw.WriteHeader(http.StatusOK)
}

func WriteBadRequestToHeader(rw http.ResponseWriter, resp *ErrorResponse) {
	rw.Header().Set("Content-Type", "application/json; charset=utf-8")
	rw.WriteHeader(http.StatusBadRequest)

	body, _ := json.Marshal(resp)
	rw.Write(body)
}
