package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	chunkservice "ragtime-backend/internal/chunking/service"
)

func NewRouter(service *chunkservice.Service) http.Handler {
	r := chi.NewRouter()
	h := NewHandler(service)
	r.Post("/v1/kb/{kbID}/documents/{documentID}/chunking", h.InitiateDocumentChunking)
	r.Post("/v1/kb/{kbID}/chunks/{chunkID}/embed", h.EmbedChunkByID)
	return r
}
