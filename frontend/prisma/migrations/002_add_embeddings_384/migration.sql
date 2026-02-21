CREATE TABLE embeddings_384 (
    id uuid PRIMARY KEY,
    kb_id uuid NOT NULL REFERENCES knowledge_bases(id) ON DELETE CASCADE,
    content_hash text NOT NULL,
    embedding_model_id text NOT NULL REFERENCES embedding_models(id) ON DELETE RESTRICT,
    embedding_vector vector(384) NOT NULL,
    embedding_model_version text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT embeddings_384_dedupe_unique UNIQUE (kb_id, content_hash, embedding_model_id, embedding_model_version)
);

CREATE INDEX embeddings_384_kb_id_idx ON embeddings_384 (kb_id);
CREATE INDEX embeddings_384_content_hash_idx ON embeddings_384 (content_hash);

CREATE INDEX embeddings_384_vector_idx
    ON embeddings_384
    USING ivfflat (embedding_vector vector_cosine_ops);

INSERT INTO embedding_models (id, name, vector_dimension, provider, parameters, is_active)
VALUES (
    'all-MiniLM-L6-v2',
    'all-MiniLM-L6-v2',
    384,
    'fastapi',
    '{}'::jsonb,
    true
)
ON CONFLICT (id) DO UPDATE
SET
    name = EXCLUDED.name,
    vector_dimension = EXCLUDED.vector_dimension,
    provider = EXCLUDED.provider,
    parameters = EXCLUDED.parameters,
    is_active = EXCLUDED.is_active;
