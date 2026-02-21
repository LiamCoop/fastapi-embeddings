package repository

import (
	"context"

	"ragtime-backend/internal/retrieval"
)

// Store persists and searches retrieval-related records.
type Store interface {
	InsertRetrievalRequest(ctx context.Context, req retrieval.RetrievalRequestRecord) (*retrieval.RetrievalRequestRecord, error)
	UpdateRetrievalRequest(ctx context.Context, requestID string, resultCount int, latencyMS int64, emptyResult bool) error
	InsertRetrievalResults(ctx context.Context, results []retrieval.RetrievalResultRecord) error
	SearchSemantic(ctx context.Context, params retrieval.SearchParams) ([]retrieval.ScoredChunk, error)
	SearchLexical(ctx context.Context, params retrieval.SearchParams) ([]retrieval.ScoredChunk, error)
	GetChunksWithDocuments(ctx context.Context, chunkIDs []string) ([]retrieval.ChunkRecord, error)
}
