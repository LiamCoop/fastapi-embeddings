package cache

import (
	"context"

	"ragtime-backend/internal/chunking/repository"
	"ragtime-backend/internal/domain"
)

// NoopLayer currently performs no caching and delegates to repository.
type NoopLayer struct {
	store repository.Store
}

func NewNoopLayer(store repository.Store) *NoopLayer {
	return &NoopLayer{store: store}
}

func (c *NoopLayer) InsertChunks(ctx context.Context, chunks []domain.Chunk) error {
	return c.store.InsertChunks(ctx, chunks)
}

func (c *NoopLayer) DeleteChunksByDocumentVersion(ctx context.Context, documentVersionID string) error {
	return c.store.DeleteChunksByDocumentVersion(ctx, documentVersionID)
}

func (c *NoopLayer) DeleteChunksByDocument(ctx context.Context, knowledgeBaseID, documentID string) error {
	return c.store.DeleteChunksByDocument(ctx, knowledgeBaseID, documentID)
}

func (c *NoopLayer) GetChunkByID(ctx context.Context, knowledgeBaseID, chunkID string) (*domain.Chunk, error) {
	return c.store.GetChunkByID(ctx, knowledgeBaseID, chunkID)
}

func (c *NoopLayer) UpdateChunkEmbedding(ctx context.Context, knowledgeBaseID, chunkID, embeddingID string) error {
	return c.store.UpdateChunkEmbedding(ctx, knowledgeBaseID, chunkID, embeddingID)
}

func (c *NoopLayer) UpdateDocumentVersionStatus(ctx context.Context, versionID, status string, errorMessage *string) error {
	return c.store.UpdateDocumentVersionStatus(ctx, versionID, status, errorMessage)
}

func (c *NoopLayer) GetLatestDocumentVersionForDocument(ctx context.Context, knowledgeBaseID, documentID string) (*repository.DocumentVersionRef, error) {
	return c.store.GetLatestDocumentVersionForDocument(ctx, knowledgeBaseID, documentID)
}

func (c *NoopLayer) ActivateDocumentVersion(ctx context.Context, versionID string) error {
	return c.store.ActivateDocumentVersion(ctx, versionID)
}

var _ Layer = (*NoopLayer)(nil)
