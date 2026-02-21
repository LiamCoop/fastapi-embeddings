package handler

import (
	"net/http"

	chunkhttp "ragtime-backend/internal/chunking/http"
	chunkservice "ragtime-backend/internal/chunking/service"
)

// NewRouter creates and configures the HTTP router with all application routes.
func NewRouter(chunkingSvc *chunkservice.Service) http.Handler {
	return chunkhttp.NewRouter(chunkingSvc)
}
