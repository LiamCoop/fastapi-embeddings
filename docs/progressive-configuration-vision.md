# Progressive Configuration: Easy Mode to Power User Mode

## Product Philosophy

**Core Tension**: RAG systems face a fundamental tradeoff:
- **Too Simple**: Users hit walls when defaults don't work for their content
- **Too Complex**: Users are overwhelmed and never get started

**Solution**: Progressive disclosure — start with magic, graduate to mastery.

### The Spectrum

```
┌─────────────────────────────────────────────────────────────┐
│  Easy Mode                                      Power Mode   │
│  "Just make it work"                   "I know what I need"  │
│                                                               │
│  ░░░░░░░░░░░░░░░░░▓▓▓▓▓▓▓▓▓▓▓▓▓▓███████████████             │
│  Zero Config       Some Config        Full Control           │
│                                                               │
│  80% of users      15% of users       5% of users            │
│  Start here        Graduate here      End here               │
└─────────────────────────────────────────────────────────────┘
```

### Design Mantras

1. **"Show me the magic first"** - New users should see value in 60 seconds
2. **"Give me the controls when I'm ready"** - Reveal complexity on demand
3. **"Let me experiment safely"** - Configuration changes shouldn't be scary
4. **"Teach me what's possible"** - System should educate users about options

---

## The Four Tiers

### Tier 1: Easy Mode (MVP) ✅

**Target User**: "I just want RAG to work"
- Data scientists getting started
- Product teams evaluating RAG
- Anyone who doesn't know what "chunking strategy" means

**Experience**:
```bash
# Upload documents
curl -X POST https://api.ragtime.dev/v1/kb/my-kb/documents \
  -F "file=@docs.zip"

# Query immediately
curl https://api.ragtime.dev/v1/kb/my-kb/retrieve?query="how do I authenticate"
```

**Under the Hood**:
- Single global chunking strategy: `structural` (header-based)
- Single global embedding model: `text-embedding-3-small` (1536-dim)
- Optimized hybrid search (semantic + lexical)
- Automatic citation generation

**Why This Works**:
- Zero decisions required
- Fast time-to-value
- Defaults are tuned for general-purpose retrieval
- Handles 80% of use cases well enough

**Success Metric**: User sees relevant results in first query

---

### Tier 2: KB-Level Configuration

**Target User**: "Different knowledge bases have different needs"
- Technical writer with code docs + prose docs in separate KBs
- Company with internal KB (needs high precision) + public KB (needs high recall)
- User who's hit quality ceiling with defaults

**Experience**:
```bash
# Create KB with configuration
curl -X POST https://api.ragtime.dev/v1/kb \
  -d '{
    "name": "API Reference Docs",
    "configuration": {
      "chunking_strategy": "structural",
      "embedding_model": "text-embedding-3-large",
      "hybrid_weight": 0.7
    }
  }'

# All documents in this KB inherit these settings
```

**Configuration Options**:

**Chunking Strategies**:
- `structural`: Header/section-based (good for docs, code)
- `fixed_window`: Token-based with overlap (good for prose, chat logs)
- `semantic`: Embedding-similarity-based boundaries (good for unstructured content)

**Embedding Models**:
- `text-embedding-3-small` (1536-dim): Fast, cheap, good enough
- `text-embedding-3-large` (3072-dim): Better quality, 2x cost
- `custom/your-model`: Bring your own embedding service

**Hybrid Search Weight** (0.0 - 1.0):
- `1.0`: Pure semantic (best for conceptual queries)
- `0.5`: Balanced (default)
- `0.0`: Pure lexical (best for exact phrase matching)

**Why This Matters**:
- Different content types need different strategies
- Users learn what works through experimentation
- KB isolation makes A/B testing safe
- Costs scale with needs (small model for dev, large for prod)

**UX Question**: Should we auto-suggest configurations based on content analysis?
- "We detected mostly Markdown files with code blocks → recommend `structural` chunking"
- "We detected long-form narrative → recommend `semantic` chunking"

---

### Tier 3: Document-Level Overrides

**Target User**: "Most documents work with defaults, but some need special handling"
- Technical docs with a few deeply technical API references
- Knowledge base with mix of structured (code) and unstructured (meeting notes) content
- User who wants to optimize specific high-traffic documents

**Experience**:
```bash
# Upload document with override
curl -X POST https://api.ragtime.dev/v1/kb/my-kb/documents \
  -F "file=@complex-api-spec.md" \
  -F 'configuration={
    "chunking_strategy": "semantic",
    "embedding_model": "text-embedding-3-large",
    "max_chunk_size": 2048
  }'

# Document uses custom config, rest of KB uses defaults
```

**Real-World Scenarios**:

**Scenario 1: Code vs Prose**
- Default: `structural` chunking for markdown docs
- Override: `fixed_window` for Jupyter notebooks (no clear headers)

**Scenario 2: High-Value Documents**
- Default: `text-embedding-3-small` for general docs
- Override: `text-embedding-3-large` for critical reference docs (SLAs, security policies)

**Scenario 3: Chunk Size Tuning**
- Default: 512 tokens (good for most content)
- Override: 128 tokens for FAQ documents (tighter answers)
- Override: 2048 tokens for research papers (preserve context)

**Why This Matters**:
- 80/20 rule: Most docs use defaults, few need tuning
- Users can experiment without re-indexing entire KB
- Costs are targeted (only critical docs get expensive embeddings)

**UX Considerations**:
- Should UI show "This document uses custom configuration" badge?
- Should we show comparative stats: "This override improved recall by 15%"?
- How do users discover which documents might benefit from overrides?

---

### Tier 4: Section-Level Overrides (Most Complex)

**Target User**: "I need surgical precision for specific content"
- Technical writer: "This glossary section needs tighter chunking"
- Legal team: "These contract clauses must be chunked at paragraph level"
- Research team: "Mathematical proofs need semantic chunking, prose needs structural"

**Experience (Visual UI)**:

```
┌─────────────────────────────────────────────────┐
│  Document: architecture.md                      │
├─────────────────────────────────────────────────┤
│                                                  │
│  # System Architecture                          │  ← Default: structural
│                                                  │
│  ## Overview                                    │
│  Our system uses a microservices architecture...│
│                                                  │
│  ## Mathematical Model                          │  ← Highlighted section
│  ┌──────────────────────────────────────────┐  │
│  │ Let f(x) = ∫ e^(-x²) dx                  │  │  Override: semantic
│  │ Where x ∈ ℝ and...                       │  │  Max chunk: 256 tokens
│  │                                          │  │  Embedding: large
│  └──────────────────────────────────────────┘  │
│                                                  │
│  ## Implementation Details                      │  ← Default: structural
│  ...                                            │
│                                                  │
└─────────────────────────────────────────────────┘
```

**Configuration UI**:
1. User highlights section of document
2. Sidebar shows options:
   - Chunking strategy for this section
   - Embedding model for this section
   - Max chunk size
   - Overlap percentage (for fixed_window)
3. System shows preview: "This will create N chunks"
4. User saves, document re-processes only affected section

**Storage Schema**:
```sql
CREATE TABLE document_section_configuration (
    id UUID PRIMARY KEY,
    document_id UUID NOT NULL,
    section_start_offset INTEGER NOT NULL,  -- Character offset in document
    section_end_offset INTEGER NOT NULL,
    chunking_strategy TEXT NOT NULL,
    embedding_model TEXT NOT NULL,
    config_json JSONB,  -- Strategy-specific params
    created_at TIMESTAMPTZ NOT NULL
);

-- Chunks store which section config created them
ALTER TABLE chunks ADD COLUMN section_config_id UUID REFERENCES document_section_configuration(id);
```

**Why This Matters**:
- Heterogeneous documents are real (specs with code + prose + tables)
- Precision at this level can dramatically improve retrieval quality
- Users become experts in their own content

**Why This Is Hard**:
- UX complexity is significant (selection, visualization, preview)
- Re-chunking parts of a document is architecturally complex
- Chunk boundaries can conflict (section override vs structural chunk)
- Versioning becomes complicated (did section config change or content?)

**Open Questions**:
- What happens when section boundaries overlap?
- Can users nest section configurations?
- How do we show "coverage map" (which chunks use which configs)?

---

## The Multi-Dimensional Embedding Challenge

### The Problem

Users want:
```
KB: "engineering-docs"
  ├─ Architecture docs → text-embedding-3-large (3072-dim)  [high precision]
  ├─ API references   → text-embedding-3-small (1536-dim)  [fast, cheap]
  └─ Meeting notes    → custom/local-model (768-dim)       [privacy]
```

But retrieval requires comparing embeddings: **you can't compute cosine similarity across different dimensions.**

### Solution 1: Multiple Search Indexes (Isolation)

**Approach**: Each KB has multiple search indexes, each with homogeneous embeddings.

```sql
CREATE TABLE search_indexes (
    id UUID PRIMARY KEY,
    kb_id UUID NOT NULL,
    name TEXT NOT NULL,  -- "high-precision", "fast", "privacy"
    embedding_model TEXT NOT NULL,
    dimension INTEGER NOT NULL,
    is_default BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL
);

-- Chunks are indexed in one or more search indexes
CREATE TABLE index_embeddings (
    search_index_id UUID NOT NULL,
    chunk_id UUID NOT NULL,
    embedding vector,  -- Dimension matches search_index.dimension
    PRIMARY KEY (search_index_id, chunk_id)
);
```

**Retrieval API**:
```bash
# Query default index (1536-dim)
GET /v1/kb/engineering-docs/retrieve?query=authentication

# Query specific index (3072-dim)
GET /v1/kb/engineering-docs/retrieve?query=authentication&index=high-precision

# Query multiple indexes, return best from each (no cross-index ranking)
GET /v1/kb/engineering-docs/retrieve?query=authentication&indexes=fast,high-precision&merge=interleave
```

**Pros**:
- Clean architecture (no dimension mixing)
- Users explicitly choose precision/cost tradeoff per query
- Can index same chunk in multiple indexes (reuse chunk_id)

**Cons**:
- User must know which index to query
- Can't automatically route query to "best" index
- Storage cost: Each index stores embeddings for all indexed chunks

**Use Cases**:
- Development vs Production (cheap vs expensive)
- Public vs Private (cloud vs local)
- Fast vs Accurate (small vs large model)

---

### Solution 2: Learned Score Normalization (Research Problem)

**Approach**: Compute similarity in each embedding space, normalize scores, blend results.

```python
# Query uses multiple embeddings
query_emb_1536 = embed(query, model="text-embedding-3-small")
query_emb_3072 = embed(query, model="text-embedding-3-large")

results = []
for chunk in chunks:
    if chunk.embedding_model == "text-embedding-3-small":
        score_1536 = cosine_similarity(query_emb_1536, chunk.embedding_1536)
        score_normalized = normalize(score_1536, distribution_1536)
    elif chunk.embedding_model == "text-embedding-3-large":
        score_3072 = cosine_similarity(query_emb_3072, chunk.embedding_3072)
        score_normalized = normalize(score_3072, distribution_3072)

    results.append((chunk, score_normalized))

results.sort(key=lambda x: x[1], reverse=True)
```

**Challenge**: How do you normalize scores across embedding spaces?
- Empirical distribution: Collect statistics on typical similarity scores per model
- Z-score normalization: (score - mean) / stddev
- Learned calibration: Train a ranker to blend heterogeneous scores

**Pros**:
- Users don't need to choose index (system finds best results)
- Can automatically route important queries to expensive models

**Cons**:
- Requires significant evaluation data to calibrate
- Score distributions may be query-dependent
- Computationally expensive (multiple embeddings per query)
- Hard to explain to users: "Why did this result rank higher?"

**Verdict**: Interesting research problem, but too risky for MVP. Defer to Tier 4+.

---

### Solution 3: KB "Lanes" (Pragmatic Middle Ground)

**Approach**: Partition KB into sub-collections with homogeneous configs.

```bash
# Create KB with lanes
curl -X POST https://api.ragtime.dev/v1/kb \
  -d '{
    "name": "engineering-docs",
    "lanes": [
      {"name": "architecture", "embedding_model": "text-embedding-3-large"},
      {"name": "api", "embedding_model": "text-embedding-3-small"},
      {"name": "notes", "embedding_model": "custom/local"}
    ]
  }'

# Upload to specific lane
curl -X POST https://api.ragtime.dev/v1/kb/engineering-docs/lanes/architecture/documents \
  -F "file=@system-design.md"

# Query specific lane
GET /v1/kb/engineering-docs/lanes/architecture/retrieve?query=...

# Query all lanes (round-robin results, no score comparison)
GET /v1/kb/engineering-docs/retrieve?query=...&lanes=*
```

**Pros**:
- Explicit partitioning (no magic score blending)
- Users understand boundaries
- Can query lanes independently or together

**Cons**:
- Users must assign documents to lanes
- Cross-lane queries are limited (no unified ranking)

**Use Cases**:
- Organizational boundaries (teams, projects)
- Content types (docs, code, data)
- Security zones (public, internal, confidential)

---

## User Journey: Graduating Through Tiers

### Act 1: Discovery (Tier 1 - Easy Mode)

**Day 1**: Alice is a product manager evaluating RAG systems.
```bash
# Upload company wiki
ragtime upload --kb company-wiki ./wiki-export.zip

# First query
ragtime query --kb company-wiki "how do we handle refunds"
```

**Result**: Gets decent results in 60 seconds. Impressed. Signs up.

**Friction**: None. Magic works.

---

### Act 2: Optimization (Tier 2 - KB-Level Config)

**Week 2**: Alice notices API docs have worse retrieval than policy docs.

**Investigation**:
- API docs are highly structured (code blocks, function signatures)
- Policy docs are prose

**Action**: Creates separate KBs with different configs.
```bash
ragtime create-kb api-docs --chunking structural --embedding text-embedding-3-small
ragtime create-kb policies --chunking semantic --embedding text-embedding-3-large
```

**Result**: API doc retrieval improves 30%. Alice is hooked.

**Friction**: Minimal. Config options are discoverable. System suggests values based on content.

---

### Act 3: Precision (Tier 3 - Document-Level Overrides)

**Month 2**: Alice finds that one specific document (security-framework.md) has poor retrieval.

**Investigation**:
- Document is very long (10k tokens)
- Has mix of high-level overview + detailed procedures
- Default chunks are too large (lose precision) or too small (lose context)

**Action**: Override just this document.
```bash
ragtime configure-document security-framework.md \
  --chunking semantic \
  --max-chunk-size 1024 \
  --embedding text-embedding-3-large
```

**Result**: Retrieval quality for this doc jumps to 90%. Rest of KB unchanged.

**Friction**: Low. UI shows before/after preview. Can rollback if worse.

---

### Act 4: Mastery (Tier 4 - Section-Level Overrides)

**Month 6**: Alice is a power user. She's tuning a critical compliance document.

**Investigation**:
- Document has glossary (needs tight chunking)
- Has procedures (needs semantic chunking to preserve steps)
- Has legal clauses (must chunk at paragraph boundaries)

**Action**: Uses visual editor to highlight sections and assign configs.

**Result**: Compliance queries now hit correct clauses 95% of the time. Alice is a Ragtime evangelist.

**Friction**: Higher. Requires time investment. But Alice is motivated (compliance is critical).

---

## Implementation Roadmap

### Phase 1: Foundation (MVP) ✅
- Tier 1 only (Easy Mode)
- Single global defaults
- Prove core retrieval quality
- Establish observability baseline

**Success Criteria**:
- 80% of users get satisfactory results with defaults
- Can measure retrieval quality (Recall@10, MRR)
- Error rates < 1%

---

### Phase 2: KB-Level Configuration (Post-MVP)

**Prerequisites**:
1. Evaluation framework working
2. 3+ months of usage data
3. User feedback system operational

**Implementation**:
```
1. Add kb_configuration table
2. Build ConfigurationResolver service
3. Extend ingestion pipeline to respect KB config
4. Update UI: KB creation wizard with config options
5. Add config comparison tool (A/B test configs)
6. Document best practices guide
```

**Success Criteria**:
- 20% of users experiment with non-default configs
- Can prove that config changes improve retrieval quality
- Users understand tradeoffs (cost, latency, quality)

**Timeline**: 3-4 months

---

### Phase 3: Multiple Search Indexes (Tier 2+)

**Prerequisites**:
1. Clear user demand for multi-dimensional embeddings
2. Storage cost model validated
3. Query routing strategy defined

**Implementation**:
```
1. Add search_indexes and index_embeddings tables
2. Refactor embedding storage to support multiple dimensions
3. Update retrieval API to accept index parameter
4. Build index management UI
5. Implement index interleaving (merge strategy)
6. Add cost estimator (storage per index)
```

**Success Criteria**:
- Users can create multiple indexes per KB
- Query routing is intuitive (default index vs explicit selection)
- Cost transparency (users understand storage implications)

**Timeline**: 2-3 months

---

### Phase 4: Document-Level Overrides (Tier 3)

**Prerequisites**:
1. Users are comfortable with KB-level configs
2. Clear patterns emerge for which documents need overrides
3. Cost tracking per document exists

**Implementation**:
```
1. Add document_configuration table
2. Extend ConfigurationResolver to walk hierarchy
3. Update document upload API to accept config overrides
4. Build document config UI (edit after upload)
5. Add "suggest override" feature (based on retrieval analytics)
6. Show config attribution in retrieval results
```

**Success Criteria**:
- 5-10% of documents use overrides
- Overrides correlate with measurable quality improvements
- Users can easily revert overrides

**Timeline**: 2-3 months

---

### Phase 5: Section-Level Overrides (Tier 4)

**Prerequisites**:
1. Strong user demand (research/legal use cases validated)
2. Visual document editor prototyped
3. Partial re-chunking architecture proven

**Implementation**:
```
1. Build visual document editor (highlight + configure)
2. Add document_section_configuration table
3. Implement section-aware chunking engine
4. Add chunk "coverage map" visualization
5. Build section config preview (show chunk boundaries)
6. Implement conflict resolution (overlapping configs)
```

**Success Criteria**:
- Power users can surgically optimize documents
- Section overrides don't break document versioning
- UX is intuitive (5-minute time-to-first-override)

**Timeline**: 4-6 months (high UX complexity)

---

## Business Model Implications

### Pricing Tiers Aligned with Configuration Tiers

**Free Tier**: Easy Mode Only
- 1 KB, 100 documents, 1000 queries/month
- Default configs only
- Perfect for evaluation

**Pro Tier ($49/mo)**: KB-Level Configuration
- 10 KBs, 10k documents, 50k queries/month
- Choose chunking strategies and embedding models per KB
- A/B testing tools
- Priority support

**Enterprise Tier ($499/mo)**: Document + Section Overrides
- Unlimited KBs, documents, queries
- Document-level and section-level overrides
- Multiple search indexes per KB
- Custom embedding models
- White-glove onboarding

### Value Ladder

```
Free → Pro → Enterprise
  ↓       ↓       ↓
Try  → Optimize → Master
```

**Why This Works**:
- Users graduate naturally through value tiers
- Configuration complexity gates pricing (fair proxy for value)
- Power users self-select into higher tiers
- Costs scale with sophistication (advanced features = more compute)

---

## Open Questions & Research Directions

### 1. Auto-Configuration
Can we automatically suggest configurations based on content analysis?
```python
def suggest_config(documents):
    if mostly_code(documents):
        return {"chunking": "structural", "embedding": "small"}
    elif highly_technical(documents):
        return {"chunking": "semantic", "embedding": "large"}
    elif short_form(documents):
        return {"chunking": "fixed_window", "max_chunk": 256}
```

**Tradeoff**: Convenience vs control. Do users trust auto-suggestions?

---

### 2. Query Routing
Can we automatically route queries to the best index/config?
```python
def route_query(query, kb):
    if query_is_precise(query):  # "what is the refund policy"
        return kb.indexes["high-precision"]
    elif query_is_exploratory(query):  # "tell me about refunds"
        return kb.indexes["fast"]
```

**Challenge**: Requires query classification. Adds latency.

---

### 3. Configuration Recommendations
Can we analyze retrieval quality and suggest config changes?
```
🔔 Notification: "Documents in KB 'api-docs' have 20% lower recall than average.
   Try switching to 'semantic' chunking to improve results."
```

**Requirements**:
- Evaluation data per KB
- Baseline quality metrics
- Change attribution (did config change actually help?)

---

### 4. Config Versioning
Should configuration changes create new KB/document versions?
- **Pro**: Full audit trail, can rollback config changes
- **Con**: Version explosion, confusing UX

**Alternative**: Treat config as metadata (not versioned), but log all changes.

---

### 5. Cross-Index Search (Federated)
Can we query multiple indexes and merge results intelligently?
```bash
GET /v1/kb/docs/retrieve?query=auth&indexes=*&merge=score_normalized
```

**Approaches**:
- **Interleave**: Round-robin results from each index (no score comparison)
- **Score normalization**: Calibrate scores, rank globally (hard)
- **Learned ranker**: Train model to blend heterogeneous results (research)

**Verdict**: Defer until Tier 4+, requires significant research.

---

## Conclusion: Why This Matters

### For Users
- **Beginners**: Get value immediately without learning RAG internals
- **Practitioners**: Optimize incrementally as they learn what works
- **Experts**: Have full control to tune for their specific use cases

### For Ragtime
- **Differentiation**: Most RAG services are one-size-fits-all
- **Stickiness**: Users invest time learning the system (switching cost increases)
- **Expansion**: Natural upsell path (free → pro → enterprise)
- **Data flywheel**: Power users generate evaluation data that improves defaults for easy mode

### For the Industry
- **Transparency**: Makes RAG configuration explicit (not magic black box)
- **Education**: Teaches users what's possible (chunking, embeddings, hybrid search)
- **Standards**: Establishes patterns for configurable RAG systems

---

## Next Steps

1. **Validate MVP** (Tier 1):
   - Launch with easy mode only
   - Collect 3 months of usage data
   - Measure: What % of users get satisfactory results with defaults?

2. **User Research** (Pre-Tier 2):
   - Identify users hitting quality ceiling
   - Interview: What knobs do they wish they had?
   - Validate: Would KB-level config solve their problems?

3. **Build Tier 2** (KB-Level Config):
   - Start with 2-3 chunking strategies only
   - Start with 2 embedding models only
   - Prove that config changes improve quality (evaluation framework)

4. **Iterate**: Only proceed to Tier 3/4 if user demand is clear and Tier 2 is stable.

---

## Appendix: Configuration Schema (Full)

```sql
-- Global defaults (implicit, not stored)
-- chunking_strategy: 'structural'
-- embedding_model: 'text-embedding-3-small'

-- KB-level configuration
CREATE TABLE kb_configuration (
    kb_id UUID PRIMARY KEY REFERENCES knowledge_bases(id),
    chunking_strategy TEXT NOT NULL DEFAULT 'structural',
    embedding_model TEXT NOT NULL DEFAULT 'text-embedding-3-small',
    hybrid_weight DECIMAL(3,2) DEFAULT 0.50,
    max_chunk_size INTEGER DEFAULT 512,
    chunk_overlap INTEGER DEFAULT 50,
    config_json JSONB,  -- Strategy-specific params
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Document-level overrides (Tier 3)
CREATE TABLE document_configuration (
    document_id UUID PRIMARY KEY REFERENCES documents(id),
    chunking_strategy_override TEXT,
    embedding_model_override TEXT,
    search_index_override UUID REFERENCES search_indexes(id),
    config_json_override JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Section-level overrides (Tier 4)
CREATE TABLE document_section_configuration (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    document_id UUID NOT NULL REFERENCES documents(id),
    section_start_offset INTEGER NOT NULL,
    section_end_offset INTEGER NOT NULL,
    chunking_strategy TEXT NOT NULL,
    embedding_model TEXT NOT NULL,
    config_json JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Multiple search indexes per KB (multi-dimensional embeddings)
CREATE TABLE search_indexes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    kb_id UUID NOT NULL REFERENCES knowledge_bases(id),
    name TEXT NOT NULL,
    embedding_model TEXT NOT NULL,
    dimension INTEGER NOT NULL,
    is_default BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(kb_id, name)
);

-- Embeddings stored per index (dimension-specific)
CREATE TABLE index_embeddings (
    search_index_id UUID NOT NULL REFERENCES search_indexes(id),
    chunk_id UUID NOT NULL REFERENCES chunks(id),
    embedding vector,  -- Dimension matches search_index.dimension
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (search_index_id, chunk_id)
);

-- Configuration change audit log
CREATE TABLE configuration_changes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    entity_type TEXT NOT NULL,  -- 'kb', 'document', 'section'
    entity_id UUID NOT NULL,
    changed_by TEXT NOT NULL,
    old_config JSONB,
    new_config JSONB,
    reason TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Query feedback (connects to Tier 2+ for config optimization)
CREATE TABLE retrieval_feedback (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    retrieval_request_id UUID NOT NULL REFERENCES retrieval_requests(id),
    kb_id UUID NOT NULL REFERENCES knowledge_bases(id),
    overall_rating INTEGER CHECK (overall_rating BETWEEN 1 AND 5),
    feedback_text TEXT,
    chunk_ratings JSONB,
    missing_information BOOLEAN DEFAULT FALSE,
    incorrect_results BOOLEAN DEFAULT FALSE,
    citation_problems BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

---

**Document Status**: Vision / Not Implemented
**Last Updated**: 2026-02-04
**Owner**: Product & Engineering
**Next Review**: After MVP Launch + 3 months
