CREATE OR REPLACE FUNCTION cleanup_orphan_embedding_after_chunk_delete()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
    IF OLD.embedding_id IS NOT NULL THEN
        DELETE FROM embeddings e
        WHERE e.id = OLD.embedding_id
          AND NOT EXISTS (
              SELECT 1
              FROM chunks c
              WHERE c.embedding_id = e.id
          );
    END IF;

    RETURN OLD;
END;
$$;

CREATE TRIGGER chunks_cleanup_orphan_embedding_after_delete
AFTER DELETE ON chunks
FOR EACH ROW
EXECUTE FUNCTION cleanup_orphan_embedding_after_chunk_delete();

-- Remove stale orphan embeddings that may already exist.
DELETE FROM embeddings e
WHERE NOT EXISTS (
    SELECT 1
    FROM chunks c
    WHERE c.embedding_id = e.id
);
