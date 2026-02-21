package repository

import (
	"context"

	"ragtime-backend/internal/domain"
)

// Store defines the repository layer contract for chunking persistence.
type Store interface {
	InsertChunks(ctx context.Context, chunks []domain.Chunk) error
	DeleteChunksByDocumentVersion(ctx context.Context, documentVersionID string) error
	DeleteChunksByDocument(ctx context.Context, knowledgeBaseID, documentID string) error
	GetChunkByID(ctx context.Context, knowledgeBaseID, chunkID string) (*domain.Chunk, error)
	UpdateChunkEmbedding(ctx context.Context, knowledgeBaseID, chunkID, embeddingID string) error
	UpdateDocumentVersionStatus(ctx context.Context, versionID, status string, errorMessage *string) error
	GetLatestDocumentVersionForDocument(ctx context.Context, knowledgeBaseID, documentID string) (*DocumentVersionRef, error)
	ActivateDocumentVersion(ctx context.Context, versionID string) error
}

// DocumentVersionRef is the minimal version payload required to fetch stored bytes.
type DocumentVersionRef struct {
	DocumentVersionID string
	RawContentURI     string
}
