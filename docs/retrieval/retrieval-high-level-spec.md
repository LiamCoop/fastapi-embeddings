# Retrieval API Design for LLM-Agnostic RAG Systems

## Overview

This document describes the key design principles and response structures for a retrieval-focused RAG platform where:

- The platform owns ingestion, chunking, indexing, retrieval, and grounding
- Users bring their own LLM
- The platform exposes a Query API that returns structured, citable knowledge context

The goal is to provide high-quality, reproducible, and debuggable retrieval that integrates cleanly into any LLM workflow (agents, tools, local models, hosted models, MCP servers, etc.).

**This is retrieval as infrastructure, not answer generation.**

---

## Core Design Principles

### 1. Return Structured Knowledge, Not Just Text

A retrieval API should return citable passages with metadata, not raw text blobs. This enables:

- Citation generation
- Debugging retrieval failures
- Deterministic reproduction
- Selective hydration of context

### 2. Stable Identity and Versioning

Retrieval results must be traceable across time. Key requirements:

- Stable `chunk_id`
- Document-level version identifiers
- Index version identifiers
- Ability to reproduce previous retrieval results

This prevents "silent drift" when knowledge bases change.

### 3. Retrieval Explainability

Consumers need visibility into how results were produced. Useful signals include:

- Retrieval strategy used
- Filters applied
- Reranking enabled or not
- Scoring signals

This is critical for evaluation and debugging.

### 4. Separation of Search vs Hydration

Retrieval is typically a two-step process:

1. **Search** — return ranked references and snippets
2. **Hydrate** — fetch full chunk content or expanded context

This improves latency, reduces token waste, and supports agentic workflows.

---

## Recommended Retrieval Features

### Hybrid Retrieval

Dense-only retrieval is insufficient for many domains (code, APIs, proper nouns). Recommended support:

- Dense vector retrieval
- Sparse keyword retrieval
- Hybrid fusion
- Optional reranking

Configurable per index or query.

### Metadata Filtering and Routing

Retrieval quality depends heavily on metadata.

**Index-time metadata:**

- Tenant / workspace ID
- Project or repository ID
- Document path
- Document type
- Language (for codebases)
- Creation / update timestamps
- User-defined tags

**Query-time filters:**

- Project scoping
- Path prefix filtering
- Tag filtering
- Temporal filtering

### Chunk Stitching and Context Expansion

Individual chunks are often insufficient. Helpful retrieval behaviors:

- Adjacent chunk expansion
- Section-level grouping
- Breadcrumb preservation
- Deduplication of overlapping content

### Deterministic Chunk Identity

Each chunk should have:

- Stable content-derived identifier
- Deterministic position within a document
- Reproducible metadata

This enables citation persistence and reproducibility across re-indexing.

---

## Recommended Query API Response Shape

### Required Passage Fields

Each retrieval result should include:

| Field | Description |
|---|---|
| `chunk_id` | Stable unique identifier |
| `document_id` | Source document reference |
| `source_uri` | File path, URL, or repo reference |
| `title` | Document or section title |
| `section_path` | Hierarchical breadcrumb |
| `text` | Snippet or partial content |
| `score` | Retrieval relevance score |
| `offsets` | Optional line or character ranges |

### Document-Level Grouping

Grouping passages by document improves downstream context assembly. Groups may include:

- Document metadata
- Best passages from that document
- Document-level relevance score

### Retrieval Metadata

Include optional diagnostic metadata such as:

- Retrieval strategy used
- Reranker usage
- Filters applied
- Latency metrics

### Index and Document Versioning

Top-level response fields should include:

- `query_id`
- `index_version`
- Per-document version identifiers

This supports debugging and auditability.

---

## Recommended Endpoints

### 1. Query Endpoint

**Purpose:** Retrieve ranked chunk references and snippets.

Responsibilities:
- Hybrid retrieval
- Filtering
- Reranking
- Lightweight snippet return

### 2. Hydration Endpoint

**Purpose:** Fetch full chunk content or expanded context given IDs.

Responsibilities:
- Chunk expansion
- Adjacent chunk stitching
- Returning full metadata

### 3. Ingestion Endpoint

**Purpose:** Accept documents and metadata for indexing.

Should support:
- Content ingestion
- Metadata attachment
- Idempotent updates
- Incremental re-indexing

### 4. Index Management Endpoints

Optional but valuable:
- Delete documents
- Rebuild indexes
- Index statistics
- Indexing status

---

## Integration Patterns with LLM Tooling

### Retriever Tool

LLM calls the Query API directly and receives structured context.

### Citation-Grounded Prompting

Returned passages are injected into prompts with enforced citation behavior.

### Agentic Retrieval Loops

Agents perform iterative search → read → search cycles using query + hydration endpoints.

---

## Common Pitfalls to Avoid

- Returning large unstructured text blobs
- Lack of stable chunk identity
- Absence of versioning
- No metadata filtering
- Dense-only retrieval strategies
- Mixing retrieval with generation responsibilities

---

## Strategic Positioning

The retrieval platform should aim to be:

- **Deterministic** — same inputs produce same outputs
- **Explainable** — consumers can understand how results were produced
- **Multi-tenant safe** — strict KB isolation
- **Composable** — integrates cleanly with any LLM stack
- **Optimized for debugging and evaluation** — not just raw performance

The value lies not in embeddings themselves, but in reliable context construction for reasoning systems.

---

## Future Extensions

- Reranker-as-a-service
- Query rewriting
- Retrieval evaluation tooling
- Answerability detection
- Retrieval observability dashboards
- Structured citation graphs
