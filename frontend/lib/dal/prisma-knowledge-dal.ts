import { Prisma } from "@prisma/client";
import { prisma } from "@/lib/prisma";
import type { KnowledgeDal } from "@/lib/dal/contracts";
import type {
  ChunkCountRow,
  CreateOrUpdateDocumentInput,
  CreatedDocumentResult,
  DocumentRecord,
  KnowledgeBaseRecord,
  LatestDocumentVersionRecord,
  OrgRecord,
  RawChunkRow,
} from "@/lib/dal/types";

function asRecord(value: unknown): Record<string, unknown> {
  return value && typeof value === "object" && !Array.isArray(value)
    ? (value as Record<string, unknown>)
    : {};
}

function toInputJson(value: Record<string, unknown>): Prisma.InputJsonValue {
  return value as Prisma.InputJsonValue;
}

function normalizeDate(value: Date | string): Date {
  return value instanceof Date ? value : new Date(value);
}

export class PrismaKnowledgeDal implements KnowledgeDal {
  async ping(): Promise<void> {
    await prisma.$queryRaw`SELECT 1`;
  }

  async getOrganizationBySlug(slug: string): Promise<OrgRecord | null> {
    const org = await prisma.organization.findUnique({
      where: { slug },
      select: { id: true, slug: true },
    });

    return org ? { id: org.id, slug: org.slug } : null;
  }

  async getOrganizationSlugById(id: string): Promise<string | null> {
    const org = await prisma.organization.findUnique({ where: { id }, select: { slug: true } });
    return org?.slug ?? null;
  }

  async listKnowledgeBasesForOrg(orgId: string, orgSlug: string): Promise<KnowledgeBaseRecord[]> {
    const rows = await prisma.$queryRaw<Array<KnowledgeBaseRecord>>`
      SELECT
        id,
        name,
        metadata,
        created_at AS "createdAt",
        updated_at AS "updatedAt"
      FROM knowledge_bases
      WHERE metadata ->> 'organization_id' = ${orgId}
         OR metadata ->> 'org_slug' = ${orgSlug}
      ORDER BY updated_at DESC
    `;

    return rows.map((row) => ({
      ...row,
      metadata: asRecord(row.metadata),
      createdAt: normalizeDate(row.createdAt),
      updatedAt: normalizeDate(row.updatedAt),
    }));
  }

  async createKnowledgeBase(name: string, metadata: Record<string, unknown>): Promise<KnowledgeBaseRecord> {
    const kb = await prisma.knowledgeBase.create({
      data: {
        id: crypto.randomUUID(),
        name,
        metadata: toInputJson(metadata),
      },
    });

    return {
      id: kb.id,
      name: kb.name,
      metadata: asRecord(kb.metadata),
      createdAt: kb.createdAt,
      updatedAt: kb.updatedAt,
    };
  }

  async getKnowledgeBaseById(id: string): Promise<KnowledgeBaseRecord | null> {
    const kb = await prisma.knowledgeBase.findUnique({
      where: { id },
      select: {
        id: true,
        name: true,
        metadata: true,
        createdAt: true,
        updatedAt: true,
      },
    });

    if (!kb) {
      return null;
    }

    return {
      id: kb.id,
      name: kb.name,
      metadata: asRecord(kb.metadata),
      createdAt: kb.createdAt,
      updatedAt: kb.updatedAt,
    };
  }

  async getDocumentIdByKbPath(kbId: string, path: string): Promise<string | null> {
    const document = await prisma.document.findFirst({
      where: { kbId, path },
      select: { id: true },
    });

    return document?.id ?? null;
  }

  async createOrUpdateDocumentAndInsertVersion(input: CreateOrUpdateDocumentInput): Promise<CreatedDocumentResult> {
    return prisma.$transaction(async (tx) => {
      let document = await tx.document.findFirst({
        where: { kbId: input.kbId, path: input.documentPath },
      });

      if (!document) {
        document = await tx.document.create({
          data: {
            id: input.documentId,
            kbId: input.kbId,
            path: input.documentPath,
            title: input.documentTitle,
            documentType: input.documentType,
            sourceMetadata: toInputJson({
              original_filename: input.originalFilename,
              size_bytes: input.sizeBytes,
            }),
            activeVersionId: null,
          },
        });
      } else {
        document = await tx.document.update({
          where: { id: document.id },
          data: {
            title: input.documentTitle,
            documentType: input.documentType,
            sourceMetadata: toInputJson({
              ...asRecord(document.sourceMetadata),
              original_filename: input.originalFilename,
              size_bytes: input.sizeBytes,
            }),
          },
        });
      }

      const aggregate = await tx.documentVersion.aggregate({
        _max: { versionNumber: true },
        where: { documentId: document.id },
      });
      const versionNumber = (aggregate._max.versionNumber ?? 0) + 1;

      await tx.documentVersion.create({
        data: {
          id: crypto.randomUUID(),
          documentId: document.id,
          kbId: input.kbId,
          versionNumber,
          rawContentUri: input.rawContentUri,
          processingStatus: "STORED",
          isActive: false,
        },
      });

      return {
        id: document.id,
        path: document.path,
        title: document.title,
        documentType: document.documentType,
        sourceMetadata: asRecord(document.sourceMetadata),
        activeVersionId: document.activeVersionId,
        processingStatus: "STORED",
        versionNumber,
        createdAt: document.createdAt,
        updatedAt: document.updatedAt,
      };
    });
  }

  async deleteDocumentById(kbId: string, documentId: string): Promise<boolean> {
    const deleted = await prisma.document.deleteMany({ where: { id: documentId, kbId } });
    return deleted.count > 0;
  }

  async listDocumentsForKb(kbId: string): Promise<DocumentRecord[]> {
    const rows = await prisma.document.findMany({ where: { kbId }, orderBy: { updatedAt: "desc" } });

    return rows.map((doc) => ({
      id: doc.id,
      kbId: doc.kbId,
      path: doc.path,
      title: doc.title,
      documentType: doc.documentType,
      sourceMetadata: asRecord(doc.sourceMetadata),
      activeVersionId: doc.activeVersionId,
      createdAt: doc.createdAt,
      updatedAt: doc.updatedAt,
    }));
  }

  async listLatestVersionsForDocuments(documentIds: string[]): Promise<LatestDocumentVersionRecord[]> {
    if (documentIds.length === 0) {
      return [];
    }

    const values = Prisma.join(documentIds);
    const rows = await prisma.$queryRaw<Array<{ document_id: string; processing_status: string; version_number: number }>>`
      SELECT DISTINCT ON (dv.document_id)
        dv.document_id,
        dv.processing_status,
        dv.version_number
      FROM document_versions dv
      WHERE dv.document_id IN (${values})
      ORDER BY dv.document_id ASC, dv.version_number DESC
    `;

    return rows.map((row) => ({
      documentId: row.document_id,
      processingStatus: row.processing_status,
      versionNumber: Number(row.version_number),
    }));
  }

  async listActiveChunksForKb(kbId: string): Promise<RawChunkRow[]> {
    const rows = await prisma.$queryRaw<Array<Omit<RawChunkRow, "created_at"> & { created_at: Date | string }>>`
      SELECT
        c.id AS id,
        dv.document_id AS document_id,
        c.document_version_id AS document_version_id,
        c.sequence_number AS sequence_number,
        c.content AS content,
        c.content_hash AS content_hash,
        c.metadata AS metadata,
        c.chunking_strategy AS chunking_strategy,
        c.embedding_id AS embedding_id,
        c.created_at AS created_at
      FROM chunks c
      INNER JOIN document_versions dv ON dv.id = c.document_version_id
      INNER JOIN documents d ON d.id = dv.document_id
      WHERE c.kb_id = ${kbId}
        AND dv.is_active = TRUE
        AND d.kb_id = ${kbId}
      ORDER BY dv.document_id ASC, c.sequence_number ASC, c.created_at ASC
    `;

    return rows.map((row) => ({
      ...row,
      created_at: normalizeDate(row.created_at),
    }));
  }

  async listChunkCountsForKb(kbId: string): Promise<ChunkCountRow[]> {
    const rows = await prisma.$queryRaw<Array<{ document_id: string; chunk_count: number | string }>>`
      SELECT
        d.id AS document_id,
        COUNT(c.id)::int AS chunk_count
      FROM documents d
      LEFT JOIN document_versions dv
        ON dv.document_id = d.id
       AND dv.is_active = TRUE
      LEFT JOIN chunks c
        ON c.document_version_id = dv.id
       AND c.kb_id = ${kbId}
      WHERE d.kb_id = ${kbId}
      GROUP BY d.id
    `;

    return rows.map((row) => ({
      document_id: row.document_id,
      chunk_count: Number(row.chunk_count),
    }));
  }

  async listActiveChunksForDocument(kbId: string, documentId: string): Promise<RawChunkRow[]> {
    const rows = await prisma.$queryRaw<Array<Omit<RawChunkRow, "created_at"> & { created_at: Date | string }>>`
      SELECT
        c.id AS id,
        dv.document_id AS document_id,
        c.document_version_id AS document_version_id,
        c.sequence_number AS sequence_number,
        c.content AS content,
        c.content_hash AS content_hash,
        c.metadata AS metadata,
        c.chunking_strategy AS chunking_strategy,
        c.embedding_id AS embedding_id,
        c.created_at AS created_at
      FROM chunks c
      INNER JOIN document_versions dv ON dv.id = c.document_version_id
      INNER JOIN documents d ON d.id = dv.document_id
      WHERE c.kb_id = ${kbId}
        AND d.kb_id = ${kbId}
        AND dv.is_active = TRUE
        AND dv.document_id = ${documentId}
      ORDER BY c.sequence_number ASC, c.created_at ASC
    `;

    return rows.map((row) => ({
      ...row,
      created_at: normalizeDate(row.created_at),
    }));
  }
}

export const prismaKnowledgeDal = new PrismaKnowledgeDal();
