package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/pgvector/pgvector-go"

	"ragtime-backend/internal/retrieval"
	"ragtime-backend/internal/storage/sqlc"
)

// PostgresStore persists retrieval data to Postgres.
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

func (r *PostgresStore) InsertRetrievalRequest(ctx context.Context, req retrieval.RetrievalRequestRecord) (*retrieval.RetrievalRequestRecord, error) {
	reqID, err := uuid.Parse(req.ID)
	if err != nil {
		return nil, err
	}
	kbID, err := uuid.Parse(req.KnowledgeBase)
	if err != nil {
		return nil, err
	}

	filtersJSON := encodeJSON(req.Filters)

	row, err := r.queries.InsertRetrievalRequest(ctx, sqlc.InsertRetrievalRequestParams{
		ID:           reqID,
		KbID:         kbID,
		Query:        req.Query,
		Filters:      filtersJSON,
		TopK:         int32(req.TopK),
		HybridWeight: req.HybridWeight,
		ResultCount:  int32(req.ResultCount),
		LatencyMs:    req.LatencyMS,
		EmptyResult:  req.EmptyResult,
		CreatedAt:    req.CreatedAt,
	})
	if err != nil {
		return nil, err
	}

	return &retrieval.RetrievalRequestRecord{
		ID:            row.ID.String(),
		KnowledgeBase: row.KbID.String(),
		Query:         row.Query,
		Filters:       decodeJSON(row.Filters),
		TopK:          int(row.TopK),
		HybridWeight:  row.HybridWeight,
		ResultCount:   int(row.ResultCount),
		LatencyMS:     row.LatencyMs,
		EmptyResult:   row.EmptyResult,
		CreatedAt:     row.CreatedAt,
	}, nil
}

func (r *PostgresStore) UpdateRetrievalRequest(ctx context.Context, requestID string, resultCount int, latencyMS int64, emptyResult bool) error {
	reqID, err := uuid.Parse(requestID)
	if err != nil {
		return err
	}

	return r.queries.UpdateRetrievalRequest(ctx, sqlc.UpdateRetrievalRequestParams{
		ID:          reqID,
		ResultCount: int32(resultCount),
		LatencyMs:   latencyMS,
		EmptyResult: emptyResult,
	})
}

func (r *PostgresStore) InsertRetrievalResults(ctx context.Context, results []retrieval.RetrievalResultRecord) error {
	if len(results) == 0 {
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
	for _, result := range results {
		resultID, err := uuid.Parse(result.ID)
		if err != nil {
			rollback()
			return err
		}
		requestID, err := uuid.Parse(result.RetrievalRequestID)
		if err != nil {
			rollback()
			return err
		}
		chunkID, err := uuid.Parse(result.ChunkID)
		if err != nil {
			rollback()
			return err
		}

		if err := queries.InsertRetrievalResult(ctx, sqlc.InsertRetrievalResultParams{
			ID:                 resultID,
			RetrievalRequestID: requestID,
			ChunkID:            chunkID,
			Rank:               int32(result.Rank),
			SemanticScore:      result.SemanticScore,
			LexicalScore:       result.LexicalScore,
			FinalScore:         result.FinalScore,
			CreatedAt:          result.CreatedAt,
		}); err != nil {
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

func (r *PostgresStore) SearchSemantic(ctx context.Context, params retrieval.SearchParams) ([]retrieval.ScoredChunk, error) {
	kbID, err := uuid.Parse(params.KnowledgeBaseID)
	if err != nil {
		return nil, err
	}

	table := fmt.Sprintf("embeddings_%d", params.VectorDimension)
	vector := pgvector.NewVector(params.QueryVector)

	query := fmt.Sprintf(`
SELECT
    c.id AS chunk_id,
    CAST(1.0 - (e.embedding_vector <=> $1::vector) AS double precision) AS semantic_score
FROM chunks c
JOIN %s e ON c.embedding_id = e.id
JOIN document_versions dv ON c.document_version_id = dv.id
JOIN documents d ON dv.document_id = d.id
WHERE dv.is_active = true
  AND c.kb_id = $2
  AND ($3::text IS NULL OR d.document_type = $3)
  AND ($4::text IS NULL OR d.path LIKE $4)
  AND ($5::text IS NULL OR d.source_metadata ->> 'source' = $5)
  AND ($6::jsonb = '{}'::jsonb OR d.source_metadata @> $6::jsonb)
  AND ($7::timestamptz IS NULL OR dv.created_at >= $7)
  AND ($8::timestamptz IS NULL OR dv.created_at <= $8)
ORDER BY e.embedding_vector <=> $1::vector
LIMIT $9`, table)

	rows, err := r.db.QueryContext(ctx, query,
		vector,
		kbID,
		toNullString(params.DocumentType),
		toNullString(params.PathPrefix),
		toNullString(params.Source),
		toJSON(params.TagsFilter),
		toNullTime(params.CreatedAfter),
		toNullTime(params.CreatedBefore),
		int32(params.Limit),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []retrieval.ScoredChunk
	for rows.Next() {
		var chunkID uuid.UUID
		var score float64
		if err := rows.Scan(&chunkID, &score); err != nil {
			return nil, err
		}
		results = append(results, retrieval.ScoredChunk{ChunkID: chunkID.String(), Score: score})
	}
	return results, rows.Err()
}

func (r *PostgresStore) SearchLexical(ctx context.Context, params retrieval.SearchParams) ([]retrieval.ScoredChunk, error) {
	kbID, err := uuid.Parse(params.KnowledgeBaseID)
	if err != nil {
		return nil, err
	}

	rows, err := r.queries.SearchLexical(ctx, sqlc.SearchLexicalParams{
		Query:         params.Query,
		KbID:          kbID,
		DocumentType:  toNullString(params.DocumentType),
		PathPrefix:    toNullString(params.PathPrefix),
		Source:        toNullString(params.Source),
		Tags:          toJSON(params.TagsFilter),
		CreatedAfter:  toNullTime(params.CreatedAfter),
		CreatedBefore: toNullTime(params.CreatedBefore),
		Limit:         int32(params.Limit),
	})
	if err != nil {
		return nil, err
	}

	results := make([]retrieval.ScoredChunk, 0, len(rows))
	for _, row := range rows {
		results = append(results, retrieval.ScoredChunk{
			ChunkID: row.ChunkID.String(),
			Score:   float64(row.LexicalScore),
		})
	}

	return results, nil
}

func (r *PostgresStore) GetChunksWithDocuments(ctx context.Context, chunkIDs []string) ([]retrieval.ChunkRecord, error) {
	if len(chunkIDs) == 0 {
		return nil, nil
	}

	ids := make([]uuid.UUID, 0, len(chunkIDs))
	for _, id := range chunkIDs {
		parsed, err := uuid.Parse(id)
		if err != nil {
			return nil, err
		}
		ids = append(ids, parsed)
	}

	rows, err := r.queries.GetChunksWithDocuments(ctx, ids)
	if err != nil {
		return nil, err
	}

	results := make([]retrieval.ChunkRecord, 0, len(rows))
	for _, row := range rows {
		results = append(results, retrieval.ChunkRecord{
			ChunkID:           row.ChunkID.String(),
			DocumentID:        row.DocumentID.String(),
			DocumentVersionID: row.DocumentVersionID.String(),
			DocumentPath:      row.DocumentPath,
			DocumentTitle:     nullStringPtr(row.DocumentTitle),
			DocumentType:      row.DocumentType,
			Content:           row.Content,
			Metadata:          decodeJSON(row.Metadata),
			VersionNumber:     row.VersionNumber,
			SequenceNumber:    row.SequenceNumber,
			SourceMetadata:    decodeJSON(row.SourceMetadata),
		})
	}

	return results, nil
}

func toNullString(value *string) sql.NullString {
	if value == nil || *value == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: *value, Valid: true}
}

func toNullTime(value *time.Time) sql.NullTime {
	if value == nil {
		return sql.NullTime{}
	}
	return sql.NullTime{Time: *value, Valid: true}
}

func toJSON(value map[string]any) json.RawMessage {
	if value == nil {
		return json.RawMessage([]byte("{}"))
	}
	encoded, err := json.Marshal(value)
	if err != nil {
		return json.RawMessage([]byte("{}"))
	}
	return encoded
}

func encodeJSON(value map[string]any) json.RawMessage {
	if value == nil {
		return json.RawMessage([]byte("{}"))
	}
	encoded, err := json.Marshal(value)
	if err != nil {
		return json.RawMessage([]byte("{}"))
	}
	return encoded
}

func decodeJSON(value json.RawMessage) map[string]any {
	if len(value) == 0 {
		return map[string]any{}
	}
	var decoded map[string]any
	if err := json.Unmarshal(value, &decoded); err != nil {
		return map[string]any{}
	}
	return decoded
}

func nullStringPtr(value sql.NullString) *string {
	if !value.Valid {
		return nil
	}
	return &value.String
}

var _ Store = (*PostgresStore)(nil)
