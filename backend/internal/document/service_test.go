package document

import (
	"context"
	"errors"
	"io"
	"testing"
	"time"

	chunkservice "ragtime-backend/internal/chunking/service"
	"ragtime-backend/internal/domain"
)

type fakeRepo struct {
	documents     map[string]*DocumentRecord
	versionCounts map[string]int32
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{
		documents:     map[string]*DocumentRecord{},
		versionCounts: map[string]int32{},
	}
}

func (r *fakeRepo) GetDocumentByKBPath(ctx context.Context, kbID, path string) (*DocumentRecord, error) {
	_ = ctx
	key := kbID + "|" + path
	doc, ok := r.documents[key]
	if !ok {
		return nil, nil
	}
	copy := *doc
	return &copy, nil
}

func (r *fakeRepo) InsertDocument(ctx context.Context, doc DocumentRecord) (*DocumentRecord, error) {
	_ = ctx
	key := doc.KnowledgeBaseID + "|" + doc.Path
	copy := doc
	r.documents[key] = &copy
	return &doc, nil
}

func (r *fakeRepo) UpdateDocument(ctx context.Context, doc DocumentRecord) (*DocumentRecord, error) {
	_ = ctx
	key := doc.KnowledgeBaseID + "|" + doc.Path
	copy := doc
	r.documents[key] = &copy
	return &doc, nil
}

func (r *fakeRepo) InsertDocumentVersion(ctx context.Context, version DocumentVersionRecord) (*DocumentVersionRecord, error) {
	_ = ctx
	r.versionCounts[version.DocumentID]++
	version.VersionNumber = r.versionCounts[version.DocumentID]
	return &version, nil
}

func (r *fakeRepo) UpdateDocumentVersionStatus(ctx context.Context, versionID, status string, errorMessage *string) error {
	_ = ctx
	_ = versionID
	_ = status
	_ = errorMessage
	return nil
}

func (r *fakeRepo) ActivateVersion(ctx context.Context, versionID string) error {
	_ = ctx
	_ = versionID
	return nil
}

type fakeStore struct {
	uriPrefix string
	putCalls  int
	lastKey   string
}

func (s *fakeStore) Put(ctx context.Context, key string, r io.Reader) (string, int64, error) {
	_ = ctx
	_, err := io.ReadAll(r)
	if err != nil {
		return "", 0, err
	}
	s.putCalls++
	s.lastKey = key
	return s.URIForKey(key), 0, nil
}

func (s *fakeStore) URIForKey(key string) string {
	return s.uriPrefix + key
}

func (s *fakeStore) PresignPut(ctx context.Context, key, contentType string) (string, map[string]string, string, error) {
	_ = ctx
	_ = key
	_ = contentType
	return "", nil, "", errors.New("not supported")
}

func (s *fakeStore) Get(ctx context.Context, uri string) (io.ReadCloser, int64, error) {
	_ = ctx
	_ = uri
	return nil, 0, errors.New("not supported")
}

func TestUploadCreatesNewVersion(t *testing.T) {
	repo := newFakeRepo()
	store := &fakeStore{uriPrefix: "file:///"}
	service := NewService(repo, store)
	service.now = func() time.Time { return time.Date(2026, 2, 2, 10, 0, 0, 0, time.UTC) }

	req := UploadRequest{
		KnowledgeBaseID: "kb-1",
		Path:            "docs/guide.md",
		FileContent:     []byte("hello"),
		FileName:        "guide.md",
		ContentType:     "text/markdown",
	}

	res, err := service.Upload(context.Background(), req)
	if err != nil {
		t.Fatalf("upload failed: %v", err)
	}
	if res.VersionNumber != 1 {
		t.Fatalf("expected version 1, got %d", res.VersionNumber)
	}
	if res.ProcessingStatus != string(domain.StatusStored) {
		t.Fatalf("expected status STORED, got %s", res.ProcessingStatus)
	}
	if store.putCalls != 1 {
		t.Fatalf("expected store Put to be called once")
	}

	res2, err := service.Upload(context.Background(), req)
	if err != nil {
		t.Fatalf("second upload failed: %v", err)
	}
	if res2.VersionNumber != 2 {
		t.Fatalf("expected version 2, got %d", res2.VersionNumber)
	}
	if res2.DocumentID != res.DocumentID {
		t.Fatalf("expected same document id on update")
	}
}

func TestUploadUnsupportedTypeSkipsProcessing(t *testing.T) {
	repo := newFakeRepo()
	store := &fakeStore{uriPrefix: "file:///"}
	service := NewService(repo, store)
	service.now = func() time.Time { return time.Date(2026, 2, 2, 10, 0, 0, 0, time.UTC) }

	docType := "pdf"
	res, err := service.Upload(context.Background(), UploadRequest{
		KnowledgeBaseID: "kb-1",
		Path:            "docs/file.pdf",
		DocumentType:    &docType,
		FileContent:     []byte("pdf"),
		FileName:        "file.pdf",
		ContentType:     "application/pdf",
	})
	if err != nil {
		t.Fatalf("upload failed: %v", err)
	}
	if res.ProcessingStatus != string(domain.StatusSkippedUnsupported) {
		t.Fatalf("expected status SKIPPED_UNSUPPORTED, got %s", res.ProcessingStatus)
	}
}

func TestUploadWithRawContentURISkipsPut(t *testing.T) {
	repo := newFakeRepo()
	store := &fakeStore{uriPrefix: "file:///"}
	service := NewService(repo, store)
	service.now = func() time.Time { return time.Date(2026, 2, 2, 10, 0, 0, 0, time.UTC) }

	uri := "s3://bucket/path"
	res, err := service.Upload(context.Background(), UploadRequest{
		KnowledgeBaseID: "kb-1",
		Path:            "docs/remote.md",
		RawContentURI:   &uri,
		ContentType:     "text/markdown",
	})
	if err != nil {
		t.Fatalf("upload failed: %v", err)
	}
	if res.RawContentURI != uri {
		t.Fatalf("expected raw_content_uri to be preserved")
	}
	if store.putCalls != 0 {
		t.Fatalf("expected store Put to be skipped")
	}
}

func TestUploadEnqueuesChunkingRequest(t *testing.T) {
	repo := newFakeRepo()
	store := &fakeStore{uriPrefix: "file:///"}
	ch := make(chan chunkservice.DocumentRequest, 1)
	service := NewServiceWithChunking(repo, store, ch)
	service.now = func() time.Time { return time.Date(2026, 2, 2, 10, 0, 0, 0, time.UTC) }

	req := UploadRequest{
		KnowledgeBaseID: "kb-1",
		Path:            "docs/guide.md",
		FileContent:     []byte("hello world"),
		FileName:        "guide.md",
		ContentType:     "text/markdown",
	}

	res, err := service.Upload(context.Background(), req)
	if err != nil {
		t.Fatalf("upload failed: %v", err)
	}

	select {
	case msg := <-ch:
		if msg.KnowledgeBaseID != req.KnowledgeBaseID {
			t.Fatalf("expected knowledge base id %s, got %s", req.KnowledgeBaseID, msg.KnowledgeBaseID)
		}
		if msg.DocumentVersionID != res.DocumentVersionID {
			t.Fatalf("expected version id %s, got %s", res.DocumentVersionID, msg.DocumentVersionID)
		}
		if msg.Content != string(req.FileContent) {
			t.Fatalf("expected content to match upload payload")
		}
	default:
		t.Fatalf("expected chunking request to be enqueued")
	}
}
