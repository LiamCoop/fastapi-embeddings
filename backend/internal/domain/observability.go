package domain

import "time"

type RetrievalRequest struct {
	ID           string
	KBID         string
	Query        string
	Filters      JSONMap
	TopK         int
	HybridWeight float64
	ResultCount  int
	LatencyMS    int64
	EmptyResult  bool
	CreatedAt    time.Time
}

type RetrievalResult struct {
	ID                 string
	RetrievalRequestID string
	ChunkID            string
	Rank               int
	SemanticScore      float64
	LexicalScore       float64
	FinalScore         float64
	CreatedAt          time.Time
}

type JobStatus string

const (
	JobQueued     JobStatus = "QUEUED"
	JobInProgress JobStatus = "IN_PROGRESS"
	JobSuccess    JobStatus = "SUCCESS"
	JobFailed     JobStatus = "FAILED"
)

type IngestionJob struct {
	ID                string
	DocumentVersionID string
	KBID              string
	JobStatus         JobStatus
	ErrorMessage      *string
	StartedAt         *time.Time
	CompletedAt       *time.Time
}

type ProcessingStage string

const (
	StageReceived  ProcessingStage = "RECEIVED"
	StageStored    ProcessingStage = "STORED"
	StageExtracted ProcessingStage = "EXTRACTED"
	StageChunked   ProcessingStage = "CHUNKED"
	StageEmbedded  ProcessingStage = "EMBEDDED"
	StageActivated ProcessingStage = "ACTIVATED"
)

type ProcessingMetric struct {
	ID             string
	IngestionJobID string
	Stage          ProcessingStage
	DurationMS     int64
	ItemCount      int
	Status         string
	ErrorMessage   *string
	CreatedAt      time.Time
}
