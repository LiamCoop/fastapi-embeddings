package retrieval

import (
	"context"
	"time"
)

type Repository interface {
	InsertRetrievalRequest(ctx context.Context, req RetrievalRequestRecord) (*RetrievalRequestRecord, error)
	UpdateRetrievalRequest(ctx context.Context, requestID string, resultCount int, latencyMS int64, emptyResult bool) error
	InsertRetrievalResults(ctx context.Context, results []RetrievalResultRecord) error
	SearchSemantic(ctx context.Context, params SearchParams) ([]ScoredChunk, error)
	SearchLexical(ctx context.Context, params SearchParams) ([]ScoredChunk, error)
	GetChunksWithDocuments(ctx context.Context, chunkIDs []string) ([]ChunkRecord, error)
	GetChunksWithDocumentsForKB(ctx context.Context, knowledgeBaseID string, chunkIDs []string) ([]ChunkRecord, error)
	GetChunksByDocumentVersionRange(ctx context.Context, documentVersionID string, startSeq int32, endSeq int32) ([]ChunkRecord, error)
}

type RetrievalRequestRecord struct {
	ID            string
	KnowledgeBase string
	Query         string
	Filters       map[string]any
	TopK          int
	HybridWeight  float64
	ResultCount   int
	LatencyMS     int64
	EmptyResult   bool
	CreatedAt     time.Time
}

type RetrievalResultRecord struct {
	ID                 string
	RetrievalRequestID string
	ChunkID            string
	Rank               int
	SemanticScore      float64
	LexicalScore       float64
	FinalScore         float64
	CreatedAt          time.Time
}

type SearchParams struct {
	KnowledgeBaseID string
	Query           string
	QueryVector     []float32
	VectorDimension int
	DocumentType    *string
	PathPrefix      *string
	Source          *string
	TagsFilter      map[string]any
	CreatedAfter    *time.Time
	CreatedBefore   *time.Time
	Limit           int
}

type ScoredChunk struct {
	ChunkID string
	Score   float64
}

type ChunkRecord struct {
	ChunkID           string
	DocumentID        string
	DocumentVersionID string
	DocumentPath      string
	DocumentTitle     *string
	DocumentType      string
	Content           string
	Metadata          map[string]any
	VersionNumber     int32
	SequenceNumber    int32
	SourceMetadata    map[string]any
}
