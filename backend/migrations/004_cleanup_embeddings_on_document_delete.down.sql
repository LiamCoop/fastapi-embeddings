DROP TRIGGER IF EXISTS chunks_cleanup_orphan_embedding_after_delete ON chunks;
DROP FUNCTION IF EXISTS cleanup_orphan_embedding_after_chunk_delete();
