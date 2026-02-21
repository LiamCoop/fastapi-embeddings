-- name: HasEmbedding :one
SELECT 1
FROM embeddings
WHERE kb_id = $1
  AND content_hash = $2
  AND embedding_model_id = $3
LIMIT 1;

-- name: InsertEmbedding :exec
INSERT INTO embeddings (
    id,
    kb_id,
    content_hash,
    embedding_model_id,
    embedding_vector,
    created_at
) VALUES ($1, $2, $3, $4, $5, $6);
