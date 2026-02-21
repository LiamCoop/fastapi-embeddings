# Explicitly Deferred Features (Post-MVP “Not-To-Do” List)

## Purpose

This document enumerates features that are **intentionally out of scope for MVP**.

These items are:
- Valuable
- Likely requested
- Architecturally anticipated

…but are deferred to:
- Preserve MVP focus
- Avoid premature abstraction
- Ensure core correctness and observability first

Any addition of these features **requires explicit product and architecture review**.

---

## 1. Multiple Chunking Strategies

### Description
Support for selecting or mixing different chunking strategies, such as:
- Fixed token windows
- Semantic chunking
- Per-document or per-KB chunking configuration

### Why It’s Deferred
- Increases surface area for bugs
- Complicates evaluation and comparability
- Makes chunk stability guarantees harder
- Adds UI and API complexity early

### Architectural Posture
- Chunking engine is pluggable
- Chunk metadata supports strategy attribution
- MVP uses a single, deterministic strategy

---

## 2. Multiple Embedding Strategies

### Description
Support for:
- Multiple embedding models
- Per-KB or per-document embeddings
- Hybrid or multi-vector embeddings

### Why It’s Deferred
- Significantly complicates evaluation
- Introduces migration and re-embedding complexity
- Obscures early retrieval quality signals

### Architectural Posture
- Embeddings are versioned and model-attributed
- Schema supports coexistence of multiple models
- MVP operates with a single global embedding strategy

---

## 3. Manual Chunk Intervention UI

### Description
User-facing tools to:
- Manually split or merge chunks
- Edit chunk boundaries
- Override chunk text or metadata

### Why It’s Deferred
- Breaks determinism
- Introduces human error into retrieval
- Complicates diffing and reuse semantics
- Requires extensive UX and guardrails

### Architectural Posture
- Chunk boundaries are machine-generated
- Chunk lineage is preserved for audit/debug
- Future UI can layer on top without breaking core contracts

---

## 4. Fine-Grained Chunk Invalidation

### Description
Invalidate or re-embed:
- Individual chunks
- Partial document sections
- Chunks affected by upstream changes

### Why It’s Deferred
- Adds complexity to versioning logic
- Risky without strong evaluation tooling
- Premature optimization for most document sizes

### Architectural Posture
- Document-version-level invalidation is primary
- Chunk-level reuse is opportunistic, not guaranteed
- Invalidation is coarse by design in MVP

---

## 5. Cross-Knowledge-Base Retrieval

### Description
Query across multiple knowledge bases simultaneously.

### Why It’s Deferred
- Complicates scoring normalization
- Blurs ownership and access boundaries
- Harder to reason about evaluation and relevance

### Architectural Posture
- KB isolation is strict in MVP
- Retrieval API enforces single-KB scope
- Cross-KB search can be layered later if needed

---

## 6. Automatic Document Rewriting or Optimization

### Description
Automatically:
- Rewrite documents for better chunking
- Normalize or summarize content
- Inject metadata or structure

### Why It’s Deferred
- Alters user data without explicit consent
- Hard to evaluate correctness
- High trust and UX risk

### Architectural Posture
- System may *recommend* improvements
- All content mutations remain user-controlled

---

## 7. Real-Time or Streaming Retrieval

### Description
- Streaming partial retrieval results
- Progressive ranking updates

### Why It’s Deferred
- Adds latency and consistency complexity
- Not required for most agent use cases
- Difficult to observe and debug

### Architectural Posture
- Retrieval is request/response
- Latency optimization focuses on predictable p95

---

## 8. Retrieval-Time LLM Reasoning

### Description
Using LLMs inside the retrieval path to:
- Rewrite queries
- Re-rank results
- Perform semantic reasoning inline

### Why It’s Deferred
- Increases cost and latency
- Obscures retrieval correctness
- Complicates reproducibility and evaluation

### Architectural Posture
- Retrieval is deterministic and model-free
- Agents handle reasoning externally

---
---

## 9. Assisted Document Rewriting with Diff Review UI

### Description
System-generated suggestions that propose changes to user documents in order to:
- Improve chunk stability
- Increase retrieval recall
- Reduce redundancy
- Clarify structure or metadata

All proposed changes must be presented as **diffs**, never applied automatically.

---

### Core Requirements (When Implemented)

- Show a clear, side-by-side or inline diff:
  - Original content
  - Suggested changes
- Explain *why* each change is recommended:
  - Retrieval coverage improvements
  - Chunk boundary stability
  - Reduced embedding churn
  - Improved citation fidelity
- Allow users to:
  - Accept changes
  - Reject changes
  - Manually resolve conflicts
- Preserve original document content unless explicitly replaced by the user

---

### Why It’s Deferred

- Requires high trust and strong UX
- Depends on reliable evaluation signals
- Risky without robust observability and rollback
- Premature before users understand baseline system behavior

---

### Architectural Posture

- The system **must never mutate user documents automatically**
- All recommendations are advisory
- Suggested changes are versioned and auditable
- Document versioning model already supports this cleanly:
  - Suggested edits → proposed new version
  - User acceptance → activation
  - Rejection → no state change

---

### Design Principle

> **The system may recommend, but the user must decide.**

Explainability and user control are mandatory.
---

## Guiding Principle

> **MVP prioritizes correctness, stability, and observability over flexibility.**

Flexibility is intentionally postponed until:
- Retrieval quality is measurable
- Failure modes are understood
- Core contracts have proven durable

---

## Reconsideration Criteria

Deferred features may be reconsidered when:
- Evaluation data demonstrates a clear need
- Observability shows stable system behavior
- Product demand justifies added complexity
