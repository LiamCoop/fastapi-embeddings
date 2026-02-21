# Ragtime: Versioned RAG Knowledge Base System

## Project Mission
Provide a **versioned, observable, retrieval-first knowledge base system** for Retrieval-Augmented Generation (RAG) workloads. Users upload their knowledge base, and we provide a stable API for querying their corpus.

## Core Design Principle
> **Knowledge changes over time — retrieval systems must make that change explicit, observable, and safe.**

Versioning, observability, and clean contracts are not add-ons; they are foundational.

---
# Project instructions

## Linear usage
- Only use Linear MCP when explicitly asked OR when a task obviously requires creating/updating issues for planning.
- When creating a Linear issue, ALWAYS use one of the standard templates below (Feature, Bug, Debt, Spike).
- Keep descriptions short but complete; fill every section with at least 1 line (use "TBD" only if truly unknown).

## Issue naming
- Feature: "feat: <outcome>"
- Bug: "bug: <what breaks>"
- Debt: "debt: <cleanup>"
- Spike: "spike: <question>"

## Template bodies (copy/paste)
### Feature
Problem
    What user pain / business need?

Outcome
    What changes for the user when done?

Scope
    In: …
    Out: …

Acceptance criteria
    ...

Implementation notes
    ...
Links, API notes, UI notes
    ...

Risks / Unknowns
    ...

### Bug
Summary

Steps to reproduce

Expected

Actual

Impact

Severity: (low/med/high)

Frequency: (rare/sometimes/often)

Logs / screenshots

Suspected cause

Fix approach

Definition of done
    - repro fails before fix, passes after
    - tests added/updated



### Debt
Why now

Current state

Proposed change

Non-goals

Rollout plan

Acceptance criteria

    …

Risk / rollback

### Spike
Question

Why it matters

Constraints

Plan

Deliverable

e.g. short doc + recommendation + next tickets

Timebox

Exit criteria

decision made

follow-up issues created

## Splitting work
- If a Feature is > 1–2 days, create sub-issues:
  - "feat: ..." parent
  - children: "task: ..." with 3–6 acceptance checkboxes each
---

## Architecture Overview

### Technology Stack
- **Backend**: Go 1.22+ (stdlib http routing, no external router)
- **Database**: PostgreSQL 15+ with pgvector extension
- **Embedding Service**: Separate Python/FastAPI microservice
- **Object Storage**: Railway Buckets (S3-compatible API)
- **Observability**: OpenTelemetry Collector (logs, metrics, traces)
- **Processing Model**: Synchronous (MVP)

### Four-Layer Architecture (Repository Pattern)
```
┌─────────────────────────────────────────┐
│          HTTP Handlers Layer            │  ← Request/response handling
│  (stdlib http routes, request/response) │
└─────────────────────────────────────────┘
                    │
┌─────────────────────────────────────────┐
│          Service Layer                  │  ← Business logic & orchestration
│  - DocumentService                      │
│  - RetrievalService                     │
│  - IngestionService                     │
│  - ChunkingService                      │
│  - EmbeddingService (HTTP client)       │
└─────────────────────────────────────────┘
                    │
┌─────────────────────────────────────────┐
│        Repository Layer                 │  ← Database access (sqlc)
│  - DocumentRepository                   │
│  - ChunkRepository                      │
│  - EmbeddingRepository                  │
│  - RetrievalRepository                  │
└─────────────────────────────────────────┘
                    │
┌─────────────────────────────────────────┐
│     Infrastructure Layer                │  ← External dependencies
│  - PostgreSQL (pgvector)                │
│  - Railway Buckets (S3-compatible)      │
│  - FastAPI Embedding Service (HTTP)     │
│  - OTEL Collector (observability)       │
└─────────────────────────────────────────┘
```

---

## Core Concepts

### Knowledge Base (KB)
- An isolated, queryable corpus of documents
- The unit of access control, indexing, and evaluation
- Retrieval always happens within a single KB (no cross-KB queries in MVP)

### Document
- A logical source of knowledge with stable identity over time
- Metadata (path, title, tags) may change
- Content changes are tracked via versions

### Document Version
- A specific snapshot of document content (immutable once created)
- Exactly one version is active at any given time
- **Retrieval only operates on active versions**
- State machine: RECEIVED → STORED → EXTRACTED → CHUNKED → EMBEDDED → ACTIVATED

### Chunk
- The atomic retrievable unit derived from a specific document version
- Has stable identifiers when content is unchanged (enables embedding reuse)
- Always correlated back to source document and version

### Embedding
- Vector representation stored in dimension-specific tables (e.g., `embeddings_1536`, `embeddings_3072`)
- Reused via content_hash when chunks are unchanged
- Tracked by model identity and version

---

## Key Architectural Decisions

### MVP Constraints (Intentional Simplifications)
1. **Chunking**: Structural chunking only (based on headers, sections, code blocks)
   - Deterministic and stable
   - Preserves semantic structure
   - Better citations and lower re-embedding churn
   - No fixed token windows or semantic chunking in MVP

2. **Embeddings**: Single global embedding model (text-embedding-3-small)
   - Simplest operational model
   - Easier evaluation
   - Compatible with chunk reuse via hashing
   - No per-KB or per-document models in MVP

3. **Processing**: Synchronous ingestion pipeline
   - Simpler error handling and observability
   - Async processing deferred to post-MVP

4. **Document Types**: Markdown-first MVP
   - **Fully supported (processed)**: Markdown (`.md`, `.mdx`)
   - **Accepted but not processed (stored-only)**: PDFs, images, Office docs
   - Stored-only documents skip EXTRACTED/CHUNKED/EMBEDDED stages (status: SKIPPED_UNSUPPORTED)

### Critical Safety Guarantees
- **Partial ingestion never affects retrieval** - only complete, activated versions are queryable
- **Activation is atomic** - version switches happen in a single transaction
- **Old versions excluded immediately** - retrieval never mixes chunks from different versions
- **User documents never silently mutated** - system may recommend, but user must decide
- **Failures preserve existing state** - old version remains active on failure

---

## Explicitly Deferred Features (Post-MVP)
See `docs/not-to-do-list.md` for full details. Key deferrals:
- Multiple chunking strategies
- Multiple embedding strategies
- Manual chunk intervention UI
- Cross-knowledge-base retrieval
- Automatic document rewriting
- Real-time/streaming retrieval
- Retrieval-time LLM reasoning

**Principle**: MVP prioritizes correctness, stability, and observability over flexibility.

---

## Ingestion Pipeline Flow (Synchronous)

```
1. Client uploads document 
2. Create Document + DocumentVersion (status=RECEIVED)
3. Upload raw bytes to Railway Buckets (status=STORED)
4. Type check → If unsupported: SKIPPED_UNSUPPORTED, exit
5. Extract content, compute hash (status=EXTRACTED)
   └─> If duplicate hash: reuse existing chunks/embeddings, jump to activation
6. Structural chunking (status=CHUNKED)
7. Generate embeddings via $EMBEDDING_PROVIDER
    - Starting with local first / fastapi-embeddings
    - May use openAI or another provider later on after we progress.
    - reuse by content_hash (status=EMBEDDED)
8. ATOMIC ACTIVATION:
   └─> Set old version is_active=false
   └─> Set new version is_active=true
   └─> Update Document.active_version_id
   └─> Commit transaction (status=ACTIVATED)
9. Return HTTP 201 Created
```

**Error Handling**: On failure, record error_message, set status=FAILED, rollback transaction. Old version remains active.

---

## Retrieval API

### Hybrid Search (Semantic + Lexical)
1. Embed query via FastAPI
2. Semantic search: pgvector similarity (cosine distance)
3. Lexical search: Postgres full-text search (tsvector/tsquery)
4. Blend scores: `final_score = (hybrid_weight * semantic) + ((1 - hybrid_weight) * lexical)`
5. Apply metadata filters (path, tags, timestamps)
6. Return top-K with citations

### Guarantees
- Stateless and idempotent
- Deterministic ordering for identical inputs
- Only returns chunks from active document versions
- Every result includes citation metadata (document path, version, section, offsets)

---

## Observability (First-Class)

### Required Signals
- **Traces**: Ingestion jobs and retrieval requests (OpenTelemetry)
- **Metrics**: Latency, throughput, failures, embedding reuse rate
- **Logs**: Structured JSON with stable request and document identifiers

### Observability Entities
- `IngestionJob`: Tracks document processing (QUEUED → IN_PROGRESS → SUCCESS/FAILED)
- `ProcessingMetric`: Per-stage timing and item counts (STORED, EXTRACTED, CHUNKED, EMBEDDED, ACTIVATED)
- `RetrievalRequest`: Query, filters, top_k, latency_ms, result_count
- `RetrievalResult`: Per-chunk scores (semantic, lexical, final) and rankings

**Goal**: Debug ingestion failures, understand retrieval quality regressions, support evaluation.

---

## Development Guidelines

### When Writing Code
- **Read before modifying**: Always read files before suggesting changes
- **Follow repository pattern**: Keep handlers, services, repositories, and domain models separate
- **Maintain layer boundaries**: Handlers call services, services call repositories, repositories use sqlc
- **Preserve guarantees**: Never expose partially ingested content to retrieval
- **Deterministic behavior**: Same input must produce same output (chunking, hashing)
- **Error context**: Include document_id, version_id, stage in all error messages
- **Observability**: Emit traces and metrics at service layer boundaries

### Security Considerations
- Validate user input at handler layer
- Use parameterized queries (sqlc handles this)
- Never trust document content (sanitize before embedding/storing)
- Respect KB isolation boundaries (no cross-KB queries)

### Testing Priorities
1. Chunking determinism (same input = same chunks)
2. Score blending logic
3. Atomic activation (version switch integrity)
4. Embedding reuse via content_hash
5. Retrieval version-safety (only active versions returned)

---

## Evaluation (Minimal Schema - Not Implemented Yet)

Database schema created for future use:
- `Evaluation`: Named evaluation runs with status
- `EvaluationQuery`: Query + expected chunks (ground truth)
- `EvaluationResult`: Retrieval results + Recall@K, Precision@K

**No API endpoints or business logic in MVP** - schema exists for future extension.

---

## References

For detailed information, see:
- `docs/OVERVIEW.md` - High-level design and core concepts
- `docs/PLAN.md` - Implementation plan, ER diagram, API structure
- `docs/not-to-do-list.md` - Explicitly deferred features
- `docs/chunking-ard.md` - Chunking strategy decision
- `docs/embedding-ard.md` - Embedding strategy decision
- `docs/ingestion-prd.md` - Ingestion pipeline requirements
- `docs/retrieval-prd.md` - Retrieval API requirements
- `docs/evaluation-ard.md` - Evaluation framework (post-MVP)
- `docs/QUERY_OPTIMIZATION.md` - pgvector index strategy for multi-tenant vector search and KB isolation
- `docs/dashboard-page-hierarchy-area-naming.md` - Frontend page hierarchy, area naming, and user mental model for the dashboard
- `docs/product-vibes-interaction-direction.md` - Product emotional tone, interaction principles, and UX direction
- `docs/progressive-configuration-vision.md` - Philosophy for progressive disclosure from zero-config to power-user control
