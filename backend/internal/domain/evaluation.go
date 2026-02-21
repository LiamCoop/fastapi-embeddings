package domain

import "time"

type Evaluation struct {
	ID          string
	KBID        string
	Name        string
	Description string
	Status      string
	CreatedAt   time.Time
	CompletedAt *time.Time
}

type EvaluationQuery struct {
	ID               string
	EvaluationID     string
	Query            string
	ExpectedChunkIDs []string
	CreatedAt        time.Time
}

type EvaluationResult struct {
	ID                string
	EvaluationID      string
	QueryID           string
	RetrievedChunkIDs []string
	RecallAtK         float64
	PrecisionAtK      float64
	CreatedAt         time.Time
}
