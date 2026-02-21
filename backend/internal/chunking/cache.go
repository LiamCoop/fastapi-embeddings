package chunking

import (
	"context"

	"ragtime-backend/internal/domain"
)

// NoopCacheLayer is a pass-through cache layer for MVP.
// It preserves layer boundaries while introducing no caching behavior yet.
type NoopCacheLayer struct {
	repo Repository
}

func NewNoopCacheLayer(repo Repository) *NoopCacheLayer {
	return &NoopCacheLayer{repo: repo}
}

func (c *NoopCacheLayer) InsertChunks(ctx context.Context, chunks []domain.Chunk) error {
	return c.repo.InsertChunks(ctx, chunks)
}

func (c *NoopCacheLayer) DeleteChunksByDocumentVersion(ctx context.Context, documentVersionID string) error {
	return c.repo.DeleteChunksByDocumentVersion(ctx, documentVersionID)
}

func (c *NoopCacheLayer) UpdateDocumentVersionStatus(ctx context.Context, versionID, status string, errorMessage *string) error {
	return c.repo.UpdateDocumentVersionStatus(ctx, versionID, status, errorMessage)
}

func (c *NoopCacheLayer) GetLatestDocumentVersionForDocument(ctx context.Context, knowledgeBaseID, documentID string) (*DocumentVersionRef, error) {
	return c.repo.GetLatestDocumentVersionForDocument(ctx, knowledgeBaseID, documentID)
}

var _ CacheLayer = (*NoopCacheLayer)(nil)
