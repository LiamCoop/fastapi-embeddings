-- name: GetDocumentByKBPath :one
SELECT * FROM documents
WHERE kb_id = $1 AND path = $2;

-- name: InsertDocument :one
INSERT INTO documents (
    id,
    kb_id,
    path,
    title,
    document_type,
    source_metadata,
    created_at,
    updated_at
) VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7,
    $8
)
RETURNING *;

-- name: UpdateDocument :one
UPDATE documents
SET title = $2,
    document_type = $3,
    source_metadata = $4,
    updated_at = $5
WHERE id = $1
RETURNING *;

-- name: InsertDocumentVersion :one
INSERT INTO document_versions (
    id,
    document_id,
    kb_id,
    version_number,
    raw_content_uri,
    processing_status,
    error_message,
    is_active,
    created_at
) VALUES (
    $1,
    $2,
    $3,
    (SELECT COALESCE(MAX(version_number), 0) + 1 FROM document_versions WHERE document_id = $2),
    $4,
    $5,
    $6,
    $7,
    $8
)
RETURNING *;

-- name: UpdateDocumentVersionStatus :exec
UPDATE document_versions
SET processing_status = $2,
    error_message = $3
WHERE id = $1;

-- name: ActivateDocumentVersion :exec
BEGIN;
UPDATE document_versions SET is_active = false
WHERE document_id = (SELECT document_id FROM document_versions WHERE id = $1);

UPDATE document_versions SET is_active = true, processing_status = 'ACTIVATED'
WHERE id = $1;

UPDATE documents SET active_version_id = $1, updated_at = now()
WHERE id = (SELECT document_id FROM document_versions WHERE id = $1);
COMMIT;
