package domain

import "time"

// JSONMap represents JSONB metadata fields.
type JSONMap map[string]any

type KnowledgeBase struct {
	ID        string
	Name      string
	Metadata  JSONMap
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Document struct {
	ID              string
	KBID            string
	Path            string
	Title           string
	DocumentType    string
	SourceMetadata  JSONMap
	ActiveVersionID *string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type ProcessingStatus string

const (
	StatusReceived           ProcessingStatus = "RECEIVED"
	StatusStored             ProcessingStatus = "STORED"
	StatusExtracted          ProcessingStatus = "EXTRACTED"
	StatusChunked            ProcessingStatus = "CHUNKED"
	StatusEmbedded           ProcessingStatus = "EMBEDDED"
	StatusActivated          ProcessingStatus = "ACTIVATED"
	StatusSkippedUnsupported ProcessingStatus = "SKIPPED_UNSUPPORTED"
	StatusFailed             ProcessingStatus = "FAILED"
)

type DocumentVersion struct {
	ID                   string
	DocumentID           string
	KBID                 string
	VersionNumber        int
	RawContentURI        string
	ExtractedContent     *string
	ExtractedContentHash *string
	ProcessingStatus     ProcessingStatus
	ErrorMessage         *string
	IsActive             bool
	CreatedAt            time.Time
}

type Chunk struct {
	ID                string
	DocumentVersionID string
	KBID              string
	SequenceNumber    int
	Content           string
	ContentHash       string
	Metadata          JSONMap
	ChunkingStrategy  string
	EmbeddingID       *string
	CreatedAt         time.Time
}

type Embedding struct {
	ID                   string
	KBID                 string
	ContentHash          string
	EmbeddingModelID     string
	EmbeddingVector      []float32
	CreatedAt            time.Time
}

type EmbeddingModel struct {
	ID              string
	Name            string
	VectorDimension int
	Provider        string
	Parameters      JSONMap
	IsActive        bool
	CreatedAt       time.Time
}
