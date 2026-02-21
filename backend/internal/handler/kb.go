package handler

import (
	"net/http"
	"strings"

	chunkhttp "ragtime-backend/internal/chunking/http"
	chunkservice "ragtime-backend/internal/chunking/service"
	"ragtime-backend/internal/document"
	"ragtime-backend/internal/ingestion"
	retrievalservice "ragtime-backend/internal/retrieval/service"
)

// KBHandler routes KB-scoped requests to the appropriate handler.
type KBHandler struct {
	documents http.Handler
	chunking  http.Handler
	ingestion http.Handler
	retrieval http.Handler
}

func NewKBHandler(
	documentService *document.Service,
	chunkingService *chunkservice.Service,
	ingestionSvc *ingestion.Service,
	retrievalService *retrievalservice.Service,
) *KBHandler {
	return &KBHandler{
		documents: NewDocumentHandler(documentService),
		chunking:  http.HandlerFunc(chunkhttp.NewHandler(chunkingService).InitiateDocumentChunking),
		ingestion: NewIngestionHandler(ingestionSvc),
		retrieval: NewRetrievalHandler(retrievalService),
	}
}

func (h *KBHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.Trim(r.URL.Path, "/")
	parts := strings.Split(path, "/")

	if len(parts) < 4 || parts[0] != "v1" || parts[1] != "kb" {
		writeError(w, http.StatusNotFound, "route not found")
		return
	}

	switch parts[3] {
	case "documents":
		if len(parts) == 6 && parts[5] == "chunking" {
			h.chunking.ServeHTTP(w, r)
			return
		}
		h.documents.ServeHTTP(w, r)
	case "ingest":
		h.ingestion.ServeHTTP(w, r)
	case "retrieve":
		h.retrieval.ServeHTTP(w, r)
	default:
		writeError(w, http.StatusNotFound, "route not found")
	}
}
