package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"ragtime-backend/internal/document"
)

type presignStore struct {
	lastKey         string
	lastContentType string
}

func (s *presignStore) Put(ctx context.Context, key string, r io.Reader) (string, int64, error) {
	_ = ctx
	_ = key
	_, _ = io.ReadAll(r)
	return "", 0, nil
}

func (s *presignStore) URIForKey(key string) string {
	return "s3://bucket/" + key
}

func (s *presignStore) PresignPut(ctx context.Context, key, contentType string) (string, map[string]string, string, error) {
	_ = ctx
	s.lastKey = key
	s.lastContentType = contentType
	return "https://signed.example/upload", map[string]string{"Content-Type": contentType}, s.URIForKey(key), nil
}

func (s *presignStore) Get(ctx context.Context, uri string) (io.ReadCloser, int64, error) {
	_ = ctx
	_ = uri
	return nil, 0, nil
}

func TestPresignEndpoint(t *testing.T) {
	store := &presignStore{}
	service := document.NewService(nil, store)
	h := NewDocumentHandler(service)

	body := []byte(`{"file_name":"guide.md","content_type":"text/markdown"}`)
	req := httptest.NewRequest(http.MethodPost, "/v1/kb/kb-1/documents/presign", bytes.NewReader(body))
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

	if payload["upload_url"] == "" {
		t.Fatalf("expected upload_url in response")
	}
	if payload["raw_content_uri"] == "" {
		t.Fatalf("expected raw_content_uri in response")
	}
	if store.lastContentType != "text/markdown" {
		t.Fatalf("expected content type to be passed to presigner")
	}
	if store.lastKey == "" {
		t.Fatalf("expected presign to receive a key")
	}
}

func TestPresignEndpointRequiresFileName(t *testing.T) {
	store := &presignStore{}
	service := document.NewService(nil, store)
	h := NewDocumentHandler(service)

	req := httptest.NewRequest(http.MethodPost, "/v1/kb/kb-1/documents/presign", bytes.NewReader([]byte(`{}`)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}
