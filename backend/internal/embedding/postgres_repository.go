package embedding

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/pgvector/pgvector-go"
)

// PostgresRepository persists embeddings to Postgres with pgvector support.
type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

// embeddingTable returns the dimension-specific table name (e.g. "embeddings_384", "embeddings_1536").
func embeddingTable(dimension int) string {
	return fmt.Sprintf("embeddings_%d", dimension)
}

// lookupModelDimension queries embedding_models for the vector_dimension of the given model ID.
func lookupModelDimension(ctx context.Context, db *sql.DB, modelID string) (int, error) {
	var dim int
	err := db.QueryRowContext(ctx,
		`SELECT vector_dimension FROM embedding_models WHERE id = $1`,
		modelID,
	).Scan(&dim)
	if err != nil {
		return 0, fmt.Errorf("lookup dimension for model %q: %w", modelID, err)
	}
	return dim, nil
}

func (r *PostgresRepository) HasEmbedding(
	ctx context.Context,
	knowledgeBaseID, contentHash, modelID string,
) (bool, error) {
	_, found, err := r.FindEmbeddingID(ctx, knowledgeBaseID, contentHash, modelID)
	return found, err
}

func (r *PostgresRepository) FindEmbeddingID(
	ctx context.Context,
	knowledgeBaseID, contentHash, modelID string,
) (string, bool, error) {
	kbUUID, err := uuid.Parse(knowledgeBaseID)
	if err != nil {
		return "", false, err
	}

	dim, err := lookupModelDimension(ctx, r.db, modelID)
	if err != nil {
		return "", false, err
	}

	var embeddingID uuid.UUID
	err = r.db.QueryRowContext(ctx,
		fmt.Sprintf(`SELECT id FROM %s WHERE kb_id = $1 AND content_hash = $2 AND embedding_model_id = $3 LIMIT 1`,
			embeddingTable(dim)),
		kbUUID, contentHash, modelID,
	).Scan(&embeddingID)
	if err == sql.ErrNoRows {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}

	return embeddingID.String(), true, nil
}

func (r *PostgresRepository) SaveEmbeddings(ctx context.Context, embeddings []EmbeddingResult) ([]EmbeddingResult, error) {
	if len(embeddings) == 0 {
		return nil, nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}

	rollback := func() {
		_ = tx.Rollback()
	}

	now := time.Now().UTC()
	for i := range embeddings {
		kbUUID, err := uuid.Parse(embeddings[i].KnowledgeBaseID)
		if err != nil {
			rollback()
			return nil, err
		}

		if embeddings[i].EmbeddingID == "" {
			embeddings[i].EmbeddingID = uuid.NewString()
		}

		embedUUID, err := uuid.Parse(embeddings[i].EmbeddingID)
		if err != nil {
			rollback()
			return nil, err
		}

		table := embeddingTable(embeddings[i].VectorDimension)
		vector := pgvector.NewVector(embeddings[i].Vector)
		_, err = tx.ExecContext(ctx,
			fmt.Sprintf(`INSERT INTO %s (id, kb_id, content_hash, embedding_model_id, embedding_vector, created_at)
VALUES ($1, $2, $3, $4, $5, $6)`, table),
			embedUUID, kbUUID, embeddings[i].ContentHash, embeddings[i].ModelID, vector, now,
		)
		if err != nil {
			rollback()
			return nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		rollback()
		return nil, err
	}

	return embeddings, nil
}
