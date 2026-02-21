# Plan: Document Upload for Ingestion Page

## Context

The ingestion page (`app/(app)/org/[slug]/knowledge/[kbId]/ingestion/page.tsx`) has a static, non-functional drag-and-drop placeholder. The goal is to make it functional: users drag-and-drop (or click to select) a file, it uploads to Railway Buckets (S3-compatible), and a `Document` + `DocumentVersion` record is written to the database. The document list then updates to reflect the new entry.

Everything stays within Next.js — no calls to the Go backend.

---

## Architecture

```
User drops file
    ↓
<UploadZone> (client component)
    ↓ POST multipart/form-data
/api/org/[slug]/knowledge/[kbId]/documents  (Next.js Route Handler)
    ├── Validates org + KB ownership (TypeORM)
    ├── Uploads bytes to Railway Bucket (AWS SDK / S3-compatible)
    ├── Inserts Document row (or updates if path already exists)
    ├── Inserts DocumentVersion row (status = STORED)
    └── Returns created document
    ↓ 201 Created
<UploadZone> calls router.refresh()
    ↓
Server component re-runs → document list refreshes
```

---

## Files to Change

### 1. `package.json`
Add `@aws-sdk/client-s3` dependency (Railway Buckets are S3-compatible).

### 2. `.env.local.example`
Add Railway bucket env vars:
```
RAILWAY_BUCKET_ENDPOINT=https://...
RAILWAY_BUCKET_NAME=...
RAILWAY_BUCKET_ACCESS_KEY_ID=...
RAILWAY_BUCKET_SECRET_ACCESS_KEY=...
RAILWAY_BUCKET_REGION=auto
```

### 3. `lib/entities/document-version.entity.ts`
Add missing `raw_content_uri` column (it's `NOT NULL` in the DB schema — inserts fail without it):
```ts
@Column({ type: "text", name: "raw_content_uri" })
rawContentUri!: string;
```

### 4. `app/lib/storage.server.ts` *(new)*
S3 client singleton + upload helper. Keeps bucket concerns isolated and reusable:
```ts
import "server-only";
// S3Client from @aws-sdk/client-s3, configured from env vars
// export async function uploadFileToStorage(key: string, body: Buffer, contentType: string): Promise<string>
// Returns the storage URI (e.g. s3://bucket/key)
```

### 5. `app/api/org/[slug]/knowledge/[kbId]/documents/route.ts`
Add `POST` handler:
- Parse `multipart/form-data` via `request.formData()`
- Validate org + KB ownership (same pattern as existing GET)
- Call `uploadFileToStorage(key, buffer, contentType)` → URI
- Upsert `Document` (unique constraint is `(kb_id, path)`) → get document ID
- Insert `DocumentVersion` with `status = "STORED"`, `is_active = false`, `version_number = 1` (or increment)
- Return `201` with the new document shape matching `IngestionDocument`

Key: storage key pattern → `kb/{kbId}/docs/{documentId}/{filename}`

### 6. `app/(app)/org/[slug]/knowledge/[kbId]/ingestion/_components/UploadZone.tsx` *(new)*
`"use client"` component:
- Accepts `slug` and `kbId` as props
- Hidden `<input type="file">` triggered by click or drag events
- Drag-and-drop handlers: `onDragOver`, `onDragLeave`, `onDrop` (visual state feedback)
- On file select: POST `FormData` to `/api/org/${slug}/knowledge/${kbId}/documents`
- Upload state: idle → uploading → done/error
- On success: calls `router.refresh()` to re-run server component and show new document
- On error: shows inline error message

### 7. `app/(app)/org/[slug]/knowledge/[kbId]/ingestion/page.tsx`
- Replace the static drag-and-drop `<div>` with `<UploadZone slug={slug} kbId={kbId} />`
- Enable (or remove) the "Upload Sources" button — can wire it to trigger the UploadZone's file input via a forwarded ref, or remove it since the drop zone is now interactive

---

## Data Flow Detail

**POST `/api/org/[slug]/knowledge/[kbId]/documents`**

Request: `multipart/form-data`
- `file` — the file blob
- `path` (optional) — defaults to `file.name`
- `title` (optional)

Response `201`:
```json
{
  "id": "...",
  "path": "report.md",
  "title": null,
  "document_type": "text/markdown",
  "processing_status": "STORED",
  "version_number": 1,
  "active_version_id": null,
  "created_at": "...",
  "updated_at": "..."
}
```

**Document upsert logic (on path collision)**:
If a document with the same `(kb_id, path)` already exists, insert a new `DocumentVersion` with `version_number = max + 1`. The existing document row is updated (`updatedAt` refreshes). This matches the backend's design intent where re-uploading the same path = new version.

---

## Verification

1. Run `npm install` after adding `@aws-sdk/client-s3`
2. Set Railway bucket env vars in `.env`
3. Start dev server, navigate to an ingestion page
4. Drag a `.md` file into the upload zone — should show uploading state
5. On success, document appears in the list with status `STORED`
6. Re-uploading same filename creates a new version (version_number increments)
7. Check DB: `documents` and `document_versions` tables have correct rows
8. Check Railway bucket console: file appears at `kb/{kbId}/docs/{docId}/{filename}`
