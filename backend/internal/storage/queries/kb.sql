-- name: InsertKnowledgeBase :one
INSERT INTO knowledge_bases (
    id,
    name,
    metadata,
    created_at,
    updated_at
) VALUES (
    $1,
    $2,
    $3,
    $4,
    $5
)
RETURNING *;

-- name: GetKnowledgeBase :one
SELECT * FROM knowledge_bases
WHERE id = $1;

-- name: ListKnowledgeBases :many
SELECT * FROM knowledge_bases
ORDER BY created_at DESC, id ASC;

-- name: UpdateKnowledgeBase :one
UPDATE knowledge_bases
SET name = $2,
    metadata = $3,
    updated_at = $4
WHERE id = $1
RETURNING *;

-- name: DeleteKnowledgeBase :execrows
DELETE FROM knowledge_bases
WHERE id = $1;
