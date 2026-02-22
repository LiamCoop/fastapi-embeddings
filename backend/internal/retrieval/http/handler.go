package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"

	"ragtime-backend/internal/logger"
	"ragtime-backend/internal/retrieval"
	retrievalservice "ragtime-backend/internal/retrieval/service"
)

// Handler handles retrieval requests.
type Handler struct {
	service *retrievalservice.Service
	metrics *retrievalMetrics
}

type retrievalMetrics struct {
	requests    metric.Int64Counter
	latencyMS   metric.Float64Histogram
	resultCount metric.Int64Histogram
}

func NewHandler(service *retrievalservice.Service) *Handler {
	h := &Handler{service: service}
	h.metrics = newRetrievalMetrics()
	return h
}

// Retrieve preserves the existing route while serving the updated query contract.
func (h *Handler) Retrieve(w http.ResponseWriter, r *http.Request) {
	h.Query(w, r)
}

func (h *Handler) Query(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	statusCode := http.StatusOK
	outcome := "success"
	resultCount := int64(0)
	defer func() {
		h.recordMetrics(r, "/v1/kb/{kbID}/query", start, statusCode, outcome, resultCount)
	}()

	kbID := strings.TrimSpace(chi.URLParam(r, "kbID"))
	if kbID == "" {
		statusCode = http.StatusBadRequest
		outcome = "client_error"
		writeError(w, http.StatusBadRequest, "kbID is required")
		return
	}

	var payload queryRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&payload); err != nil {
		statusCode = http.StatusBadRequest
		outcome = "client_error"
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	req, err := buildRetrievalRequest(kbID, payload)
	if err != nil {
		statusCode = http.StatusBadRequest
		outcome = "client_error"
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	res, err := h.service.Retrieve(r.Context(), req)
	if err != nil {
		if isRetrievalClientError(err) {
			statusCode = http.StatusBadRequest
			outcome = "client_error"
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		statusCode = http.StatusInternalServerError
		outcome = "server_error"
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	statusCode = http.StatusOK
	resultCount = int64(res.ResultCount)
	writeJSON(w, http.StatusOK, res)
}

func (h *Handler) Hydrate(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	statusCode := http.StatusOK
	outcome := "success"
	resultCount := int64(0)
	defer func() {
		h.recordMetrics(r, "/v1/kb/{kbID}/hydrate", start, statusCode, outcome, resultCount)
	}()

	kbID := strings.TrimSpace(chi.URLParam(r, "kbID"))
	if kbID == "" {
		statusCode = http.StatusBadRequest
		outcome = "client_error"
		writeError(w, http.StatusBadRequest, "kbID is required")
		return
	}

	var payload hydrateRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&payload); err != nil {
		statusCode = http.StatusBadRequest
		outcome = "client_error"
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	req := retrieval.HydrateRequest{
		KnowledgeBaseID: kbID,
		ChunkIDs:        payload.ChunkIDs,
		AdjacentBefore:  payload.AdjacentBefore,
		AdjacentAfter:   payload.AdjacentAfter,
	}

	res, err := h.service.Hydrate(r.Context(), req)
	if err != nil {
		if isRetrievalClientError(err) {
			statusCode = http.StatusBadRequest
			outcome = "client_error"
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		statusCode = http.StatusInternalServerError
		outcome = "server_error"
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	statusCode = http.StatusOK
	resultCount = int64(res.ChunkCount)
	writeJSON(w, http.StatusOK, res)
}

func (h *Handler) recordMetrics(r *http.Request, route string, startedAt time.Time, statusCode int, outcome string, resultCount int64) {
	if h.metrics == nil {
		return
	}
	attrs := metric.WithAttributes(
		attribute.String("http.route", route),
		attribute.Int("http.status_code", statusCode),
		attribute.String("retrieval.outcome", outcome),
	)
	h.metrics.requests.Add(r.Context(), 1, attrs)
	h.metrics.latencyMS.Record(r.Context(), float64(time.Since(startedAt).Milliseconds()), attrs)
	h.metrics.resultCount.Record(r.Context(), resultCount, attrs)
}

func newRetrievalMetrics() *retrievalMetrics {
	meter := otel.Meter("ragtime-backend/retrieval/http")

	requests, err := meter.Int64Counter(
		"ragtime.retrieval.requests",
		metric.WithDescription("Total retrieval HTTP requests."),
		metric.WithUnit("{request}"),
	)
	if err != nil {
		logger.Warn("failed to initialize retrieval requests metric", "error", err)
		return nil
	}

	latencyMS, err := meter.Float64Histogram(
		"ragtime.retrieval.latency",
		metric.WithDescription("Retrieval HTTP latency in milliseconds."),
		metric.WithUnit("ms"),
	)
	if err != nil {
		logger.Warn("failed to initialize retrieval latency metric", "error", err)
		return nil
	}

	resultCount, err := meter.Int64Histogram(
		"ragtime.retrieval.result_count",
		metric.WithDescription("Number of results returned by retrieval requests."),
		metric.WithUnit("{result}"),
	)
	if err != nil {
		logger.Warn("failed to initialize retrieval result_count metric", "error", err)
		return nil
	}

	return &retrievalMetrics{
		requests:    requests,
		latencyMS:   latencyMS,
		resultCount: resultCount,
	}
}

type queryRequest struct {
	Query            string       `json:"query"`
	TopK             *int         `json:"top_k"`
	HybridWeight     *float64     `json:"hybrid_weight"`
	RetrievalProfile *string      `json:"retrieval_profile"`
	SemanticWeight   *float64     `json:"semantic_weight"`
	Debug            bool         `json:"debug"`
	Filters          *filtersJSON `json:"filters"`
}

type hydrateRequest struct {
	ChunkIDs       []string `json:"chunk_ids"`
	AdjacentBefore int      `json:"adjacent_before"`
	AdjacentAfter  int      `json:"adjacent_after"`
}

type filtersJSON struct {
	PathPrefix    *string  `json:"path_prefix"`
	DocumentType  *string  `json:"document_type"`
	Source        *string  `json:"source"`
	Tags          []string `json:"tags"`
	CreatedAfter  *string  `json:"created_after"`
	CreatedBefore *string  `json:"created_before"`
	UpdatedAfter  *string  `json:"updated_after"`
}

func buildRetrievalRequest(kbID string, payload queryRequest) (retrieval.Request, error) {
	req := retrieval.Request{
		KnowledgeBaseID: kbID,
		Query:           strings.TrimSpace(payload.Query),
		Debug:           payload.Debug,
	}

	if payload.TopK != nil {
		req.TopK = *payload.TopK
	}
	if payload.HybridWeight != nil {
		req.HybridWeight = *payload.HybridWeight
		req.HybridWeightSet = true
	}
	if payload.RetrievalProfile != nil {
		req.RetrievalProfile = strings.TrimSpace(*payload.RetrievalProfile)
	}
	if payload.SemanticWeight != nil {
		req.SemanticWeight = *payload.SemanticWeight
		req.SemanticWeightSet = true
		req.HybridWeight = *payload.SemanticWeight
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

	createdAfter := payload.Filters.CreatedAfter
	if payload.Filters.UpdatedAfter != nil && strings.TrimSpace(*payload.Filters.UpdatedAfter) != "" {
		createdAfter = payload.Filters.UpdatedAfter
	}

	if createdAfter != nil && *createdAfter != "" {
		parsed, err := time.Parse(time.RFC3339, *createdAfter)
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
		errors.Is(err, retrieval.ErrInvalidProfile) ||
		errors.Is(err, retrieval.ErrInvalidCreatedAfter) ||
		errors.Is(err, retrieval.ErrMissingChunkIDs) ||
		errors.Is(err, retrieval.ErrTooManyChunkIDs) ||
		errors.Is(err, retrieval.ErrInvalidAdjacentRange)
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
