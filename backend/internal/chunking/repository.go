package chunking

import (
	"context"

	"ragtime-backend/internal/domain"
)

// Repository persists chunks and updates document version status.
type Repository interface {
	InsertChunks(ctx context.Context, chunks []domain.Chunk) error
	DeleteChunksByDocumentVersion(ctx context.Context, documentVersionID string) error
	UpdateDocumentVersionStatus(ctx context.Context, versionID, status string, errorMessage *string) error
	GetLatestDocumentVersionForDocument(ctx context.Context, knowledgeBaseID, documentID string) (*DocumentVersionRef, error)
}

// CacheLayer is the service-facing data access abstraction.
// Current no-op cache implementation delegates all calls to the repository layer.
type CacheLayer interface {
	InsertChunks(ctx context.Context, chunks []domain.Chunk) error
	DeleteChunksByDocumentVersion(ctx context.Context, documentVersionID string) error
	UpdateDocumentVersionStatus(ctx context.Context, versionID, status string, errorMessage *string) error
	GetLatestDocumentVersionForDocument(ctx context.Context, knowledgeBaseID, documentID string) (*DocumentVersionRef, error)
}

// DocumentVersionRef is the minimal document version data required to chunk content from object storage.
type DocumentVersionRef struct {
	DocumentVersionID string
	RawContentURI     string
}
