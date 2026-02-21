package cache

import (
	"context"

	"ragtime-backend/internal/retrieval"
	"ragtime-backend/internal/retrieval/repository"
)

// Layer is the service-facing cache abstraction for retrieval data access.
type Layer interface {
	InsertRetrievalRequest(ctx context.Context, req retrieval.RetrievalRequestRecord) (*retrieval.RetrievalRequestRecord, error)
	UpdateRetrievalRequest(ctx context.Context, requestID string, resultCount int, latencyMS int64, emptyResult bool) error
	InsertRetrievalResults(ctx context.Context, results []retrieval.RetrievalResultRecord) error
	SearchSemantic(ctx context.Context, params retrieval.SearchParams) ([]retrieval.ScoredChunk, error)
	SearchLexical(ctx context.Context, params retrieval.SearchParams) ([]retrieval.ScoredChunk, error)
	GetChunksWithDocuments(ctx context.Context, chunkIDs []string) ([]retrieval.ChunkRecord, error)
}

var _ Layer = (repository.Store)(nil)
