package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"ragtime-backend/internal/chunking"
)

// ChunkingHandler handles chunking preview endpoints.
type ChunkingHandler struct {
	service *chunking.Service
}

func NewChunkingHandler(service *chunking.Service) *ChunkingHandler {
	return &ChunkingHandler{service: service}
}

func (h *ChunkingHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/v1/chunking/preview" {
		writeError(w, http.StatusNotFound, "route not found")
		return
	}
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var payload chunkingPreviewRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	strategy := chunking.Strategy(strings.TrimSpace(payload.Strategy))
	languageHints, err := chunking.ParseLanguageHints(payload.LanguageHints)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	res, err := h.service.Chunk(chunking.Request{
		Text:          payload.Text,
		Strategy:      strategy,
		MaxRunes:      payload.MaxRunes,
		OverlapRunes:  payload.OverlapRunes,
		Separators:    payload.Separators,
		LanguageHints: languageHints,
	})
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, chunkingPreviewResponse{
		Strategy:     string(res.Strategy),
		MaxRunes:     res.MaxRunes,
		OverlapRunes: res.OverlapRunes,
		Separators:   res.Separators,
		Chunks:       res.Chunks,
	})
}

type chunkingPreviewRequest struct {
	Text          string   `json:"text"`
	Strategy      string   `json:"strategy"`
	MaxRunes      int      `json:"max_runes"`
	OverlapRunes  int      `json:"overlap_runes"`
	Separators    []string `json:"separators"`
	LanguageHints []string `json:"language_hints"`
}

type chunkingPreviewResponse struct {
	Strategy     string           `json:"strategy"`
	MaxRunes     int              `json:"max_runes"`
	OverlapRunes int              `json:"overlap_runes"`
	Separators   []string         `json:"separators,omitempty"`
	Chunks       []chunking.Chunk `json:"chunks"`
}
