# Chunking Strategies – Architecture Decision Record

## Problem

Documents must be broken into retrievable units that balance:
- Semantic coherence
- Retrieval recall
- Stability across document updates

---

## Options Considered

### Option A: Fixed Token Windows
**Description**
- Split content into fixed-size token windows with overlap

**Pros**
- Simple
- Fast
- Predictable

**Cons**
- Breaks semantic structure
- High churn on small edits
- Poor citation fidelity

---

### Option B: Structural Chunking
**Description**
- Chunk based on document structure (headers, sections, lists, code blocks)

**Pros**
- Preserves meaning
- Stable boundaries
- Better citations
- Lower re-embedding churn

**Cons**
- Requires parsing
- More complex implementation

---

### Option C: Semantic Chunking
**Description**
- Use models to dynamically segment text by meaning

**Pros**
- High-quality chunks
- Adaptive to content

**Cons**
- Expensive
- Non-deterministic
- Hard to debug

---

## MVP Decision

**Selected: Option B – Structural Chunking**

### Rationale
- Best balance of quality and stability
- Deterministic behavior
- Enables future optimizations
- Aligns with versioned document model

---

## MVP Constraints

- Chunk size bounds enforced
- No cross-section merging
- One chunk maps to one document version

---

## Future Power-User Knobs

- Token window overrides
- Hybrid structural + semantic chunking
- Chunk size tuning per document type