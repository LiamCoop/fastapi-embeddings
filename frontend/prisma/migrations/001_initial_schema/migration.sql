CREATE EXTENSION IF NOT EXISTS vector;

CREATE TABLE knowledge_bases (
    id uuid PRIMARY KEY,
    name text NOT NULL,
    metadata jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE documents (
    id uuid PRIMARY KEY,
    kb_id uuid NOT NULL REFERENCES knowledge_bases(id) ON DELETE CASCADE,
    path text NOT NULL,
    title text,
    document_type text NOT NULL,
    source_metadata jsonb NOT NULL DEFAULT '{}'::jsonb,
    active_version_id uuid,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT documents_kb_path_unique UNIQUE (kb_id, path)
);

CREATE TABLE document_versions (
    id uuid PRIMARY KEY,
    document_id uuid NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    kb_id uuid NOT NULL REFERENCES knowledge_bases(id) ON DELETE CASCADE,
    version_number integer NOT NULL,
    raw_content_uri text NOT NULL,
    extracted_content text,
    extracted_content_hash text,
    processing_status text NOT NULL,
    error_message text,
    is_active boolean NOT NULL DEFAULT false,
    created_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT document_versions_document_version_unique UNIQUE (document_id, version_number),
    CONSTRAINT document_versions_processing_status_check CHECK (
        processing_status IN (
            'RECEIVED',
            'STORED',
            'EXTRACTED',
            'CHUNKED',
            'EMBEDDED',
            'ACTIVATED',
            'SKIPPED_UNSUPPORTED',
            'FAILED'
        )
    )
);

ALTER TABLE documents
    ADD CONSTRAINT documents_active_version_fk
    FOREIGN KEY (active_version_id)
    REFERENCES document_versions(id)
    ON DELETE SET NULL;

CREATE TABLE embedding_models (
    id text PRIMARY KEY,
    name text NOT NULL,
    vector_dimension integer NOT NULL,
    provider text NOT NULL,
    parameters jsonb NOT NULL DEFAULT '{}'::jsonb,
    is_active boolean NOT NULL DEFAULT false,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE embeddings (
    id uuid PRIMARY KEY,
    kb_id uuid NOT NULL REFERENCES knowledge_bases(id) ON DELETE CASCADE,
    content_hash text NOT NULL,
    embedding_model_id text NOT NULL REFERENCES embedding_models(id) ON DELETE RESTRICT,
    embedding_vector vector(1536) NOT NULL,
    embedding_model_version text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT embeddings_dedupe_unique UNIQUE (kb_id, content_hash, embedding_model_id, embedding_model_version)
);

CREATE TABLE chunks (
    id uuid PRIMARY KEY,
    document_version_id uuid NOT NULL REFERENCES document_versions(id) ON DELETE CASCADE,
    kb_id uuid NOT NULL REFERENCES knowledge_bases(id) ON DELETE CASCADE,
    sequence_number integer NOT NULL,
    content text NOT NULL,
    content_hash text NOT NULL,
    metadata jsonb NOT NULL DEFAULT '{}'::jsonb,
    chunking_strategy text NOT NULL,
    embedding_id uuid REFERENCES embeddings(id) ON DELETE SET NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT chunks_version_sequence_unique UNIQUE (document_version_id, sequence_number)
);

CREATE TABLE retrieval_requests (
    id uuid PRIMARY KEY,
    kb_id uuid NOT NULL REFERENCES knowledge_bases(id) ON DELETE CASCADE,
    query text NOT NULL,
    filters jsonb NOT NULL DEFAULT '{}'::jsonb,
    top_k integer NOT NULL,
    hybrid_weight double precision NOT NULL,
    result_count integer NOT NULL,
    latency_ms bigint NOT NULL,
    empty_result boolean NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE retrieval_results (
    id uuid PRIMARY KEY,
    retrieval_request_id uuid NOT NULL REFERENCES retrieval_requests(id) ON DELETE CASCADE,
    chunk_id uuid NOT NULL REFERENCES chunks(id) ON DELETE CASCADE,
    rank integer NOT NULL,
    semantic_score double precision NOT NULL,
    lexical_score double precision NOT NULL,
    final_score double precision NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT retrieval_results_rank_unique UNIQUE (retrieval_request_id, rank)
);

CREATE TABLE ingestion_jobs (
    id uuid PRIMARY KEY,
    document_version_id uuid NOT NULL REFERENCES document_versions(id) ON DELETE CASCADE,
    kb_id uuid NOT NULL REFERENCES knowledge_bases(id) ON DELETE CASCADE,
    job_status text NOT NULL,
    error_message text,
    started_at timestamptz,
    completed_at timestamptz,
    CONSTRAINT ingestion_jobs_status_check CHECK (
        job_status IN ('QUEUED', 'IN_PROGRESS', 'SUCCESS', 'FAILED')
    )
);

CREATE TABLE processing_metrics (
    id uuid PRIMARY KEY,
    ingestion_job_id uuid NOT NULL REFERENCES ingestion_jobs(id) ON DELETE CASCADE,
    stage text NOT NULL,
    duration_ms bigint NOT NULL,
    item_count integer NOT NULL,
    status text NOT NULL,
    error_message text,
    created_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT processing_metrics_stage_check CHECK (
        stage IN ('RECEIVED', 'STORED', 'EXTRACTED', 'CHUNKED', 'EMBEDDED', 'ACTIVATED')
    )
);

CREATE TABLE evaluations (
    id uuid PRIMARY KEY,
    kb_id uuid NOT NULL REFERENCES knowledge_bases(id) ON DELETE CASCADE,
    name text NOT NULL,
    description text,
    status text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    completed_at timestamptz
);

CREATE TABLE evaluation_queries (
    id uuid PRIMARY KEY,
    evaluation_id uuid NOT NULL REFERENCES evaluations(id) ON DELETE CASCADE,
    query text NOT NULL,
    expected_chunk_ids jsonb NOT NULL DEFAULT '[]'::jsonb,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE evaluation_results (
    id uuid PRIMARY KEY,
    evaluation_id uuid NOT NULL REFERENCES evaluations(id) ON DELETE CASCADE,
    query_id uuid NOT NULL REFERENCES evaluation_queries(id) ON DELETE CASCADE,
    retrieved_chunk_ids jsonb NOT NULL DEFAULT '[]'::jsonb,
    recall_at_k double precision NOT NULL,
    precision_at_k double precision NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT evaluation_results_unique UNIQUE (evaluation_id, query_id)
);

CREATE UNIQUE INDEX document_versions_active_unique
    ON document_versions (document_id)
    WHERE is_active;

CREATE INDEX documents_kb_id_idx ON documents (kb_id);
CREATE INDEX documents_active_version_id_idx ON documents (active_version_id);
CREATE INDEX document_versions_kb_id_idx ON document_versions (kb_id);
CREATE INDEX document_versions_document_id_idx ON document_versions (document_id);
CREATE INDEX document_versions_processing_status_idx ON document_versions (processing_status);
CREATE INDEX chunks_document_version_id_idx ON chunks (document_version_id);
CREATE INDEX chunks_kb_id_idx ON chunks (kb_id);
CREATE INDEX chunks_embedding_id_idx ON chunks (embedding_id);
CREATE INDEX chunks_content_hash_idx ON chunks (content_hash);
CREATE INDEX embeddings_kb_id_idx ON embeddings (kb_id);
CREATE INDEX embeddings_content_hash_idx ON embeddings (content_hash);
CREATE INDEX retrieval_requests_kb_id_idx ON retrieval_requests (kb_id);
CREATE INDEX retrieval_requests_created_at_idx ON retrieval_requests (created_at);
CREATE INDEX retrieval_results_request_id_idx ON retrieval_results (retrieval_request_id);
CREATE INDEX retrieval_results_chunk_id_idx ON retrieval_results (chunk_id);
CREATE INDEX ingestion_jobs_document_version_id_idx ON ingestion_jobs (document_version_id);
CREATE INDEX ingestion_jobs_kb_id_idx ON ingestion_jobs (kb_id);
CREATE INDEX ingestion_jobs_status_idx ON ingestion_jobs (job_status);
CREATE INDEX processing_metrics_ingestion_job_id_idx ON processing_metrics (ingestion_job_id);
CREATE INDEX processing_metrics_stage_idx ON processing_metrics (stage);
CREATE INDEX evaluations_kb_id_idx ON evaluations (kb_id);
CREATE INDEX evaluation_queries_evaluation_id_idx ON evaluation_queries (evaluation_id);
CREATE INDEX evaluation_results_evaluation_id_idx ON evaluation_results (evaluation_id);

CREATE INDEX chunks_metadata_gin_idx ON chunks USING GIN (metadata);
CREATE INDEX documents_source_metadata_gin_idx ON documents USING GIN (source_metadata);
CREATE INDEX knowledge_bases_metadata_gin_idx ON knowledge_bases USING GIN (metadata);

CREATE INDEX chunks_content_tsv_idx ON chunks USING GIN (to_tsvector('english', content));

CREATE INDEX embeddings_vector_idx
    ON embeddings
    USING ivfflat (embedding_vector vector_cosine_ops);

CREATE UNIQUE INDEX embedding_models_active_unique
    ON embedding_models (is_active)
    WHERE is_active;
