-- name: InsertChunk :one
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
) VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7,
    $8,
    $9,
    $10
)
RETURNING *;
