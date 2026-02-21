package retrieval

import (
	"errors"
	"time"
)

const (
	DefaultTopK         = 5
	DefaultHybridWeight = 0.7
	MaxTopK             = 50
)

var (
	ErrNilRepository        = errors.New("repository is required")
	ErrNilEmbedder          = errors.New("embedding client is required")
	ErrMissingKnowledgeBase = errors.New("knowledgebase_id is required")
	ErrMissingQuery         = errors.New("query is required")
	ErrInvalidTopK          = errors.New("top_k must be between 1 and 50")
	ErrInvalidHybridWeight  = errors.New("hybrid_weight must be between 0 and 1")
	ErrInvalidCreatedAfter  = errors.New("created_after must be before created_before")
)

type Filters struct {
	DocumentType  *string
	PathPrefix    *string
	Source        *string
	Tags          []string
	CreatedAfter  *time.Time
	CreatedBefore *time.Time
}

type Request struct {
	KnowledgeBaseID string
	Query           string
	TopK            int
	HybridWeight    float64
	HybridWeightSet bool
	Filters         Filters
}

type Score struct {
	Semantic float64 `json:"semantic"`
	Lexical  float64 `json:"lexical"`
	Final    float64 `json:"final"`
}

type Citation struct {
	DocumentID        string  `json:"document_id"`
	DocumentVersionID string  `json:"document_version_id"`
	Path              string  `json:"path"`
	Title             *string `json:"title,omitempty"`
	VersionNumber     int32   `json:"version_number"`
	ChunkSequence     int32   `json:"chunk_sequence"`
	StartRune         *int    `json:"start_rune,omitempty"`
	EndRune           *int    `json:"end_rune,omitempty"`
	RuneLength        *int    `json:"rune_length,omitempty"`
}

type Result struct {
	ChunkID           string         `json:"chunk_id"`
	DocumentID        string         `json:"document_id"`
	DocumentVersionID string         `json:"document_version_id"`
	DocumentPath      string         `json:"document_path"`
	DocumentTitle     *string        `json:"document_title,omitempty"`
	DocumentType      string         `json:"document_type"`
	Content           string         `json:"content"`
	Metadata          map[string]any `json:"metadata"`
	Scores            Score          `json:"scores"`
	Citation          Citation       `json:"citation"`
}

type Response struct {
	RequestID       string   `json:"request_id"`
	KnowledgeBaseID string   `json:"kb_id"`
	Query           string   `json:"query"`
	TopK            int      `json:"top_k"`
	HybridWeight    float64  `json:"hybrid_weight"`
	ResultCount     int      `json:"result_count"`
	LatencyMS       int64    `json:"latency_ms"`
	Results         []Result `json:"results"`
}

func ValidateRequest(req Request) error {
	if req.KnowledgeBaseID == "" {
		return ErrMissingKnowledgeBase
	}
	if req.Query == "" {
		return ErrMissingQuery
	}
	if req.TopK < 1 || req.TopK > MaxTopK {
		return ErrInvalidTopK
	}
	if req.HybridWeight < 0 || req.HybridWeight > 1 {
		return ErrInvalidHybridWeight
	}
	if req.Filters.CreatedAfter != nil && req.Filters.CreatedBefore != nil {
		if req.Filters.CreatedAfter.After(*req.Filters.CreatedBefore) {
			return ErrInvalidCreatedAfter
		}
	}
	return nil
}
