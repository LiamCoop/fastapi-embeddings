DELETE FROM embedding_models WHERE id = 'all-MiniLM-L6-v2';

DROP INDEX IF EXISTS embeddings_384_vector_idx;
DROP INDEX IF EXISTS embeddings_384_content_hash_idx;
DROP INDEX IF EXISTS embeddings_384_kb_id_idx;

DROP TABLE IF EXISTS embeddings_384;
