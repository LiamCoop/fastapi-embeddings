package retrieval

import (
	"errors"
	"strings"
	"time"
)

const (
	DefaultTopK              = 5
	DefaultHybridWeight      = 0.7
	DefaultRetrievalProfile  = "auto"
	RetrievalProfileAuto     = "auto"
	RetrievalProfileExact    = "exact"
	RetrievalProfileBalanced = "balanced"
	RetrievalProfileSemantic = "semantic"
	MaxTopK                  = 50
	MaxHydrateChunkIDs       = 100
)

var (
	ErrNilRepository        = errors.New("repository is required")
	ErrNilEmbedder          = errors.New("embedding client is required")
	ErrMissingKnowledgeBase = errors.New("knowledgebase_id is required")
	ErrMissingQuery         = errors.New("query is required")
	ErrInvalidTopK          = errors.New("top_k must be between 1 and 50")
	ErrInvalidHybridWeight  = errors.New("hybrid_weight must be between 0 and 1")
	ErrInvalidProfile       = errors.New("retrieval_profile must be one of: auto, exact, balanced, semantic")
	ErrInvalidCreatedAfter  = errors.New("created_after must be before created_before")
	ErrMissingChunkIDs      = errors.New("chunk_ids is required")
	ErrTooManyChunkIDs      = errors.New("chunk_ids exceeds maximum of 100")
	ErrInvalidAdjacentRange = errors.New("adjacent_before and adjacent_after must be between 0 and 10")
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
	KnowledgeBaseID   string
	Query             string
	TopK              int
	HybridWeight      float64
	HybridWeightSet   bool
	RetrievalProfile  string
	SemanticWeight    float64
	SemanticWeightSet bool
	Debug             bool
	Filters           Filters
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
	SourceURI         string         `json:"source_uri"`
	Title             *string        `json:"title,omitempty"`
	SectionPath       []string       `json:"section_path,omitempty"`
	Text              string         `json:"text"`
	Score             float64        `json:"score"`
	ScoreDetail       Score          `json:"score_detail"`
	Offsets           *Offsets       `json:"offsets,omitempty"`
}

type Response struct {
	RequestID       string         `json:"request_id"`
	QueryID         string         `json:"query_id"`
	IndexVersion    string         `json:"index_version"`
	KnowledgeBaseID string         `json:"kb_id"`
	Query           string         `json:"query"`
	TopK            int            `json:"top_k"`
	HybridWeight    float64        `json:"hybrid_weight"`
	ResultCount     int            `json:"result_count"`
	LatencyMS       int64          `json:"latency_ms"`
	Results         []Result       `json:"results"`
	Passages        []Result       `json:"passages"`
	Debug           *DebugMetadata `json:"debug,omitempty"`
}

type Offsets struct {
	StartRune  *int `json:"start_rune,omitempty"`
	EndRune    *int `json:"end_rune,omitempty"`
	RuneLength *int `json:"rune_length,omitempty"`
}

type DebugMetadata struct {
	RetrievalProfileEffective string         `json:"retrieval_profile_effective"`
	SemanticWeightEffective   float64        `json:"semantic_weight_effective"`
	AutoSignalsDetected       []string       `json:"auto_signals_detected,omitempty"`
	LexicalCandidates         int            `json:"lexical_candidates"`
	SemanticCandidates        int            `json:"semantic_candidates"`
	RerankerApplied           bool           `json:"reranker_applied"`
	FiltersApplied            map[string]any `json:"filters_applied,omitempty"`
}

type HydrateRequest struct {
	KnowledgeBaseID string
	ChunkIDs        []string
	AdjacentBefore  int
	AdjacentAfter   int
}

type HydrateResponse struct {
	KnowledgeBaseID string   `json:"kb_id"`
	ChunkCount      int      `json:"chunk_count"`
	Chunks          []Result `json:"chunks"`
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
	if req.RetrievalProfile != "" && !IsValidRetrievalProfile(req.RetrievalProfile) {
		return ErrInvalidProfile
	}
	if req.Filters.CreatedAfter != nil && req.Filters.CreatedBefore != nil {
		if req.Filters.CreatedAfter.After(*req.Filters.CreatedBefore) {
			return ErrInvalidCreatedAfter
		}
	}
	return nil
}

func ValidateHydrateRequest(req HydrateRequest) error {
	if req.KnowledgeBaseID == "" {
		return ErrMissingKnowledgeBase
	}
	if len(req.ChunkIDs) == 0 {
		return ErrMissingChunkIDs
	}
	if len(req.ChunkIDs) > MaxHydrateChunkIDs {
		return ErrTooManyChunkIDs
	}
	if req.AdjacentBefore < 0 || req.AdjacentBefore > 10 || req.AdjacentAfter < 0 || req.AdjacentAfter > 10 {
		return ErrInvalidAdjacentRange
	}
	return nil
}

func IsValidRetrievalProfile(profile string) bool {
	switch strings.ToLower(strings.TrimSpace(profile)) {
	case RetrievalProfileAuto, RetrievalProfileExact, RetrievalProfileBalanced, RetrievalProfileSemantic:
		return true
	default:
		return false
	}
}
