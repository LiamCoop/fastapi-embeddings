ALTER TABLE embeddings_384 DROP CONSTRAINT embeddings_384_dedupe_unique;
ALTER TABLE embeddings_384 ADD COLUMN embedding_model_version text NOT NULL DEFAULT '';
ALTER TABLE embeddings_384 ADD CONSTRAINT embeddings_384_dedupe_unique UNIQUE (kb_id, content_hash, embedding_model_id, embedding_model_version);
