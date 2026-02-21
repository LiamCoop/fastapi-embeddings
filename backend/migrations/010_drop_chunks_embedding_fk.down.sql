ALTER TABLE chunks ADD CONSTRAINT chunks_embedding_id_fkey FOREIGN KEY (embedding_id) REFERENCES embeddings_1536(id) ON DELETE SET NULL;
