package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	retrievalservice "ragtime-backend/internal/retrieval/service"
)

func NewRouter(service *retrievalservice.Service) http.Handler {
	r := chi.NewRouter()
	h := NewHandler(service)
	r.Post("/v1/kb/{kbID}/query", h.Query)
	r.Post("/v1/kb/{kbID}/hydrate", h.Hydrate)
	r.Post("/v1/kb/{kbID}/retrieve", h.Retrieve)
	return r
}
