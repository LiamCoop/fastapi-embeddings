# Backend Scope (High-Level)

This backend is responsible for **retrieval and ingestion** for a versioned knowledge base system. It does **not** run the embedding model itself; embeddings are produced by an **external service**, which the backend calls to store vectors in the vector index.

## Backend Responsibilities
- **Ingestion orchestration**
  - Accept documents and updates.
  - Maintain document/version state machine (received → stored → extracted → chunked → embedded → activated).
  - Ensure atomic activation; retrieval never sees partial ingestion.
  - Call the external embedding service and persist vectors in the vector index.
  - Store unsupported document types as stored-only (excluded from retrieval).
- **Versioning guarantees**
  - Exactly one active version per document.
  - Retrieval operates only on active versions.
  - Chunk-level reuse by content hash where possible.
- **Retrieval API**
  - Deterministic, low-latency hybrid search (semantic + lexical).
  - Metadata filtering and stable ranking.
  - Citation-ready results with document, version, and location metadata.
- **Observability**
  - Traces, metrics, and logs for ingestion and retrieval.
  - Correlation IDs across pipeline stages and requests.
- **Evaluation hooks**
  - Support offline evaluation runs and metric storage (version-aware).

## External Dependencies
- **Embedding service** (external): receives chunk text and returns vectors.
- **Object storage** for raw documents and extracted artifacts.
- **Metadata store** for documents, versions, chunks, and ingestion state.
- **Vector index** for semantic retrieval.

## Non-Goals (Backend)
- Prompt orchestration or agent control loops.
- Retrieval-time LLM reasoning or reranking.
- Cross-knowledge-base retrieval.
