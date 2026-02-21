# Embedding Strategies – Architecture Decision Record

## Problem

Chunks must be embedded into vector space to support semantic retrieval, while allowing future evolution of models and strategies.

---

## Options Considered

### Option A: Single Global Embedding Model
**Pros**
- Simple
- Predictable
- Easy caching and reuse

**Cons**
- Suboptimal for diverse content types

---

### Option B: Per-Knowledge-Base Models
**Pros**
- Tunable per domain
- Better relevance

**Cons**
- Operational complexity
- Harder migrations

---

### Option C: Per-Chunk-Type Models
**Pros**
- Best theoretical quality

**Cons**
- Very complex
- Difficult evaluation

---

## MVP Decision

**Selected: Option A – Single Global Embedding Model**

### Rationale
- Simplest operational model
- Easier evaluation
- Compatible with chunk reuse via hashing
- Allows clean future migrations

---

## MVP Constraints

- One active embedding model
- Model identity recorded per embedding
- No automatic re-embedding

---

## Future Power-User Knobs

- Per-KB embedding models
- Re-embedding workflows
- Multi-vector representations