package cache

import (
	"context"

	"ragtime-backend/internal/chunking/repository"
	"ragtime-backend/internal/domain"
)

// Layer is the cache abstraction between service and repository.
type Layer interface {
	InsertChunks(ctx context.Context, chunks []domain.Chunk) error
	DeleteChunksByDocumentVersion(ctx context.Context, documentVersionID string) error
	DeleteChunksByDocument(ctx context.Context, knowledgeBaseID, documentID string) error
	GetChunkByID(ctx context.Context, knowledgeBaseID, chunkID string) (*domain.Chunk, error)
	UpdateChunkEmbedding(ctx context.Context, knowledgeBaseID, chunkID, embeddingID string) error
	UpdateDocumentVersionStatus(ctx context.Context, versionID, status string, errorMessage *string) error
	GetLatestDocumentVersionForDocument(ctx context.Context, knowledgeBaseID, documentID string) (*repository.DocumentVersionRef, error)
	ActivateDocumentVersion(ctx context.Context, versionID string) error
}
