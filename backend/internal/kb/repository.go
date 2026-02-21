package kb

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"

	"ragtime-backend/internal/storage/sqlc"
)

// Repository defines persistence behavior for knowledge bases.
type Repository interface {
	InsertKnowledgeBase(ctx context.Context, kb KnowledgeBaseRecord) (*KnowledgeBaseRecord, error)
	GetKnowledgeBase(ctx context.Context, id string) (*KnowledgeBaseRecord, error)
	ListKnowledgeBases(ctx context.Context) ([]KnowledgeBaseRecord, error)
	UpdateKnowledgeBase(ctx context.Context, kb KnowledgeBaseRecord) (*KnowledgeBaseRecord, error)
	DeleteKnowledgeBase(ctx context.Context, id string) (bool, error)
}

// PostgresRepository persists knowledge bases to Postgres.
type PostgresRepository struct {
	db      *sql.DB
	queries *sqlc.Queries
}

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db, queries: sqlc.New(db)}
}

func (r *PostgresRepository) InsertKnowledgeBase(ctx context.Context, kb KnowledgeBaseRecord) (*KnowledgeBaseRecord, error) {
	kbUUID, err := uuid.Parse(kb.ID)
	if err != nil {
		return nil, err
	}

	row, err := r.queries.InsertKnowledgeBase(ctx, sqlc.InsertKnowledgeBaseParams{
		ID:        kbUUID,
		Name:      kb.Name,
		Metadata:  encodeMetadata(kb.Metadata),
		CreatedAt: kb.CreatedAt,
		UpdatedAt: kb.UpdatedAt,
	})
	if err != nil {
		return nil, err
	}

	record := toKnowledgeBaseRecord(row)
	return &record, nil
}

func (r *PostgresRepository) GetKnowledgeBase(ctx context.Context, id string) (*KnowledgeBaseRecord, error) {
	kbUUID, err := uuid.Parse(id)
	if err != nil {
		return nil, err
	}

	row, err := r.queries.GetKnowledgeBase(ctx, kbUUID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	record := toKnowledgeBaseRecord(row)
	return &record, nil
}

func (r *PostgresRepository) ListKnowledgeBases(ctx context.Context) ([]KnowledgeBaseRecord, error) {
	rows, err := r.queries.ListKnowledgeBases(ctx)
	if err != nil {
		return nil, err
	}

	records := make([]KnowledgeBaseRecord, 0, len(rows))
	for _, row := range rows {
		records = append(records, toKnowledgeBaseRecord(row))
	}

	return records, nil
}

func (r *PostgresRepository) UpdateKnowledgeBase(ctx context.Context, kb KnowledgeBaseRecord) (*KnowledgeBaseRecord, error) {
	kbUUID, err := uuid.Parse(kb.ID)
	if err != nil {
		return nil, err
	}

	row, err := r.queries.UpdateKnowledgeBase(ctx, sqlc.UpdateKnowledgeBaseParams{
		ID:        kbUUID,
		Name:      kb.Name,
		Metadata:  encodeMetadata(kb.Metadata),
		UpdatedAt: kb.UpdatedAt,
	})
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	record := toKnowledgeBaseRecord(row)
	return &record, nil
}

func (r *PostgresRepository) DeleteKnowledgeBase(ctx context.Context, id string) (bool, error) {
	kbUUID, err := uuid.Parse(id)
	if err != nil {
		return false, err
	}

	rows, err := r.queries.DeleteKnowledgeBase(ctx, kbUUID)
	if err != nil {
		return false, err
	}

	return rows > 0, nil
}

// KnowledgeBaseRecord captures the repository representation of a KB.
type KnowledgeBaseRecord struct {
	ID        string
	Name      string
	Metadata  map[string]any
	CreatedAt time.Time
	UpdatedAt time.Time
}

func toKnowledgeBaseRecord(row sqlc.KnowledgeBasis) KnowledgeBaseRecord {
	metadata := map[string]any{}
	_ = json.Unmarshal(row.Metadata, &metadata)

	return KnowledgeBaseRecord{
		ID:        row.ID.String(),
		Name:      row.Name,
		Metadata:  metadata,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}
}

func encodeMetadata(metadata map[string]any) json.RawMessage {
	if metadata == nil {
		return json.RawMessage([]byte("{}"))
	}

	encoded, err := json.Marshal(metadata)
	if err != nil {
		return json.RawMessage([]byte("{}"))
	}

	return encoded
}

var _ Repository = (*PostgresRepository)(nil)
