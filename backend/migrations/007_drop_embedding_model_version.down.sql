ALTER TABLE embeddings DROP CONSTRAINT embeddings_dedupe_unique;
ALTER TABLE embeddings ADD COLUMN embedding_model_version text NOT NULL DEFAULT '';
ALTER TABLE embeddings ADD CONSTRAINT embeddings_dedupe_unique UNIQUE (kb_id, content_hash, embedding_model_id, embedding_model_version);
