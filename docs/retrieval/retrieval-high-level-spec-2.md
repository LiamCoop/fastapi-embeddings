# Hybrid Retrieval Control Spec

## 1. Motivation

Different queries benefit from different retrieval strategies:

| Query type | Best strategy |
|---|---|
| Identifiers, error strings, code symbols | Lexical-heavy |
| Conceptual / explanatory questions | Semantic-heavy |
| Mixed natural language with domain terms | Hybrid |
| Short ambiguous queries | Lexical anchoring + semantic expansion |

Providing callers with control over retrieval weighting improves:

- Relevance precision
- Determinism
- Debugging transparency
- Compatibility with diverse LLM agents
- Support for code/log/document-heavy knowledge bases

However, raw weight tuning is too low-level for most users.

**Design goal:** expose meaningful control without cognitive overload.

---

## 2. API Surface

### Query Parameters

```ts
retrieval_profile?: "auto" | "exact" | "balanced" | "semantic"
semantic_weight?: number   // optional override, 0–1
```

Rules:

- `retrieval_profile` defaults to `"auto"`
- `semantic_weight` overrides the profile if present
- If both omitted → index default applies

---

## 3. Retrieval Profiles

### `auto` (default)

System chooses weighting using heuristics.

Target behavior:
- Correct most cases without user involvement
- Provide observability into decisions
- Adapt to code-heavy vs prose-heavy corpora

### `exact`

Lexical-dominant retrieval.

Intended for:
- Code
- Logs
- File paths
- Config keys
- Stack traces
- Identifiers

Typical internal config:
- Lexical weight ≈ 0.75–0.9
- Phrase boosting enabled
- Minimal semantic expansion
- Aggressive metadata filtering

### `balanced`

True hybrid retrieval.

Intended for:
- Documentation
- Knowledge bases
- Mixed queries
- Onboarding material

Typical internal config:
- Lexical weight ≈ 0.4–0.6
- Semantic expansion allowed
- Reranker recommended

### `semantic`

Dense retrieval dominant.

Intended for:
- Exploratory questions
- Summarization queries
- Conceptual reasoning
- Research content

Typical internal config:
- Lexical weight ≈ 0.1–0.3
- Query expansion enabled
- Higher candidate pool for reranking

---

## 4. Numeric Weight Override

Advanced callers may specify:

```
semantic_weight ∈ [0, 1]
```

Interpretation:

```
final_score =
  semantic_weight * dense_score +
  (1 - semantic_weight) * lexical_score
```

Notes:
- Scores must be normalized prior to fusion
- Allow per-index calibration
- Expose effective weight in debug metadata

---

## 5. Auto Profile Heuristics

The `auto` classifier performs lightweight query analysis.

**Signals increasing lexical weight:**
- Quoted phrases
- Symbols (`/`, `.`, `_`, `::`, `->`)
- File paths
- `camelCase` / `snake_case` tokens
- Error-like patterns
- Short queries
- Numbers / version strings

**Signals increasing semantic weight:**
- Question form (how / why / when / what)
- Long natural language queries
- Abstract nouns
- Conversational phrasing
- Absence of rare tokens

**Mixed signals:** fallback to `balanced`.

---

## 6. Retrieval Pipeline Integration

Hybrid retrieval pipeline:

```
query
 → query analysis (auto mode only)
 → lexical retrieval (BM25 / inverted index)
 → dense retrieval (vector search)
 → score normalization
 → weighted fusion
 → optional rerank
 → context assembly
```

Weighting affects:
- Candidate depth
- Fusion scoring
- Reranker invocation threshold

---

## 7. Observability & Debugging

Optional debug response:

```json
{
  "retrieval_profile_effective": "balanced",
  "semantic_weight_effective": 0.5,
  "auto_signals_detected": [...],
  "lexical_candidates": 20,
  "semantic_candidates": 20,
  "reranker_applied": true
}
```

This enables:
- Developer trust
- Offline evaluation
- Default tuning
- Failure diagnosis

---

## 8. Index-Level Defaults

Indexes may define:

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

## 9. Failure Modes & Safeguards

### Over-Semantic Drift

Mitigations:
- Lexical anchoring
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

## 10. Agent / MCP Integration

Agents benefit from:
- Deterministic retrieval modes
- Retry strategies across profiles
- Fast lexical probing before semantic search
- Debug metadata for reasoning traces

Common agent pattern:

```
agent → query(exact)
agent → query(balanced)
agent → synthesize
```

This makes retrieval controllability highly agent-friendly.

---

## Why This Matters

Hybrid control is a subtle but high-impact capability that improves:

- Perceived intelligence of retrieval
- Grounding reliability
- Support for code-heavy knowledge bases
- Debugging transparency
- Agent interoperability

This is especially differentiating for:

- Developer RAG
- Logs and observability knowledge bases
- Multi-modal corpora
- Enterprise documentation platforms
