# Knowledge Base & Retrieval System – High-Level Design

## Overview

This document outlines the high-level architecture and design principles for a **versioned, observable, retrieval-first knowledge base system** intended to support Retrieval-Augmented Generation (RAG) workloads.

The system is designed to:
- Treat **documents and their versions as first-class entities**
- Provide a **clean, stable Retrieval API** for agents and applications
- Support **incremental updates** without corrupting active retrieval state
- Enable **observability and evaluation** from the start
- Remain **technology-agnostic**, while favoring pragmatic, production-proven patterns

> **Developer preferences (non-binding):**
> - Comfortable working in Go and Next.js
> - Prefers Postgres databases where possible  
> - TODO: Explore and validate alternative storage and retrieval engines
> - OpenTelemetry should be supported from the beginning

---

## Core Concepts

### Knowledge Base
A **Knowledge Base (KB)** is an isolated, queryable corpus of documents.
- Multi-tenant systems may host many KBs
- Retrieval always happens within a single KB
- A KB is the unit of access control, indexing, and evaluation

### Document
A **Document** represents a logical source of knowledge.
- Identity is stable over time
- Metadata (path, title, tags) may change
- Content changes are tracked via versions

### Document Version
A **Document Version** represents a specific snapshot of document content.
- Versions are immutable once created
- Exactly one version is active at any given time
- Retrieval only operates on active versions

### Chunk
A **Chunk** is the atomic retrievable unit.
- Derived from a specific document version
- Has stable identifiers when content is unchanged
- Correlated back to its source document and version

---

## System Components

### 1. Retrieval API (Hybrid Search)

The Retrieval API is the primary interface consumed by:
- Agents
- Applications
- Internal evaluation tools

#### Responsibilities
- Accept natural language queries
- Apply semantic, keyword, or hybrid search strategies
- Enforce filters and scope
- Return ranked, citation-ready results
- Guarantee deterministic, stable responses for the same inputs

#### Key Properties
- Stateless
- Idempotent
- Low-latency
- Version-aware (active document versions only)

#### Notes
- Hybrid search should be a first-class concept
- Scoring should be normalized where possible
- Results should always include sufficient metadata for citation

---

### 2. Document Ingestion Pipeline

The ingestion pipeline is responsible for transforming raw user input into retrievable knowledge.

#### High-Level Flow
1. Document submission or update detected
2. Raw content stored in durable object storage
3. New document version created
4. Content extracted and normalized
5. Chunking performed
6. Embeddings generated or reused
7. New version marked active atomically

At no point should partially processed content be visible to retrieval.

---

### 2a. Document & Version Storage

#### Raw Document Storage
- Raw document bytes stored in a blob/object store
- Optional: extracted text or parse artifacts stored alongside raw files
- Storage optimized for size and durability, not queryability

#### Metadata Storage
The database stores:
- Document identity and metadata
- Version lineage and status
- Chunk records
- Embedding references
- Ingestion state and errors

#### Version Activation
- Only one version of a document may be active
- Activation occurs only after ingestion completes successfully
- Activation should be atomic and reversible

---

### 2b. Diffing, Invalidation, and Reuse

#### Change Detection
- Document updates detected via content hashes or source metadata
- If content hash unchanged, ingestion is skipped

#### Chunk-Level Diffing
- Each chunk computes a content hash
- Identical chunks across versions can reuse embeddings
- Changed chunks trigger new embeddings

#### Invalidation Rules
- Old versions are excluded from retrieval immediately upon activation of a new version
- Old chunks remain stored for audit/debug until garbage collected

#### Guarantees
- Retrieval never mixes chunks from different versions
- Updates are eventually consistent but never partially visible

---

### 3. Chunking & Embedding Strategies

Chunking and embedding are treated as **pluggable engines**, not fixed logic.

#### Chunking Responsibilities
- Transform extracted document content into retrievable units
- Preserve semantic and structural meaning
- Produce stable chunk boundaries where possible

#### Chunking Considerations
- Different strategies per document type
- Structural chunking preferred over naive token windows
- Chunk metadata should include source offsets or locators

#### Embedding Responsibilities
- Convert chunk content into vector representations
- Track embedding model identity and parameters
- Support future re-embedding workflows

#### Design Principle
Chunking and embedding strategies should evolve independently from:
- Retrieval API
- Storage layout
- Versioning semantics

---

## Observability & Telemetry (First-Class)

The system should emit structured telemetry from day one.

### Required Signals
- Traces for ingestion jobs and retrieval requests
- Metrics for latency, throughput, failures
- Logs with stable request and document identifiers

### Goals
- Debug ingestion failures
- Understand retrieval quality regressions
- Support evaluation and experimentation
- Enable future automated optimization

> OpenTelemetry-compatible instrumentation is strongly preferred.

---

## Post-MVP Considerations

### 1. Improving Knowledge Quality

The system should eventually help users improve their inputs.

Potential directions:
- Detect low-recall or low-coverage queries
- Identify documents that never surface in retrieval
- Surface chunk size or structure warnings
- Suggest document restructuring or metadata enrichment

This is advisory only — the system should not mutate user data automatically.

---

### 2. RAG Evaluation & Feedback Loops

To improve retrieval quality over time, the system needs evaluation primitives.

Potential components:
- Query + expected answer datasets
- Retrieval recall and precision metrics
- Human-in-the-loop relevance feedback
- Version-to-version regression comparisons

Evaluation should be:
- Isolated from production retrieval
- Repeatable
- Tied to specific KB snapshots

---

## Non-Goals (for now)

- Real-time streaming retrieval
- Cross-KB retrieval
- Automatic document rewriting
- Opinionated prompt orchestration

---

## Open Questions / TODOs

- What alternative storage engines should be explored beyond relational databases?
- How much version history should be retained by default?
- Should embeddings be globally deduplicated across KBs?
- What is the minimal viable evaluation loop for early users?

---

## Summary

This system is designed around a single core idea:

> **Knowledge changes over time — retrieval systems must make that change explicit, observable, and safe.**

Versioning, observability, and clean contracts are not add-ons; they are foundational.