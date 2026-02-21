# Document Ingestion Pipeline – Product Requirements Document

## Goal

Provide a **safe, observable, and versioned ingestion pipeline** that transforms raw user documents into retrievable knowledge without corrupting active retrieval state.

The pipeline must support a **markdown-first MVP**, while remaining extensible to additional document types in the future.

---

## Core Principles

- Versioning is mandatory
- Partial ingestion must never affect retrieval
- All operations must be observable
- Failures must be diagnosable and recoverable
- User documents must never be silently mutated

---

## Supported Document Types (MVP)

### Fully Supported (Processed)
- Markdown (`text/markdown`)
- Markdown variants (e.g. `.md`, `.mdx`)

These document types proceed through the full ingestion pipeline and are eligible for retrieval.

### Accepted but Not Processed (Stored-Only)
- PDFs
- Images
- Office documents (e.g. `.docx`)
- Other binary formats

These documents are:
- Stored durably
- Versioned
- Visible in the UI with explicit status
- Excluded from retrieval (no chunks generated)

---

## Functional Requirements

### 1. Document Intake
- Accept new documents and updates
- Assign stable document identity
- Persist raw content in durable storage
- Detect document type at intake time

Document intake must succeed regardless of whether the document type is fully supported.

---

### 2. Version Creation
- Each content change produces a new document version
- Versions are immutable
- Exactly one version is active per document
- Unsupported document types may still create versions, even if no processing occurs

---

### 3. Processing Stages

Ingestion must follow a defined state machine.

#### Standard Processing Path (Supported Types)
- RECEIVED
- STORED
- EXTRACTED
- CHUNKED
- EMBEDDED
- ACTIVATED

#### Stored-Only Path (Unsupported Types)
- RECEIVED
- STORED
- SKIPPED_UNSUPPORTED
- COMPLETED

Documents in `SKIPPED_UNSUPPORTED` state must never appear in retrieval results.

---

### 4. Extraction Stage

The extraction stage is responsible for producing a normalized, structured representation of document content suitable for chunking.

#### MVP Behavior
- Markdown extraction is deterministic and structure-preserving
- No OCR or layout inference is performed
- If extraction fails, ingestion halts safely

#### Design Constraint
The extraction stage must be **pluggable**, allowing future extractors for other document types without altering downstream logic.

---

### 5. Atomic Activation
- New versions become active only after successful completion
- Activation must be atomic
- Old versions must be immediately excluded from retrieval
- Stored-only documents never reach the activation stage

---

### 6. Error Handling
- Failures must not affect existing active versions
- Errors must be persisted with context
- Retrying must be safe and idempotent
- Unsupported document types are not treated as failures

---

## Diffing & Invalidation

### Change Detection
- Detect document changes via content hashes or source metadata
- Skip ingestion when content is unchanged

### Chunk-Level Reuse
- Reuse embeddings for unchanged chunks when possible
- Only changed chunks require re-embedding
- Stored-only documents do not participate in diffing or reuse

---

## Observability Requirements

- Trace ingestion jobs end-to-end
- Emit per-stage timing metrics
- Emit document-type and ingestion-path labels
- Surface failures and skipped states with actionable diagnostics
- Correlate ingestion runs with retrieval behavior

---

## MVP Scope

- Markdown-only full ingestion
- Stored-only handling for unsupported document types
- Manual ingestion triggers
- Full document reprocessing per update
- Version pointer-based activation
- Basic retry logic

---

## Post-MVP

- PDF and image extraction (OCR, layout-aware parsing)
- Incremental ingestion
- Partial document updates
- Priority queues
- Scheduled re-ingestion
