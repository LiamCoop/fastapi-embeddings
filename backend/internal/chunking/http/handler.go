package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"ragtime-backend/internal/chunking"
	chunkservice "ragtime-backend/internal/chunking/service"
	"ragtime-backend/internal/logger"
)

// Handler handles document-scoped chunking initiation requests.
type Handler struct {
	service *chunkservice.Service
}

func NewHandler(service *chunkservice.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) InitiateDocumentChunking(w http.ResponseWriter, r *http.Request) {
	kbID := strings.TrimSpace(chi.URLParam(r, "kbID"))
	documentID := strings.TrimSpace(chi.URLParam(r, "documentID"))
	logger.Info("chunking request received", "kb_id", kbID, "document_id", documentID, "method", r.Method, "path", r.URL.Path)

	var payload request
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&payload); err != nil {
		logger.Warn("chunking request invalid json", "kb_id", kbID, "document_id", documentID, "error", err)
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	payload.Strategy = strings.TrimSpace(payload.Strategy)

	languageHints, err := chunkservice.ParseLanguageHints(payload.LanguageHints)
	if err != nil {
		logger.Warn("chunking request invalid language hints", "kb_id", kbID, "document_id", documentID, "error", err)
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if kbID == "" || documentID == "" {
		logger.Warn("chunking request missing route params", "kb_id", kbID, "document_id", documentID)
		writeError(w, http.StatusBadRequest, "kbID and documentID are required")
		return
	}

	logger.Info(
		"chunking request calling service",
		"kb_id", kbID,
		"document_id", documentID,
		"strategy", payload.Strategy,
		"max_runes", payload.MaxRunes,
		"overlap_runes", payload.OverlapRunes,
		"separators_count", len(payload.Separators),
		"language_hints_count", len(languageHints),
	)

	res, err := h.service.InitiateDocumentChunking(r.Context(), chunkservice.InitiateRequest{
		KnowledgeBaseID: kbID,
		DocumentID:      documentID,
		Strategy:        chunking.Strategy(payload.Strategy),
		MaxRunes:        payload.MaxRunes,
		OverlapRunes:    payload.OverlapRunes,
		Separators:      payload.Separators,
		LanguageHints:   languageHints,
	})
	if err != nil {
		switch {
		case errors.Is(err, chunkservice.ErrDocumentNotFound):
			logger.Warn("chunking request document not found", "kb_id", kbID, "document_id", documentID, "error", err)
			writeError(w, http.StatusNotFound, err.Error())
		default:
			logger.Error("chunking request failed", "kb_id", kbID, "document_id", documentID, "error", err)
			writeError(w, http.StatusBadRequest, err.Error())
		}
		return
	}

	logger.Info(
		"chunking request accepted",
		"kb_id", kbID,
		"document_id", res.DocumentID,
		"document_version_id", res.DocumentVersionID,
		"strategy", res.Strategy,
		"chunk_count", res.ChunkCount,
	)
	writeJSON(w, http.StatusAccepted, res)
}

func (h *Handler) EmbedChunkByID(w http.ResponseWriter, r *http.Request) {
	kbID := strings.TrimSpace(chi.URLParam(r, "kbID"))
	chunkID := strings.TrimSpace(chi.URLParam(r, "chunkID"))
	logger.Info("chunk re-embed request received", "kb_id", kbID, "chunk_id", chunkID, "method", r.Method, "path", r.URL.Path)

	if kbID == "" || chunkID == "" {
		writeError(w, http.StatusBadRequest, "kbID and chunkID are required")
		return
	}

	res, err := h.service.EmbedChunkByID(r.Context(), kbID, chunkID)
	if err != nil {
		switch {
		case errors.Is(err, chunkservice.ErrChunkNotFound):
			writeError(w, http.StatusNotFound, err.Error())
		case errors.Is(err, chunkservice.ErrEmbedderUnavailable):
			writeError(w, http.StatusServiceUnavailable, err.Error())
		default:
			writeError(w, http.StatusBadRequest, err.Error())
		}
		return
	}

	writeJSON(w, http.StatusOK, res)
}

type request struct {
	Strategy      string   `json:"strategy"`
	MaxRunes      int      `json:"max_runes"`
	OverlapRunes  int      `json:"overlap_runes"`
	Separators    []string `json:"separators"`
	LanguageHints []string `json:"language_hints"`
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, message string) {
	if status >= 500 {
		message = "internal server error"
	}
	writeJSON(w, status, map[string]string{"error": message})
}
