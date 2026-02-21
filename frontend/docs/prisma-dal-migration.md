# Prisma DAL Migration (Step 1)

This introduces a Prisma-based DAL alongside the existing TypeORM DAL.

## Toggle

Set `FRONTEND_DAL`:

- `typeorm` (default)
- `prisma`

Example:

```env
FRONTEND_DAL=prisma
```

## Scope in Step 1

DAL-backed (switchable) routes/services:

- `GET/POST /api/org/[slug]/knowledge`
- `GET/POST /api/org/[slug]/knowledge/[kbId]/documents`
- `DELETE /api/org/[slug]/knowledge/[kbId]/documents/[documentId]`
- `app/lib/org-knowledge.server.ts`
- `GET /api/health` (reports active DAL mode)

Still TypeORM-only in this step:

- Clerk webhook sync route
- Dashboard org bootstrap route

## Prisma Setup

1. Install deps:

```bash
npm install
```

2. Generate Prisma client:

```bash
npm run prisma:generate
```

3. Start app (with optional DAL override):

```bash
FRONTEND_DAL=prisma npm run dev
```

## Notes

- Prisma schema is in `frontend/prisma/schema.prisma` and maps to existing table names.
- This step does not remove TypeORM; it only adds a parallel DAL path for incremental migration.
