package http

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	chunkcache "ragtime-backend/internal/chunking/cache"
	chunkrepo "ragtime-backend/internal/chunking/repository"
	chunkservice "ragtime-backend/internal/chunking/service"
	"ragtime-backend/internal/domain"
)

type repoStub struct {
	version *chunkrepo.DocumentVersionRef
}

func (s *repoStub) InsertChunks(context.Context, []domain.Chunk) error { return nil }
func (s *repoStub) DeleteChunksByDocumentVersion(context.Context, string) error {
	return nil
}
func (s *repoStub) DeleteChunksByDocument(context.Context, string, string) error {
	return nil
}
func (s *repoStub) GetChunkByID(context.Context, string, string) (*domain.Chunk, error) {
	return &domain.Chunk{
		ID:          "5fb0f664-92a2-4ced-b3a3-1fbeecf36d98",
		Content:     "alpha",
		ContentHash: "2c1743a391305fbf367df8e4f069f9f9",
		Metadata:    map[string]any{},
	}, nil
}
func (s *repoStub) UpdateChunkEmbedding(context.Context, string, string, string) error {
	return nil
}
func (s *repoStub) UpdateDocumentVersionStatus(context.Context, string, string, *string) error {
	return nil
}
func (s *repoStub) GetLatestDocumentVersionForDocument(context.Context, string, string) (*chunkrepo.DocumentVersionRef, error) {
	return s.version, nil
}
func (s *repoStub) ActivateDocumentVersion(context.Context, string) error { return nil }

type storeStub struct {
	content string
}

func (s *storeStub) Put(context.Context, string, io.Reader) (string, int64, error) {
	return "", 0, nil
}
func (s *storeStub) URIForKey(string) string { return "" }
func (s *storeStub) PresignPut(context.Context, string, string) (string, map[string]string, string, error) {
	return "", nil, "", nil
}
func (s *storeStub) Get(context.Context, string) (io.ReadCloser, int64, error) {
	return io.NopCloser(bytes.NewBufferString(s.content)), int64(len(s.content)), nil
}

func newTestHandler() *Handler {
	repo := &repoStub{
		version: &chunkrepo.DocumentVersionRef{
			DocumentVersionID: "2f93ec77-c97d-4a86-bd98-5fb99454bf95",
			RawContentURI:     "s3://bucket/kb/k1/documents/doc-1/v1.md",
		},
	}
	var cacheLayer chunkcache.Layer = repo
	service := chunkservice.New(cacheLayer, nil, nil, &storeStub{content: "alpha\n\nbeta"}, nil)
	return NewHandler(service)
}

func TestHandlerAllowsEmptyStrategy(t *testing.T) {
	h := NewRouter(newTestHandler().service)
	req := httptest.NewRequest(http.MethodPost, "/v1/kb/kb-1/documents/doc-1/chunking", bytes.NewReader([]byte(`{}`)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)

	if w.Code != http.StatusAccepted {
		t.Fatalf("expected status 202, got %d", w.Code)
	}
}

func TestHandlerRouteAndMethod(t *testing.T) {
	h := NewRouter(newTestHandler().service)
	req := httptest.NewRequest(http.MethodPost, "/v1/kb/kb-1/documents/doc-1/chunking", bytes.NewReader([]byte(`{"strategy":"recursive","max_runes":8}`)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)

	if w.Code != http.StatusAccepted {
		t.Fatalf("expected status 202, got %d", w.Code)
	}
}

func TestHandlerRouteNotFound(t *testing.T) {
	h := NewRouter(newTestHandler().service)
	req := httptest.NewRequest(http.MethodPost, "/v1/kb/kb-1/documents/doc-1/not-chunking", bytes.NewReader([]byte(`{"strategy":"recursive"}`)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", w.Code)
	}
}

func TestEmbedChunkRouteExists(t *testing.T) {
	h := NewRouter(newTestHandler().service)
	req := httptest.NewRequest(http.MethodPost, "/v1/kb/kb-1/chunks/chunk-1/embed", nil)
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status 503, got %d", w.Code)
	}
}
