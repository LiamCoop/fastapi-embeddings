package document

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/google/uuid"

	"ragtime-backend/internal/storage/sqlc"
)

// PostgresRepository persists documents and versions to Postgres.
type PostgresRepository struct {
	db      *sql.DB
	queries *sqlc.Queries
}

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db, queries: sqlc.New(db)}
}

func (r *PostgresRepository) GetDocumentByKBPath(ctx context.Context, kbID, path string) (*DocumentRecord, error) {
	kbUUID, err := uuid.Parse(kbID)
	if err != nil {
		return nil, err
	}

	row, err := r.queries.GetDocumentByKBPath(ctx, sqlc.GetDocumentByKBPathParams{
		KbID: kbUUID,
		Path: path,
	})
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	record := toDocumentRecord(row)
	return &record, nil
}

func (r *PostgresRepository) InsertDocument(ctx context.Context, doc DocumentRecord) (*DocumentRecord, error) {
	kbUUID, err := uuid.Parse(doc.KnowledgeBaseID)
	if err != nil {
		return nil, err
	}
	docUUID, err := uuid.Parse(doc.ID)
	if err != nil {
		return nil, err
	}

	row, err := r.queries.InsertDocument(ctx, sqlc.InsertDocumentParams{
		ID:             docUUID,
		KbID:           kbUUID,
		Path:           doc.Path,
		Title:          toNullString(doc.Title),
		DocumentType:   doc.DocumentType,
		SourceMetadata: encodeMetadata(doc.SourceMetadata),
		CreatedAt:      doc.CreatedAt,
		UpdatedAt:      doc.UpdatedAt,
	})
	if err != nil {
		return nil, err
	}

	record := toDocumentRecord(row)
	return &record, nil
}

func (r *PostgresRepository) UpdateDocument(ctx context.Context, doc DocumentRecord) (*DocumentRecord, error) {
	docUUID, err := uuid.Parse(doc.ID)
	if err != nil {
		return nil, err
	}

	row, err := r.queries.UpdateDocument(ctx, sqlc.UpdateDocumentParams{
		ID:             docUUID,
		Title:          toNullString(doc.Title),
		DocumentType:   doc.DocumentType,
		SourceMetadata: encodeMetadata(doc.SourceMetadata),
		UpdatedAt:      doc.UpdatedAt,
	})
	if err != nil {
		return nil, err
	}

	record := toDocumentRecord(row)
	return &record, nil
}

func (r *PostgresRepository) InsertDocumentVersion(ctx context.Context, version DocumentVersionRecord) (*DocumentVersionRecord, error) {
	docUUID, err := uuid.Parse(version.DocumentID)
	if err != nil {
		return nil, err
	}
	kbUUID, err := uuid.Parse(version.KnowledgeBaseID)
	if err != nil {
		return nil, err
	}
	versionUUID, err := uuid.Parse(version.ID)
	if err != nil {
		return nil, err
	}

	row, err := r.queries.InsertDocumentVersion(ctx, sqlc.InsertDocumentVersionParams{
		ID:               versionUUID,
		DocumentID:       docUUID,
		KbID:             kbUUID,
		RawContentUri:    version.RawContentURI,
		ProcessingStatus: version.ProcessingStatus,
		ErrorMessage:     toNullString(version.ErrorMessage),
		IsActive:         version.IsActive,
		CreatedAt:        version.CreatedAt,
	})
	if err != nil {
		return nil, err
	}

	record := toDocumentVersionRecord(row)
	return &record, nil
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

func (r *PostgresRepository) ActivateVersion(ctx context.Context, versionID string) error {
	versionUUID, err := uuid.Parse(versionID)
	if err != nil {
		return err
	}

	return r.queries.ActivateDocumentVersion(ctx, versionUUID)
}

func toDocumentRecord(row sqlc.Document) DocumentRecord {
	metadata := map[string]any{}
	_ = json.Unmarshal(row.SourceMetadata, &metadata)

	var title *string
	if row.Title.Valid {
		value := row.Title.String
		title = &value
	}

	var activeVersionID *string
	if row.ActiveVersionID.Valid {
		value := row.ActiveVersionID.UUID.String()
		activeVersionID = &value
	}

	return DocumentRecord{
		ID:              row.ID.String(),
		KnowledgeBaseID: row.KbID.String(),
		Path:            row.Path,
		Title:           title,
		DocumentType:    row.DocumentType,
		SourceMetadata:  metadata,
		ActiveVersionID: activeVersionID,
		CreatedAt:       row.CreatedAt,
		UpdatedAt:       row.UpdatedAt,
	}
}

func toDocumentVersionRecord(row sqlc.DocumentVersion) DocumentVersionRecord {
	var errMsg *string
	if row.ErrorMessage.Valid {
		value := row.ErrorMessage.String
		errMsg = &value
	}

	return DocumentVersionRecord{
		ID:               row.ID.String(),
		DocumentID:       row.DocumentID.String(),
		KnowledgeBaseID:  row.KbID.String(),
		VersionNumber:    row.VersionNumber,
		RawContentURI:    row.RawContentUri,
		ProcessingStatus: row.ProcessingStatus,
		ErrorMessage:     errMsg,
		IsActive:         row.IsActive,
		CreatedAt:        row.CreatedAt,
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

func toNullString(value *string) sql.NullString {
	if value == nil {
		return sql.NullString{}
	}
	return sql.NullString{String: *value, Valid: true}
}

var _ Repository = (*PostgresRepository)(nil)
