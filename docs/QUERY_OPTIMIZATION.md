# Query Optimization for Multi-Tenant Vector Search

## Overview

This document describes the vector index optimization strategy for multi-tenant (multi-knowledge-base) retrieval in Ragtime. It addresses the performance characteristics of pgvector indexes when filtering by `kb_id` and provides concrete recommendations for scaling to many knowledge bases.

---

## Current State

### Knowledge Base Isolation ✅

**Data isolation is correct and enforced at multiple levels:**

1. **Query-level filtering:**
   - Both `SearchSemantic` and `SearchLexical` queries include `WHERE c.kb_id = $kb_id`
   - No possibility of cross-KB data leakage

2. **Schema-level constraints:**
   ```sql
   -- chunks table
   kb_id uuid NOT NULL REFERENCES knowledge_bases(id) ON DELETE CASCADE

   -- embeddings table
   kb_id uuid NOT NULL REFERENCES knowledge_bases(id) ON DELETE CASCADE

   -- Deduplication scoped to KB
   UNIQUE (kb_id, content_hash, embedding_model_id, embedding_model_version)
   ```

3. **Indexes on kb_id:**
   ```sql
   CREATE INDEX chunks_kb_id_idx ON chunks (kb_id);
   CREATE INDEX embeddings_kb_id_idx ON embeddings (kb_id);
   ```

**Correctness guarantee:** Querying `kb_id=2` will **never** return results from `kb_id=1` or `kb_id=3`.

---

## Performance Characteristics

### Current Vector Index (IVFFlat)

```sql
-- From migration 001_initial_schema.up.sql
CREATE INDEX embeddings_vector_idx
    ON embeddings
    USING ivfflat (embedding_vector vector_cosine_ops);
```

**How it works:**
1. IVFFlat index clusters vectors into partitions (inverted file lists)
2. Query searches nearest partitions, then scans vectors within those partitions
3. PostgreSQL applies `WHERE kb_id = $1` filter **after** vector similarity computation
4. Results are correctly filtered, but computation happens across all KBs

### Performance Impact by Scale

| Scenario | IVFFlat Behavior | Performance Impact |
|----------|------------------|-------------------|
| **1-3 KBs, <100K chunks total** | Scans entire index, filters afterward | ✅ Negligible - fast enough |
| **10+ KBs, 500K+ chunks** | Scans all KBs' vectors for each query | ⚠️ Noticeable - wasted computation |
| **100+ KBs, multi-million chunks** | Scans vectors from 99+ irrelevant KBs | ❌ Significant - needs optimization |

**Key insight:** IVFFlat was designed for global similarity search, not filtered multi-tenant scenarios.

---

## Optimization Strategies

### Option 1: HNSW Index (Recommended for MVP+)

**Upgrade to HNSW (Hierarchical Navigable Small World):**

```sql
-- Migration: upgrade_to_hnsw.up.sql
DROP INDEX IF EXISTS embeddings_vector_idx;

CREATE INDEX embeddings_vector_hnsw_idx
    ON embeddings
    USING hnsw (embedding_vector vector_cosine_ops)
    WITH (m = 16, ef_construction = 64);
```

**Advantages:**
- ✅ Better support for filtering with `WHERE` clauses
- ✅ No clustering/training required (IVFFlat needs `lists` tuning)
- ✅ Generally faster and more accurate than IVFFlat
- ✅ Graceful degradation under filtering
- ✅ Drop-in replacement (same query syntax)

**Trade-offs:**
- Slightly larger index size (~10-15% more disk space)
- Longer index build time (not relevant for incremental inserts)

**When to apply:** When you have 5+ active knowledge bases or 100K+ total chunks.

---

### Option 2: Table Partitioning (Best for Large-Scale Multi-Tenancy)

**Partition the embeddings table by `kb_id`:**

```sql
-- Migration: partition_embeddings_by_kb.up.sql

-- 1. Rename existing table
ALTER TABLE embeddings RENAME TO embeddings_old;

-- 2. Create partitioned table
CREATE TABLE embeddings (
    id uuid NOT NULL,
    kb_id uuid NOT NULL,
    content_hash text NOT NULL,
    embedding_model_id text NOT NULL,
    embedding_vector vector(1536) NOT NULL,
    embedding_model_version text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (id, kb_id),
    CONSTRAINT embeddings_kb_fk FOREIGN KEY (kb_id)
        REFERENCES knowledge_bases(id) ON DELETE CASCADE
) PARTITION BY HASH (kb_id);

-- 3. Create partitions (adjust number based on expected KBs)
CREATE TABLE embeddings_p0 PARTITION OF embeddings FOR VALUES WITH (MODULUS 16, REMAINDER 0);
CREATE TABLE embeddings_p1 PARTITION OF embeddings FOR VALUES WITH (MODULUS 16, REMAINDER 1);
-- ... repeat for p2 through p15

-- 4. Create vector index on each partition
CREATE INDEX embeddings_p0_vector_idx ON embeddings_p0
    USING hnsw (embedding_vector vector_cosine_ops);
CREATE INDEX embeddings_p1_vector_idx ON embeddings_p1
    USING hnsw (embedding_vector vector_cosine_ops);
-- ... repeat for all partitions

-- 5. Create unique constraint (must include partition key)
CREATE UNIQUE INDEX embeddings_dedupe_unique
    ON embeddings (kb_id, content_hash, embedding_model_id, embedding_model_version);

-- 6. Migrate data
INSERT INTO embeddings SELECT * FROM embeddings_old;

-- 7. Update chunks foreign key
ALTER TABLE chunks DROP CONSTRAINT IF EXISTS chunks_embedding_id_fkey;
-- Note: FK to partitioned table requires including partition key in chunks

-- 8. Drop old table
DROP TABLE embeddings_old;
```

**Advantages:**
- ✅ **Complete physical isolation** - each KB's vectors in separate partition
- ✅ Queries only touch relevant partition(s) - massive performance win
- ✅ Easier to manage and vacuum individual partitions
- ✅ Can set per-partition index parameters
- ✅ Supports table-level operations on specific KBs (backup, restore, drop)

**Trade-offs:**
- ❌ More complex schema (16+ tables instead of 1)
- ❌ Requires including `kb_id` in primary key
- ❌ Foreign key constraints become more complex
- ❌ Migration requires rewriting entire table

**When to apply:** When you have 50+ knowledge bases or need strict resource isolation per KB.

---

### Option 3: Separate Tables Per KB (Not Recommended)

**Create a new embeddings table for each KB dynamically:**

```sql
-- Example: kb_12345_embeddings, kb_67890_embeddings, etc.
```

**Why we don't recommend this:**
- ❌ Schema becomes unpredictable and hard to migrate
- ❌ Sqlc code generation breaks (can't generate queries for dynamic tables)
- ❌ No unified queries across KBs (even for admin/monitoring)
- ❌ Operational complexity (backups, monitoring, connection pooling)

**When to consider:** Only if you need complete database-level isolation for security/compliance (e.g., separate Postgres instances per customer).

---

## Recommended Migration Path

### Phase 1: MVP (Current State)
- **Keep IVFFlat index as-is**
- Monitor query performance as KBs grow
- Set up observability for p95 retrieval latency per KB

**Trigger for Phase 2:** p95 latency >500ms OR >10 active KBs

---

### Phase 2: Upgrade to HNSW
- **Apply HNSW migration** (see Option 1 above)
- Rebuild index (run during low-traffic window)
- Validate performance improvement

**Expected improvement:** 2-5x faster queries in multi-KB scenarios

**Trigger for Phase 3:** >50 KBs OR uneven KB sizes causing "noisy neighbor" issues

---

### Phase 3: Table Partitioning
- **Apply partitioning migration** (see Option 2 above)
- Requires downtime or blue-green migration
- Update application code if needed (should be transparent)

**Expected improvement:** 10-50x faster queries, linear scaling with KBs

---

## Index Tuning Parameters

### IVFFlat (Current)
```sql
-- Default: lists = 100 (not set explicitly)
-- Optimal: lists = rows / 1000 (e.g., 100K rows → lists=100)

-- To rebuild with custom lists:
DROP INDEX embeddings_vector_idx;
CREATE INDEX embeddings_vector_idx
    ON embeddings
    USING ivfflat (embedding_vector vector_cosine_ops)
    WITH (lists = 100);
```

**Rule of thumb:** Set `lists` to roughly `sqrt(total_rows)` but not less than 10.

---

### HNSW (Recommended)
```sql
CREATE INDEX embeddings_vector_hnsw_idx
    ON embeddings
    USING hnsw (embedding_vector vector_cosine_ops)
    WITH (
        m = 16,              -- Number of connections per layer (default: 16)
        ef_construction = 64 -- Size of dynamic candidate list during build (default: 64)
    );
```

**Tuning:**
- **Higher `m`** (24, 32): Better recall, larger index, slower inserts
- **Higher `ef_construction`** (128, 200): Better index quality, slower build
- **For production:** Start with defaults, increase `m` if recall is low

**Query-time parameter:**
```sql
-- Set per-query or per-session
SET hnsw.ef_search = 100;  -- Default: 40, higher = better recall but slower
```

---

## Monitoring and Observability

### Key Metrics to Track

1. **Retrieval latency by KB:**
   ```sql
   SELECT kb_id,
          percentile_cont(0.95) WITHIN GROUP (ORDER BY latency_ms) as p95_latency,
          COUNT(*) as request_count
   FROM retrieval_requests
   WHERE created_at > NOW() - INTERVAL '1 hour'
   GROUP BY kb_id;
   ```

2. **Index usage:**
   ```sql
   SELECT schemaname, tablename, indexname, idx_scan, idx_tup_read, idx_tup_fetch
   FROM pg_stat_user_indexes
   WHERE tablename = 'embeddings';
   ```

3. **Query plan analysis:**
   ```sql
   EXPLAIN (ANALYZE, BUFFERS)
   SELECT ... FROM chunks c JOIN embeddings e ...
   WHERE c.kb_id = 'xxx' AND ...;
   ```

   **Look for:**
   - "Index Scan using embeddings_vector_idx" (good)
   - "Bitmap Index Scan" (acceptable)
   - "Seq Scan" (bad - means index isn't being used)

---

## Testing Strategy

### Before Migration
1. Capture baseline performance:
   - Run 100 queries across different KBs
   - Record p50, p95, p99 latencies
   - Note recall@10 on evaluation set

### After Migration
1. Re-run same queries, compare latencies
2. Validate recall hasn't degraded (HNSW should be same or better)
3. Check index size: `SELECT pg_size_pretty(pg_relation_size('embeddings_vector_hnsw_idx'));`

### Load Testing
- Use `pgbench` or similar to simulate concurrent queries
- Test with realistic KB sizes (small, medium, large)
- Verify no regression under concurrent load

---

## Decision Matrix

| Scenario | Recommendation |
|----------|----------------|
| **<5 KBs, <50K chunks** | Keep IVFFlat, no action needed |
| **5-20 KBs, 50K-500K chunks** | Upgrade to HNSW (Phase 2) |
| **20-50 KBs, 500K-2M chunks** | HNSW + monitor for noisy neighbors |
| **>50 KBs or uneven KB sizes** | Table partitioning (Phase 3) |
| **Strict resource isolation required** | Table partitioning from day 1 |

---

## References

- [pgvector Documentation](https://github.com/pgvector/pgvector)
- [HNSW Paper (Malkov & Yashunin, 2018)](https://arxiv.org/abs/1603.09320)
- [PostgreSQL Partitioning](https://www.postgresql.org/docs/current/ddl-partitioning.html)
- Ragtime internal: `docs/retrieval-prd.md`, `docs/PLAN.md`

---

## Summary

**Current state:** Knowledge base isolation is **correct** but not optimized for multi-tenant scale.

**Recommendation:**
1. Start with IVFFlat (current state is fine for MVP)
2. Upgrade to HNSW when you hit 5-10 KBs or notice latency degradation
3. Consider partitioning only if you scale to 50+ KBs or need strict resource isolation

**Next steps:**
- Add p95 latency monitoring per KB to `retrieval_requests` table
- Set alert threshold (e.g., p95 > 500ms)
- Plan HNSW migration when threshold is crossed
