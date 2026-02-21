DROP INDEX IF EXISTS embedding_models_active_unique;
DROP INDEX IF EXISTS embeddings_vector_idx;
DROP INDEX IF EXISTS chunks_content_tsv_idx;
DROP INDEX IF EXISTS knowledge_bases_metadata_gin_idx;
DROP INDEX IF EXISTS documents_source_metadata_gin_idx;
DROP INDEX IF EXISTS chunks_metadata_gin_idx;
DROP INDEX IF EXISTS evaluation_results_evaluation_id_idx;
DROP INDEX IF EXISTS evaluation_queries_evaluation_id_idx;
DROP INDEX IF EXISTS evaluations_kb_id_idx;
DROP INDEX IF EXISTS processing_metrics_stage_idx;
DROP INDEX IF EXISTS processing_metrics_ingestion_job_id_idx;
DROP INDEX IF EXISTS ingestion_jobs_status_idx;
DROP INDEX IF EXISTS ingestion_jobs_kb_id_idx;
DROP INDEX IF EXISTS ingestion_jobs_document_version_id_idx;
DROP INDEX IF EXISTS retrieval_results_chunk_id_idx;
DROP INDEX IF EXISTS retrieval_results_request_id_idx;
DROP INDEX IF EXISTS retrieval_requests_created_at_idx;
DROP INDEX IF EXISTS retrieval_requests_kb_id_idx;
DROP INDEX IF EXISTS embeddings_content_hash_idx;
DROP INDEX IF EXISTS embeddings_kb_id_idx;
DROP INDEX IF EXISTS chunks_content_hash_idx;
DROP INDEX IF EXISTS chunks_embedding_id_idx;
DROP INDEX IF EXISTS chunks_kb_id_idx;
DROP INDEX IF EXISTS chunks_document_version_id_idx;
DROP INDEX IF EXISTS document_versions_processing_status_idx;
DROP INDEX IF EXISTS document_versions_document_id_idx;
DROP INDEX IF EXISTS document_versions_kb_id_idx;
DROP INDEX IF EXISTS documents_active_version_id_idx;
DROP INDEX IF EXISTS documents_kb_id_idx;
DROP INDEX IF EXISTS document_versions_active_unique;

DROP TABLE IF EXISTS evaluation_results;
DROP TABLE IF EXISTS evaluation_queries;
DROP TABLE IF EXISTS evaluations;
DROP TABLE IF EXISTS processing_metrics;
DROP TABLE IF EXISTS ingestion_jobs;
DROP TABLE IF EXISTS retrieval_results;
DROP TABLE IF EXISTS retrieval_requests;
DROP TABLE IF EXISTS chunks;
DROP TABLE IF EXISTS embeddings;
DROP TABLE IF EXISTS embedding_models;

ALTER TABLE documents DROP CONSTRAINT IF EXISTS documents_active_version_fk;

DROP TABLE IF EXISTS document_versions;
DROP TABLE IF EXISTS documents;
DROP TABLE IF EXISTS knowledge_bases;

DROP EXTENSION IF EXISTS vector;
