# Retrieval API Design for LLM-Agnostic RAG Systems

## Overview

This document describes the key design principles, API surface, and response structures for a retrieval-focused RAG platform where:

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

- Retrieval strategy used (profile and effective weights)
- Filters applied
- Reranking enabled or not
- Per-result scoring signals (semantic, lexical, final)

This is critical for evaluation and debugging.

### 4. Separation of Search vs Hydration

Retrieval is typically a two-step process:

1. **Search** — return ranked references and snippets
2. **Hydrate** — fetch full chunk content or expanded context

This improves latency, reduces token waste, and supports agentic workflows.

---

## Retrieval Features

### Hybrid Retrieval

Dense-only retrieval is insufficient for many domains (code, APIs, proper nouns). The platform supports:

- Dense vector retrieval (semantic)
- Sparse keyword retrieval (lexical / BM25)
- Hybrid fusion with configurable weighting
- Optional reranking

Different queries benefit from different retrieval strategies:

| Query type | Best strategy |
|---|---|
| Identifiers, error strings, code symbols | Lexical-heavy |
| Conceptual / explanatory questions | Semantic-heavy |
| Mixed natural language with domain terms | Hybrid |
| Short ambiguous queries | Lexical anchoring + semantic expansion |

Configurable per index or query via **retrieval profiles** (see Query Parameters section).

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

## Query API Design

### Query Parameters

```ts
// Retrieval control
retrieval_profile?: "auto" | "exact" | "balanced" | "semantic"
semantic_weight?: number   // optional override, 0–1

// Filtering
filters?: {
  project_id?: string
  path_prefix?: string
  tags?: string[]
  updated_after?: string  // ISO 8601
}

// Output control
top_k?: number
debug?: boolean
```

**Rules:**
- `retrieval_profile` defaults to `"auto"`
- `semantic_weight` overrides the profile if present
- If both omitted → index default applies

### Retrieval Profiles

#### `auto` (default)

System chooses weighting using lightweight query analysis.

Target behavior:
- Correct most cases without user involvement
- Provide observability into decisions
- Adapt to code-heavy vs prose-heavy corpora

**Signals increasing lexical weight:**
- Quoted phrases
- Symbols (`/`, `.`, `_`, `::`, `->`)
- File paths, `camelCase` / `snake_case` tokens
- Error-like patterns, numbers, version strings
- Short queries

**Signals increasing semantic weight:**
- Question form (how / why / when / what)
- Long natural language queries
- Abstract nouns, conversational phrasing
- Absence of rare tokens

Mixed signals fallback to `balanced`.

#### `exact`

Lexical-dominant retrieval. Intended for code, logs, file paths, config keys, stack traces, identifiers.

Typical internal config:
- Lexical weight ≈ 0.75–0.9
- Phrase boosting enabled
- Minimal semantic expansion
- Aggressive metadata filtering

#### `balanced`

True hybrid retrieval. Intended for documentation, knowledge bases, mixed queries, onboarding material.

Typical internal config:
- Lexical weight ≈ 0.4–0.6
- Semantic expansion allowed
- Reranker recommended

#### `semantic`

Dense retrieval dominant. Intended for exploratory questions, summarization queries, conceptual reasoning, research content.

Typical internal config:
- Lexical weight ≈ 0.1–0.3
- Query expansion enabled
- Higher candidate pool for reranking

### Numeric Weight Override

Advanced callers may specify `semantic_weight ∈ [0, 1]`:

```
final_score =
  semantic_weight * dense_score +
  (1 - semantic_weight) * lexical_score
```

Notes:
- Scores must be normalized prior to fusion
- Allow per-index calibration
- Expose effective weight in debug metadata

### Retrieval Pipeline

```
query
 → query analysis (auto mode only)
 → lexical retrieval (BM25 / inverted index)
 → dense retrieval (vector search)
 → score normalization
 → weighted fusion
 → metadata filtering
 → optional rerank
 → context assembly
```

---

## Response Shape

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
| `score` | Final retrieval relevance score |
| `score_detail` | `{ semantic, lexical, final }` per-signal breakdown |
| `offsets` | Optional line or character ranges |

### Document-Level Grouping

Grouping passages by document improves downstream context assembly. Groups include:

- Document metadata
- Best passages from that document
- Document-level relevance score

### Retrieval Metadata

Standard response includes:

```json
{
  "query_id": "...",
  "index_version": "...",
  "latency_ms": 42
}
```

When `debug: true`, additionally includes:

```json
{
  "retrieval_profile_effective": "balanced",
  "semantic_weight_effective": 0.5,
  "auto_signals_detected": [...],
  "lexical_candidates": 20,
  "semantic_candidates": 20,
  "reranker_applied": true,
  "filters_applied": {...}
}
```

This enables developer trust, offline evaluation, default tuning, and failure diagnosis.

---

## Endpoints

### 1. Query Endpoint

**Purpose:** Retrieve ranked chunk references and snippets.

Responsibilities:
- Hybrid retrieval with profile-controlled weighting
- Metadata filtering
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

## Index-Level Defaults

Indexes may define default retrieval behavior:

```
default_retrieval_profile
default_semantic_weight
```

| Corpus type | Default profile |
|---|---|
| Code repository | `exact` |
| Product documentation | `balanced` |
| Research corpus | `semantic` |

Caller request overrides index defaults.

---

## Integration Patterns with LLM Tooling

### Retriever Tool

LLM calls the Query API directly and receives structured context with scoring metadata.

### Citation-Grounded Prompting

Returned passages are injected into prompts with enforced citation behavior using `chunk_id`, `source_uri`, and `section_path`.

### Agentic Retrieval Loops

Agents perform iterative search → read → search cycles using query + hydration endpoints. Agents benefit from:
- Deterministic retrieval modes per profile
- Retry strategies across profiles
- Fast lexical probing before semantic search
- Debug metadata for reasoning traces

Common agent pattern:

```
agent → query(profile=exact)   // fast grounding check
agent → query(profile=balanced) // broader recall
agent → hydrate(chunk_ids)      // expand winning chunks
agent → synthesize
```

---

## Failure Modes & Safeguards

### Over-Semantic Drift

Mitigations:
- Lexical anchoring in auto mode
- Reranker enforcement
- Semantic expansion limits

### Over-Lexical Brittleness

Mitigations:
- Semantic fallback when lexical recall is low
- Synonym expansion

### Score Incompatibility

Mitigations:
- Per-index normalization calibration
- Fusion scaling

---

## Common Pitfalls to Avoid

- Returning large unstructured text blobs
- Lack of stable chunk identity
- Absence of versioning
- No metadata filtering
- Dense-only retrieval strategies
- Mixing retrieval with generation responsibilities
- Exposing raw weights without profile abstractions (cognitive overload)

---

## Strategic Positioning

The retrieval platform should aim to be:

- **Deterministic** — same inputs produce same outputs
- **Explainable** — consumers can understand how results were produced, including effective retrieval weights
- **Multi-tenant safe** — strict KB isolation
- **Composable** — integrates cleanly with any LLM stack
- **Optimized for debugging and evaluation** — not just raw performance

The value lies not in embeddings themselves, but in reliable, controllable context construction for reasoning systems.

---

## Future Extensions

- Reranker-as-a-service
- Query rewriting
- Retrieval evaluation tooling
- Answerability detection
- Retrieval observability dashboards
- Structured citation graphs
- Per-KB retrieval profile tuning via evaluation feedback
