package chunking

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/google/uuid"

	"ragtime-backend/internal/domain"
	"ragtime-backend/internal/storage/sqlc"
)

// PostgresRepository persists chunks to Postgres.
type PostgresRepository struct {
	db      *sql.DB
	queries *sqlc.Queries
}

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{
		db:      db,
		queries: sqlc.New(db),
	}
}

func (r *PostgresRepository) InsertChunks(ctx context.Context, chunks []domain.Chunk) error {
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

	queries := r.queries.WithTx(tx)
	for i := range chunks {
		kbUUID, err := uuid.Parse(chunks[i].KBID)
		if err != nil {
			rollback()
			return err
		}

		versionUUID, err := uuid.Parse(chunks[i].DocumentVersionID)
		if err != nil {
			rollback()
			return err
		}

		if chunks[i].ID == "" {
			chunks[i].ID = uuid.NewString()
		}

		chunkUUID, err := uuid.Parse(chunks[i].ID)
		if err != nil {
			rollback()
			return err
		}

		embeddingID := uuid.NullUUID{}
		if chunks[i].EmbeddingID != nil && *chunks[i].EmbeddingID != "" {
			embedUUID, err := uuid.Parse(*chunks[i].EmbeddingID)
			if err != nil {
				rollback()
				return err
			}
			embeddingID = uuid.NullUUID{UUID: embedUUID, Valid: true}
		}

		_, err = queries.InsertChunk(ctx, sqlc.InsertChunkParams{
			ID:                chunkUUID,
			DocumentVersionID: versionUUID,
			KbID:              kbUUID,
			SequenceNumber:    int32(chunks[i].SequenceNumber),
			Content:           chunks[i].Content,
			ContentHash:       chunks[i].ContentHash,
			Metadata:          encodeMetadata(chunks[i].Metadata),
			ChunkingStrategy:  chunks[i].ChunkingStrategy,
			EmbeddingID:       embeddingID,
			CreatedAt:         chunks[i].CreatedAt,
		})

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

func (r *PostgresRepository) DeleteChunksByDocumentVersion(ctx context.Context, documentVersionID string) error {
	versionUUID, err := uuid.Parse(documentVersionID)
	if err != nil {
		return err
	}

	_, err = r.db.ExecContext(ctx, `
		DELETE FROM chunks
		WHERE document_version_id = $1
	`, versionUUID)
	return err
}

func (r *PostgresRepository) UpdateDocumentVersionStatus(ctx context.Context, versionID, status string, errorMessage *string) error {
	versionUUID, err := uuid.Parse(versionID)
	if err != nil {
		return err
	}

	return r.queries.UpdateDocumentVersionStatus(ctx, sqlc.UpdateDocumentVersionStatusParams{
		ID:               versionUUID,
		ProcessingStatus: status,
		ErrorMessage:     toNullString(errorMessage),
	})
}

func (r *PostgresRepository) GetLatestDocumentVersionForDocument(
	ctx context.Context,
	knowledgeBaseID,
	documentID string,
) (*DocumentVersionRef, error) {
	kbUUID, err := uuid.Parse(knowledgeBaseID)
	if err != nil {
		return nil, err
	}
	docUUID, err := uuid.Parse(documentID)
	if err != nil {
		return nil, err
	}

	var versionID uuid.UUID
	var rawContentURI string
	err = r.db.QueryRowContext(ctx, `
		SELECT id, raw_content_uri
		FROM document_versions
		WHERE kb_id = $1 AND document_id = $2
		ORDER BY version_number DESC
		LIMIT 1
	`, kbUUID, docUUID).Scan(&versionID, &rawContentURI)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &DocumentVersionRef{
		DocumentVersionID: versionID.String(),
		RawContentURI:     rawContentURI,
	}, nil
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

func toNullString(value *string) sql.NullString {
	if value == nil {
		return sql.NullString{}
	}
	return sql.NullString{String: *value, Valid: true}
}

var _ Repository = (*PostgresRepository)(nil)
