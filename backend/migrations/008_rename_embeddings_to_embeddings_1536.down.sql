ALTER TABLE embeddings_1536 RENAME TO embeddings;

ALTER INDEX embeddings_1536_kb_id_idx RENAME TO embeddings_kb_id_idx;
ALTER INDEX embeddings_1536_content_hash_idx RENAME TO embeddings_content_hash_idx;
ALTER INDEX embeddings_1536_vector_idx RENAME TO embeddings_vector_idx;
ALTER TABLE embeddings RENAME CONSTRAINT embeddings_1536_dedupe_unique TO embeddings_dedupe_unique;
