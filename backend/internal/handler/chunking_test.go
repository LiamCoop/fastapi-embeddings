package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"ragtime-backend/internal/chunking"
)

func TestChunkingPreviewEndpoint(t *testing.T) {
	ch := make(chan chunking.DocumentRequest)
	service := chunking.NewService(nil, nil, ch)
	h := NewChunkingHandler(service)

	body := []byte(`{"text":"section1\n\nsection2","strategy":"recursive","max_runes":10,"overlap_runes":0}`)
	req := httptest.NewRequest(http.MethodPost, "/v1/chunking/preview", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var payload map[string]any
	if err := json.NewDecoder(w.Body).Decode(&payload); err != nil {
		t.Fatalf("invalid response json: %v", err)
	}

	chunks, ok := payload["chunks"].([]any)
	if !ok || len(chunks) == 0 {
		t.Fatalf("expected chunks in response")
	}
}

func TestChunkingPreviewEndpoint_InvalidLanguage(t *testing.T) {
	ch := make(chan chunking.DocumentRequest)
	service := chunking.NewService(nil, nil, ch)
	h := NewChunkingHandler(service)

	body := []byte(`{"text":"hello","max_runes":5,"language_hints":["kotlin"]}`)
	req := httptest.NewRequest(http.MethodPost, "/v1/chunking/preview", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}
