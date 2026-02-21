package ingestion

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type ChunkRecord struct {
	ID                string
	DocumentVersionID string
	KnowledgeBaseID   string
	SequenceNumber    int
	Content           string
	ContentHash       string
	Metadata          map[string]any
	ChunkingStrategy  string
	EmbeddingID       *string
	CreatedAt         time.Time
}

type JobStatus string

const (
	JobQueued     JobStatus = "QUEUED"
	JobInProgress JobStatus = "IN_PROGRESS"
	JobSuccess    JobStatus = "SUCCESS"
	JobFailed     JobStatus = "FAILED"
)

type IngestionJobRecord struct {
	ID                string
	DocumentVersionID string
	KnowledgeBaseID   string
	JobStatus         JobStatus
	ErrorMessage      *string
	StartedAt         *time.Time
	CompletedAt       *time.Time
}

// Repository persists ingestion artifacts.
type Repository interface {
	InsertIngestionJob(ctx context.Context, job IngestionJobRecord) error
	UpdateIngestionJob(ctx context.Context, jobID string, status JobStatus, errorMessage *string, completedAt *time.Time) error
	InsertChunks(ctx context.Context, chunks []ChunkRecord) error
	UpdateChunkEmbedding(ctx context.Context, chunkID, embeddingID string) error
}

// PostgresRepository persists ingestion artifacts to Postgres.
type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) InsertIngestionJob(ctx context.Context, job IngestionJobRecord) error {
	jobUUID, err := uuid.Parse(job.ID)
	if err != nil {
		return err
	}
	versionUUID, err := uuid.Parse(job.DocumentVersionID)
	if err != nil {
		return err
	}
	kbUUID, err := uuid.Parse(job.KnowledgeBaseID)
	if err != nil {
		return err
	}

	_, err = r.db.ExecContext(ctx, `
		INSERT INTO ingestion_jobs (
			id,
			document_version_id,
			kb_id,
			job_status,
			error_message,
			started_at,
			completed_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
	`,
		jobUUID,
		versionUUID,
		kbUUID,
		job.JobStatus,
		job.ErrorMessage,
		job.StartedAt,
		job.CompletedAt,
	)
	return err
}

func (r *PostgresRepository) UpdateIngestionJob(
	ctx context.Context,
	jobID string,
	status JobStatus,
	errorMessage *string,
	completedAt *time.Time,
) error {
	jobUUID, err := uuid.Parse(jobID)
	if err != nil {
		return err
	}

	_, err = r.db.ExecContext(ctx, `
		UPDATE ingestion_jobs
		SET job_status = $2,
			error_message = $3,
			completed_at = $4
		WHERE id = $1
	`, jobUUID, status, errorMessage, completedAt)
	return err
}

func (r *PostgresRepository) InsertChunks(ctx context.Context, chunks []ChunkRecord) error {
	if len(chunks) == 0 {
		return nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	rollback := func() {
		_ = tx.Rollback()
	}

	for _, chunk := range chunks {
		chunkUUID, err := uuid.Parse(chunk.ID)
		if err != nil {
			rollback()
			return err
		}
		versionUUID, err := uuid.Parse(chunk.DocumentVersionID)
		if err != nil {
			rollback()
			return err
		}
		kbUUID, err := uuid.Parse(chunk.KnowledgeBaseID)
		if err != nil {
			rollback()
			return err
		}

		metadata := json.RawMessage([]byte("{}"))
		if chunk.Metadata != nil {
			encoded, err := json.Marshal(chunk.Metadata)
			if err != nil {
				rollback()
				return err
			}
			metadata = encoded
		}

		var embeddingUUID any
		if chunk.EmbeddingID != nil && *chunk.EmbeddingID != "" {
			value, err := uuid.Parse(*chunk.EmbeddingID)
			if err != nil {
				rollback()
				return err
			}
			embeddingUUID = value
		} else {
			embeddingUUID = nil
		}

		_, err = tx.ExecContext(ctx, `
			INSERT INTO chunks (
				id,
				document_version_id,
				kb_id,
				sequence_number,
				content,
				content_hash,
				metadata,
				chunking_strategy,
				embedding_id,
				created_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		`,
			chunkUUID,
			versionUUID,
			kbUUID,
			chunk.SequenceNumber,
			chunk.Content,
			chunk.ContentHash,
			metadata,
			chunk.ChunkingStrategy,
			embeddingUUID,
			chunk.CreatedAt,
		)
		if err != nil {
			rollback()
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		rollback()
		return err
	}

	return nil
}

func (r *PostgresRepository) UpdateChunkEmbedding(ctx context.Context, chunkID, embeddingID string) error {
	chunkUUID, err := uuid.Parse(chunkID)
	if err != nil {
		return err
	}
	embeddingUUID, err := uuid.Parse(embeddingID)
	if err != nil {
		return err
	}

	_, err = r.db.ExecContext(ctx, `
		UPDATE chunks
		SET embedding_id = $2
		WHERE id = $1
	`, chunkUUID, embeddingUUID)
	return err
}
