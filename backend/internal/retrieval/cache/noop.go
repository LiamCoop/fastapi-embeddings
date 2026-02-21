package cache

import (
	"context"

	"ragtime-backend/internal/retrieval"
	"ragtime-backend/internal/retrieval/repository"
)

// NoopLayer currently performs no caching and delegates directly to repository.
type NoopLayer struct {
	store repository.Store
}

func NewNoopLayer(store repository.Store) *NoopLayer {
	return &NoopLayer{store: store}
}

func (c *NoopLayer) InsertRetrievalRequest(ctx context.Context, req retrieval.RetrievalRequestRecord) (*retrieval.RetrievalRequestRecord, error) {
	return c.store.InsertRetrievalRequest(ctx, req)
}

func (c *NoopLayer) UpdateRetrievalRequest(ctx context.Context, requestID string, resultCount int, latencyMS int64, emptyResult bool) error {
	return c.store.UpdateRetrievalRequest(ctx, requestID, resultCount, latencyMS, emptyResult)
}

func (c *NoopLayer) InsertRetrievalResults(ctx context.Context, results []retrieval.RetrievalResultRecord) error {
	return c.store.InsertRetrievalResults(ctx, results)
}

func (c *NoopLayer) SearchSemantic(ctx context.Context, params retrieval.SearchParams) ([]retrieval.ScoredChunk, error) {
	return c.store.SearchSemantic(ctx, params)
}

func (c *NoopLayer) SearchLexical(ctx context.Context, params retrieval.SearchParams) ([]retrieval.ScoredChunk, error) {
	return c.store.SearchLexical(ctx, params)
}

func (c *NoopLayer) GetChunksWithDocuments(ctx context.Context, chunkIDs []string) ([]retrieval.ChunkRecord, error) {
	return c.store.GetChunksWithDocuments(ctx, chunkIDs)
}

var _ Layer = (*NoopLayer)(nil)
