-- name: InsertRetrievalRequest :one
INSERT INTO retrieval_requests (
    id,
    kb_id,
    query,
    filters,
    top_k,
    hybrid_weight,
    result_count,
    latency_ms,
    empty_result,
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

-- name: UpdateRetrievalRequest :exec
UPDATE retrieval_requests
SET result_count = $2,
    latency_ms = $3,
    empty_result = $4
WHERE id = $1;

-- name: InsertRetrievalResult :exec
INSERT INTO retrieval_results (
    id,
    retrieval_request_id,
    chunk_id,
    rank,
    semantic_score,
    lexical_score,
    final_score,
    created_at
) VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7,
    $8
);

-- name: SearchSemantic :many
SELECT
    c.id AS chunk_id,
    CAST(1.0 - (e.embedding_vector <=> sqlc.arg('query_vector')::vector) AS double precision) AS semantic_score
FROM chunks c
JOIN embeddings e ON c.embedding_id = e.id
JOIN document_versions dv ON c.document_version_id = dv.id
JOIN documents d ON dv.document_id = d.id
WHERE dv.is_active = true
  AND c.kb_id = sqlc.arg('kb_id')
  AND (sqlc.narg('document_type')::text IS NULL OR d.document_type = sqlc.narg('document_type'))
  AND (sqlc.narg('path_prefix')::text IS NULL OR d.path LIKE sqlc.narg('path_prefix'))
  AND (sqlc.narg('source')::text IS NULL OR d.source_metadata ->> 'source' = sqlc.narg('source'))
  AND (sqlc.arg('tags')::jsonb = '{}'::jsonb OR d.source_metadata @> sqlc.arg('tags')::jsonb)
  AND (sqlc.narg('created_after')::timestamptz IS NULL OR dv.created_at >= sqlc.narg('created_after'))
  AND (sqlc.narg('created_before')::timestamptz IS NULL OR dv.created_at <= sqlc.narg('created_before'))
ORDER BY e.embedding_vector <=> sqlc.arg('query_vector')::vector
LIMIT sqlc.arg('limit');

-- name: SearchLexical :many
SELECT
    c.id AS chunk_id,
    ts_rank(to_tsvector('english', c.content), plainto_tsquery('english', sqlc.arg('query'))) AS lexical_score
FROM chunks c
JOIN document_versions dv ON c.document_version_id = dv.id
JOIN documents d ON dv.document_id = d.id
WHERE dv.is_active = true
  AND c.kb_id = sqlc.arg('kb_id')
  AND to_tsvector('english', c.content) @@ plainto_tsquery('english', sqlc.arg('query'))
  AND (sqlc.narg('document_type')::text IS NULL OR d.document_type = sqlc.narg('document_type'))
  AND (sqlc.narg('path_prefix')::text IS NULL OR d.path LIKE sqlc.narg('path_prefix'))
  AND (sqlc.narg('source')::text IS NULL OR d.source_metadata ->> 'source' = sqlc.narg('source'))
  AND (sqlc.arg('tags')::jsonb = '{}'::jsonb OR d.source_metadata @> sqlc.arg('tags')::jsonb)
  AND (sqlc.narg('created_after')::timestamptz IS NULL OR dv.created_at >= sqlc.narg('created_after'))
  AND (sqlc.narg('created_before')::timestamptz IS NULL OR dv.created_at <= sqlc.narg('created_before'))
ORDER BY ts_rank(to_tsvector('english', c.content), plainto_tsquery('english', sqlc.arg('query'))) DESC
LIMIT sqlc.arg('limit');

-- name: GetChunksWithDocuments :many
SELECT
    c.id AS chunk_id,
    c.document_version_id,
    c.sequence_number,
    c.content,
    c.metadata,
    d.id AS document_id,
    d.path AS document_path,
    d.title AS document_title,
    d.document_type AS document_type,
    d.source_metadata AS source_metadata,
    dv.version_number,
    dv.created_at AS version_created_at
FROM chunks c
JOIN document_versions dv ON c.document_version_id = dv.id
JOIN documents d ON dv.document_id = d.id
WHERE c.id = ANY(sqlc.arg('chunk_ids')::uuid[]);
