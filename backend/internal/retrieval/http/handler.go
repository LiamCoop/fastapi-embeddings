package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"ragtime-backend/internal/retrieval"
	retrievalservice "ragtime-backend/internal/retrieval/service"
)

// Handler handles retrieval requests.
type Handler struct {
	service *retrievalservice.Service
}

func NewHandler(service *retrievalservice.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Retrieve(w http.ResponseWriter, r *http.Request) {
	kbID := strings.TrimSpace(chi.URLParam(r, "kbID"))
	if kbID == "" {
		writeError(w, http.StatusBadRequest, "kbID is required")
		return
	}

	var payload request
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	req, err := buildRetrievalRequest(kbID, payload)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	res, err := h.service.Retrieve(r.Context(), req)
	if err != nil {
		if isRetrievalClientError(err) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, res)
}

type request struct {
	Query        string       `json:"query"`
	TopK         *int         `json:"top_k"`
	HybridWeight *float64     `json:"hybrid_weight"`
	Filters      *filtersJSON `json:"filters"`
}

type filtersJSON struct {
	PathPrefix    *string  `json:"path_prefix"`
	DocumentType  *string  `json:"document_type"`
	Source        *string  `json:"source"`
	Tags          []string `json:"tags"`
	CreatedAfter  *string  `json:"created_after"`
	CreatedBefore *string  `json:"created_before"`
}

func buildRetrievalRequest(kbID string, payload request) (retrieval.Request, error) {
	req := retrieval.Request{
		KnowledgeBaseID: kbID,
		Query:           strings.TrimSpace(payload.Query),
	}

	if payload.TopK != nil {
		req.TopK = *payload.TopK
	}
	if payload.HybridWeight != nil {
		req.HybridWeight = *payload.HybridWeight
		req.HybridWeightSet = true
	}

	if payload.Filters == nil {
		return req, nil
	}

	filters := retrieval.Filters{
		PathPrefix:   payload.Filters.PathPrefix,
		DocumentType: payload.Filters.DocumentType,
		Source:       payload.Filters.Source,
		Tags:         payload.Filters.Tags,
	}

	if payload.Filters.CreatedAfter != nil && *payload.Filters.CreatedAfter != "" {
		parsed, err := time.Parse(time.RFC3339, *payload.Filters.CreatedAfter)
		if err != nil {
			return req, errors.New("created_after must be RFC3339")
		}
		filters.CreatedAfter = &parsed
	}
	if payload.Filters.CreatedBefore != nil && *payload.Filters.CreatedBefore != "" {
		parsed, err := time.Parse(time.RFC3339, *payload.Filters.CreatedBefore)
		if err != nil {
			return req, errors.New("created_before must be RFC3339")
		}
		filters.CreatedBefore = &parsed
	}

	req.Filters = filters
	return req, nil
}

func isRetrievalClientError(err error) bool {
	return errors.Is(err, retrieval.ErrMissingKnowledgeBase) ||
		errors.Is(err, retrieval.ErrMissingQuery) ||
		errors.Is(err, retrieval.ErrInvalidTopK) ||
		errors.Is(err, retrieval.ErrInvalidHybridWeight) ||
		errors.Is(err, retrieval.ErrInvalidCreatedAfter)
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
