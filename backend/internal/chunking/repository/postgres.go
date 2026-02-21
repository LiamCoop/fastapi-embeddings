package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"ragtime-backend/internal/domain"
	"ragtime-backend/internal/storage/sqlc"
)

// PostgresStore persists chunking data to Postgres.
type PostgresStore struct {
	db      *sql.DB
	queries *sqlc.Queries
}

func NewPostgresStore(db *sql.DB) *PostgresStore {
	return &PostgresStore{
		db:      db,
		queries: sqlc.New(db),
	}
}

func (r *PostgresStore) InsertChunks(ctx context.Context, chunks []domain.Chunk) error {
	if len(chunks) == 0 {
		return nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	rollback := func() { _ = tx.Rollback() }

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

func (r *PostgresStore) DeleteChunksByDocumentVersion(ctx context.Context, documentVersionID string) error {
	versionUUID, err := uuid.Parse(documentVersionID)
	if err != nil {
		return err
	}
	_, err = r.db.ExecContext(ctx, `DELETE FROM chunks WHERE document_version_id = $1`, versionUUID)
	return err
}

func (r *PostgresStore) DeleteChunksByDocument(ctx context.Context, knowledgeBaseID, documentID string) error {
	kbUUID, err := uuid.Parse(knowledgeBaseID)
	if err != nil {
		return err
	}
	docUUID, err := uuid.Parse(documentID)
	if err != nil {
		return err
	}

	_, err = r.db.ExecContext(ctx, `
		DELETE FROM chunks c
		USING document_versions dv
		WHERE c.document_version_id = dv.id
		  AND dv.kb_id = $1
		  AND dv.document_id = $2
	`, kbUUID, docUUID)
	return err
}

func (r *PostgresStore) GetChunkByID(ctx context.Context, knowledgeBaseID, chunkID string) (*domain.Chunk, error) {
	kbUUID, err := uuid.Parse(knowledgeBaseID)
	if err != nil {
		return nil, err
	}
	chunkUUID, err := uuid.Parse(chunkID)
	if err != nil {
		return nil, err
	}

	var row domain.Chunk
	var metadataRaw []byte
	var embeddingID uuid.NullUUID
	err = r.db.QueryRowContext(ctx, `
		SELECT
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
		FROM chunks
		WHERE kb_id = $1 AND id = $2
	`, kbUUID, chunkUUID).Scan(
		&row.ID,
		&row.DocumentVersionID,
		&row.KBID,
		&row.SequenceNumber,
		&row.Content,
		&row.ContentHash,
		&metadataRaw,
		&row.ChunkingStrategy,
		&embeddingID,
		&row.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if len(metadataRaw) > 0 {
		if err := json.Unmarshal(metadataRaw, &row.Metadata); err != nil {
			return nil, fmt.Errorf("decode chunk metadata: %w", err)
		}
	} else {
		row.Metadata = domain.JSONMap{}
	}
	if embeddingID.Valid {
		embed := embeddingID.UUID.String()
		row.EmbeddingID = &embed
	}

	return &row, nil
}

func (r *PostgresStore) UpdateChunkEmbedding(ctx context.Context, knowledgeBaseID, chunkID, embeddingID string) error {
	kbUUID, err := uuid.Parse(knowledgeBaseID)
	if err != nil {
		return err
	}
	chunkUUID, err := uuid.Parse(chunkID)
	if err != nil {
		return err
	}
	embeddingUUID, err := uuid.Parse(embeddingID)
	if err != nil {
		return err
	}

	result, err := r.db.ExecContext(ctx, `
		UPDATE chunks
		SET embedding_id = $3
		WHERE kb_id = $1 AND id = $2
	`, kbUUID, chunkUUID, embeddingUUID)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *PostgresStore) UpdateDocumentVersionStatus(ctx context.Context, versionID, status string, errorMessage *string) error {
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

func (r *PostgresStore) ActivateDocumentVersion(ctx context.Context, versionID string) error {
	versionUUID, err := uuid.Parse(versionID)
	if err != nil {
		return err
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	rollback := func() { _ = tx.Rollback() }

	// Deactivate all other versions for the same document.
	_, err = tx.ExecContext(ctx, `
		UPDATE document_versions SET is_active = false
		WHERE document_id = (SELECT document_id FROM document_versions WHERE id = $1)
	`, versionUUID)
	if err != nil {
		rollback()
		return err
	}

	// Activate this version and advance its status.
	_, err = tx.ExecContext(ctx, `
		UPDATE document_versions SET is_active = true, processing_status = 'ACTIVATED'
		WHERE id = $1
	`, versionUUID)
	if err != nil {
		rollback()
		return err
	}

	// Point the parent document at this version.
	_, err = tx.ExecContext(ctx, `
		UPDATE documents SET active_version_id = $1, updated_at = now()
		WHERE id = (SELECT document_id FROM document_versions WHERE id = $1)
	`, versionUUID)
	if err != nil {
		rollback()
		return err
	}

	return tx.Commit()
}

func (r *PostgresStore) GetLatestDocumentVersionForDocument(
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

var _ Store = (*PostgresStore)(nil)
