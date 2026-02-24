# Batch Uploads with Durable Ingestion Jobs

## Context

Currently, directory uploads are sequential one-at-a-time POSTs through the Next.js API route, which proxies the file body to S3 server-side. This means:
- The browser tab must stay open for the entire duration
- Files upload serially (slow for large directories)
- No progress tracking after page navigation
- No server-side record of the batch as a unit

The goal is to let users upload a directory, then leave the page while the system handles uploading and processing. The approach: create a server-owned batch job, get presigned URLs for direct client→S3 uploads, finalize when uploads complete, then process asynchronously with pollable progress.

---

## Implementation Plan

### Phase 1: Database Migration

**New file**: `backend/migrations/012_add_ingestion_batches.up.sql`

```sql
CREATE TABLE ingestion_batches (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    kb_id           uuid NOT NULL REFERENCES knowledge_bases(id) ON DELETE CASCADE,
    status          text NOT NULL DEFAULT 'PENDING',
    total_files     int  NOT NULL DEFAULT 0,
    processed_files int  NOT NULL DEFAULT 0,
    failed_files    int  NOT NULL DEFAULT 0,
    error_message   text,
    created_at      timestamptz NOT NULL DEFAULT now(),
    finalized_at    timestamptz,
    completed_at    timestamptz,
    CONSTRAINT ingestion_batches_status_check CHECK (
        status IN ('PENDING','FINALIZED','PROCESSING','COMPLETED','FAILED')
    )
);

CREATE TABLE ingestion_batch_files (
    id                uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    batch_id          uuid NOT NULL REFERENCES ingestion_batches(id) ON DELETE CASCADE,
    kb_id             uuid NOT NULL,
    file_index        int  NOT NULL,
    path              text NOT NULL,
    file_name         text NOT NULL,
    content_type      text NOT NULL DEFAULT '',
    size_bytes        bigint,
    object_key        text NOT NULL,
    presigned_url     text NOT NULL,
    presigned_headers jsonb NOT NULL DEFAULT '{}'::jsonb,
    raw_content_uri   text NOT NULL,
    status            text NOT NULL DEFAULT 'PENDING',
    document_id       uuid,
    document_version_id uuid,
    ingestion_job_id  uuid,
    processing_status text,
    error_message     text,
    created_at        timestamptz NOT NULL DEFAULT now(),
    processed_at      timestamptz,
    CONSTRAINT batch_files_status_check CHECK (
        status IN ('PENDING','PROCESSING','SUCCESS','FAILED','SKIPPED')
    ),
    CONSTRAINT batch_files_index_unique UNIQUE (batch_id, file_index)
);

CREATE INDEX idx_ingestion_batches_kb ON ingestion_batches(kb_id);
CREATE INDEX idx_batch_files_batch ON ingestion_batch_files(batch_id);
```

**Down migration**: `012_add_ingestion_batches.down.sql` — drop both tables.

---

### Phase 2: Backend — New `batch` Package

Create `backend/internal/batch/` following the existing repo pattern.

#### 2a. Types & Repository Interface

**New file**: `backend/internal/batch/contract.go`

- `BatchRecord` and `BatchFileRecord` structs matching the DB tables
- `CreateBatchRequest` with `KnowledgeBaseID` + `[]FileManifestEntry{Path, FileName, ContentType, SizeBytes}`
- `CreateBatchResponse` with `BatchID` + `[]FileUploadInfo{FileIndex, Path, UploadURL, Headers, RawContentURI}`
- `BatchStatusResponse` with counts, status, and `[]FileStatus{FileIndex, Path, Status, ProcessingStatus, DocumentID, ErrorMessage}`
- `Repository` interface:
  - `InsertBatch(ctx, BatchRecord) error`
  - `GetBatch(ctx, batchID) (*BatchRecord, error)`
  - `UpdateBatchStatus(ctx, batchID, status, errorMessage) error`
  - `SetBatchFinalized(ctx, batchID, time) error`
  - `SetBatchCompleted(ctx, batchID, time) error`
  - `IncrementProcessed(ctx, batchID) error`
  - `IncrementFailed(ctx, batchID) error`
  - `InsertBatchFiles(ctx, []BatchFileRecord) error`
  - `ListBatchFiles(ctx, batchID) ([]BatchFileRecord, error)`
  - `UpdateBatchFileResult(ctx, fileID, status, docID, versionID, jobID, processingStatus, errMsg) error`
  - `ListActiveBatches(ctx, kbID) ([]BatchRecord, error)` — status in (PENDING, FINALIZED, PROCESSING)

#### 2b. Postgres Repository

**New file**: `backend/internal/batch/repository.go`

Standard `*sql.DB`-based implementation. Each method is a single SQL query. `IncrementProcessed`/`IncrementFailed` use `SET processed_files = processed_files + 1` for atomicity.

#### 2c. Service

**New file**: `backend/internal/batch/service.go`

Dependencies: `Repository`, `objectstore.Client`, `*ingestion.Service`

**`CreateBatch(ctx, req)`**:
1. Validate: kbID non-empty, files non-empty, each file has a path
2. Generate batch ID
3. For each file, call `s.store.PresignPut(ctx, buildBatchKey(kbID, batchID, i, fileName), contentType)` — reuses existing `objectstore.Client.PresignPut`
4. Insert batch + files in DB
5. Return presigned URLs

Key: `buildBatchKey` produces `kb/{kbID}/batches/{batchID}/{fileIndex}_{fileName}`

**`FinalizeBatch(ctx, kbID, batchID)`**:
1. Get batch, verify kbID match and status == PENDING
2. Set status = FINALIZED, finalized_at = now
3. Spawn `go s.processBatch(batchID)` (background goroutine with `context.Background()`)
4. Return immediately

**`GetBatchStatus(ctx, kbID, batchID)`**:
1. Get batch + files, verify kbID match
2. Map to `BatchStatusResponse`

**`processBatch(batchID)`** (private, runs in goroutine):
1. Set batch status = PROCESSING
2. List batch files
3. For each file:
   - Set file status = PROCESSING
   - Call `s.ingestion.IngestDocuments(ctx, IngestDocumentsRequest{Documents: []IngestDocumentRequest{{KnowledgeBaseID, Path, RawContentURI, DocumentType (inferred from path)}}})` — reuses the entire existing pipeline (upload → chunk → embed → activate)
   - On success: update file with docID, versionID, jobID, status=SUCCESS; increment batch processed
   - On failure: update file with error, status=FAILED; increment batch failed
4. Set batch status = COMPLETED (or FAILED if all failed), completed_at = now
5. Log completion

#### 2d. HTTP Handler

**New file**: `backend/internal/batch/http/handler.go`

Three endpoints:

| Method | Path | Handler | Response |
|--------|------|---------|----------|
| POST | `/v1/kb/{kbID}/batches` | `CreateBatch` | 201 with batch_id + presigned URLs |
| GET | `/v1/kb/{kbID}/batches/{batchID}` | `GetBatchStatus` | 200 with status + file statuses |
| POST | `/v1/kb/{kbID}/batches/{batchID}/finalize` | `FinalizeBatch` | 202 Accepted |

#### 2e. Wire into main.go

**Modify**: `backend/cmd/server/main.go`

- Import batch packages
- Add `ingestion *ingestion.Service` and `batch *batch.Service` to `appServices`
- In `newServices()`: create `ingestion.NewServiceWithPostgres(db, store, embedService)` and `batch.NewService(batchRepo, store, ingestionSvc)`
- Register 3 new routes on the chi router

---

### Phase 3: Frontend API Routes

Three new Next.js API routes that proxy to the Go backend:

**New file**: `frontend/app/api/org/[slug]/knowledge/[kbId]/batches/route.ts`
- POST: validate org+kb ownership, proxy to `POST /v1/kb/{kbId}/batches`

**New file**: `frontend/app/api/org/[slug]/knowledge/[kbId]/batches/[batchId]/route.ts`
- GET: validate org+kb, proxy to `GET /v1/kb/{kbId}/batches/{batchId}`

**New file**: `frontend/app/api/org/[slug]/knowledge/[kbId]/batches/[batchId]/finalize/route.ts`
- POST: validate org+kb, proxy to `POST /v1/kb/{kbId}/batches/{batchId}/finalize`

Follow the same patterns as existing `documents/route.ts` for auth/validation/proxy.

---

### Phase 4: Frontend Types & Helpers

**Modify**: `frontend/app/lib/org-knowledge.ts`

Add types: `BatchFile`, `CreateBatchRequest`, `CreateBatchResponse`, `CreateBatchFileResponse`, `BatchStatusResponse`, `BatchFileStatus`

Add path helpers: `knowledgeBatchesApiPath(slug, kbId)`, `knowledgeBatchStatusApiPath(slug, kbId, batchId)`, `knowledgeBatchFinalizeApiPath(slug, kbId, batchId)`

---

### Phase 5: Frontend — Rewrite UploadZone

**Modify**: `frontend/app/(app)/org/[slug]/knowledge/[kbId]/ingestion/_components/UploadZone.tsx`

New states:
- `batchId: string | null` — active batch
- `phase: 'idle' | 'uploading' | 'finalizing' | 'processing' | 'done' | 'error'`
- `uploadProgress: { completed: number; total: number; failed: string[] }`
- `processingProgress: { processed: number; failed: number; total: number }`

**Single file flow** — unchanged (keep the existing FormData POST for single files).

**Directory flow** — new batch path:
1. User selects directory → filter to markdown files
2. POST to `/api/.../batches` with manifest → receive `batch_id` + presigned URLs
3. Save `{ batchId, kbId, slug }` to `localStorage` key `ragtime:active-batch:{kbId}`
4. Upload files in parallel (concurrency limit of 4) using `fetch(url, { method: 'PUT', headers, body: file })`. Track progress via completed count.
5. After all uploads finish → POST to `/api/.../batches/{batchId}/finalize`
6. Switch to polling phase: `setInterval` every 3s calling `GET /api/.../batches/{batchId}` → update `processingProgress`
7. On COMPLETED/FAILED: clear localStorage, `router.refresh()`, show summary

**On mount**: check localStorage for active batch → if found, resume polling from step 6.

**Progress UI**:
- Phase "uploading": `Uploading {completed}/{total} files...` with a simple progress bar
- Phase "processing": `Processing {processed}/{total} files...` — user can navigate away
- Phase "done": success summary with file count
- Phase "error": error details with per-file failures expandable

---

### Phase 6: Ingestion Page — Active Batch Indicator

**Modify**: `frontend/app/(app)/org/[slug]/knowledge/[kbId]/ingestion/page.tsx`

No server-side changes needed. The UploadZone component handles everything client-side via localStorage + polling. When a batch completes, `router.refresh()` reloads the documents list.

---

## Files to Create
- `backend/migrations/012_add_ingestion_batches.up.sql`
- `backend/migrations/012_add_ingestion_batches.down.sql`
- `backend/internal/batch/contract.go`
- `backend/internal/batch/repository.go`
- `backend/internal/batch/service.go`
- `backend/internal/batch/http/handler.go`
- `frontend/app/api/org/[slug]/knowledge/[kbId]/batches/route.ts`
- `frontend/app/api/org/[slug]/knowledge/[kbId]/batches/[batchId]/route.ts`
- `frontend/app/api/org/[slug]/knowledge/[kbId]/batches/[batchId]/finalize/route.ts`

## Files to Modify
- `backend/cmd/server/main.go` — wire batch service + routes
- `frontend/app/lib/org-knowledge.ts` — add batch types and path helpers
- `frontend/app/(app)/org/[slug]/knowledge/[kbId]/ingestion/_components/UploadZone.tsx` — batch upload flow

## Key Reuse Points
- `objectstore.Client.PresignPut()` (`backend/internal/objectstore/objectstore.go:12`) — generates presigned URLs
- `ingestion.Service.IngestDocuments()` (`backend/internal/ingestion/service.go:74`) — full pipeline per doc
- `document.DetectDocumentType()` (`backend/internal/document/service.go:216`) — infer doc type from path
- `directoryRelativePath()` in UploadZone — existing path extraction logic, keep as-is

## Verification
1. **Backend unit tests**: Test batch service `CreateBatch`, `FinalizeBatch`, `processBatch` with mocked repo/store/ingestion
2. **Manual E2E**: Upload a directory of 5+ markdown files, verify presigned URLs work, finalize, watch processing progress via GET endpoint, confirm documents appear in list after completion
3. **Leave-the-page test**: Start a batch upload, navigate away mid-processing, return to page — verify localStorage restores polling and progress shows correctly
4. **Error case**: Include a malformed markdown file in the batch, verify it fails individually without blocking other files, batch completes with `failed_files=1`
