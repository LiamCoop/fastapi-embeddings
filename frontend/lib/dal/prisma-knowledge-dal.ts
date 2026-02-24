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

export class PrismaKnowledgeDal implements KnowledgeDal {
  async ping(): Promise<void> {
    await prisma.organization.count();
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
    const rows = await prisma.knowledgeBase.findMany({
      where: {
        OR: [
          { metadata: { path: ["organization_id"], equals: orgId } },
          { metadata: { path: ["org_slug"], equals: orgSlug } },
        ],
      },
      orderBy: { updatedAt: "desc" },
      select: {
        id: true,
        name: true,
        metadata: true,
        createdAt: true,
        updatedAt: true,
      },
    });

    return rows.map((row) => ({
      ...row,
      metadata: asRecord(row.metadata),
      createdAt: row.createdAt,
      updatedAt: row.updatedAt,
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

  async listRawContentUrisForKb(kbId: string): Promise<string[]> {
    const rows = await prisma.documentVersion.findMany({
      where: { kbId },
      distinct: ["rawContentUri"],
      select: { rawContentUri: true },
    });

    return rows.map((row) => row.rawContentUri);
  }

  async deleteKnowledgeBaseById(kbId: string): Promise<boolean> {
    const deleted = await prisma.knowledgeBase.deleteMany({ where: { id: kbId } });
    return deleted.count > 0;
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

    const rows = await prisma.documentVersion.findMany({
      where: { documentId: { in: documentIds } },
      orderBy: [{ documentId: "asc" }, { versionNumber: "desc" }],
      distinct: ["documentId"],
      select: {
        documentId: true,
        processingStatus: true,
        versionNumber: true,
      },
    });

    return rows.map((row) => ({
      documentId: row.documentId,
      processingStatus: row.processingStatus,
      versionNumber: row.versionNumber,
    }));
  }

  async listActiveChunksForKb(kbId: string): Promise<RawChunkRow[]> {
    const activeVersions = await prisma.documentVersion.findMany({
      where: { kbId, isActive: true },
      select: { id: true, documentId: true },
    });
    if (activeVersions.length === 0) {
      return [];
    }

    const documentIdByVersionId = new Map(activeVersions.map((version) => [version.id, version.documentId]));
    const chunks = await prisma.chunk.findMany({
      where: {
        kbId,
        documentVersionId: { in: activeVersions.map((version) => version.id) },
      },
      orderBy: [{ sequenceNumber: "asc" }, { createdAt: "asc" }],
      select: {
        id: true,
        documentVersionId: true,
        sequenceNumber: true,
        content: true,
        contentHash: true,
        metadata: true,
        chunkingStrategy: true,
        embeddingId: true,
        createdAt: true,
      },
    });

    return chunks
      .map((chunk) => ({
        id: chunk.id,
        document_id: documentIdByVersionId.get(chunk.documentVersionId) ?? "",
        document_version_id: chunk.documentVersionId,
        sequence_number: chunk.sequenceNumber,
        content: chunk.content,
        content_hash: chunk.contentHash,
        metadata: asRecord(chunk.metadata),
        chunking_strategy: chunk.chunkingStrategy,
        embedding_id: chunk.embeddingId,
        created_at: chunk.createdAt,
      }))
      .filter((chunk) => chunk.document_id !== "")
      .sort((a, b) => {
        if (a.document_id < b.document_id) return -1;
        if (a.document_id > b.document_id) return 1;
        if (a.sequence_number !== b.sequence_number) return a.sequence_number - b.sequence_number;
        return a.created_at.getTime() - b.created_at.getTime();
      });
  }

  async listChunkCountsForKb(kbId: string): Promise<ChunkCountRow[]> {
    const documents = await prisma.document.findMany({
      where: { kbId },
      select: { id: true },
    });

    const activeVersions = await prisma.documentVersion.findMany({
      where: { kbId, isActive: true },
      select: { id: true, documentId: true },
    });

    if (activeVersions.length === 0) {
      return documents.map((document) => ({ document_id: document.id, chunk_count: 0 }));
    }

    const versionToDocument = new Map(activeVersions.map((version) => [version.id, version.documentId]));
    const grouped = await prisma.chunk.groupBy({
      by: ["documentVersionId"],
      where: {
        kbId,
        documentVersionId: { in: activeVersions.map((version) => version.id) },
      },
      _count: { _all: true },
    });

    const countsByDocument = new Map<string, number>();
    for (const row of grouped) {
      const documentId = versionToDocument.get(row.documentVersionId);
      if (!documentId) {
        continue;
      }
      countsByDocument.set(documentId, (countsByDocument.get(documentId) ?? 0) + row._count._all);
    }

    return documents.map((document) => ({
      document_id: document.id,
      chunk_count: countsByDocument.get(document.id) ?? 0,
    }));
  }

  async listActiveChunksForDocument(kbId: string, documentId: string): Promise<RawChunkRow[]> {
    const activeVersions = await prisma.documentVersion.findMany({
      where: { kbId, documentId, isActive: true },
      select: { id: true },
    });
    if (activeVersions.length === 0) {
      return [];
    }

    const rows = await prisma.chunk.findMany({
      where: {
        kbId,
        documentVersionId: { in: activeVersions.map((version) => version.id) },
      },
      orderBy: [{ sequenceNumber: "asc" }, { createdAt: "asc" }],
      select: {
        id: true,
        documentVersionId: true,
        sequenceNumber: true,
        content: true,
        contentHash: true,
        metadata: true,
        chunkingStrategy: true,
        embeddingId: true,
        createdAt: true,
      },
    });

    return rows.map((row) => ({
      id: row.id,
      document_id: documentId,
      document_version_id: row.documentVersionId,
      sequence_number: row.sequenceNumber,
      content: row.content,
      content_hash: row.contentHash,
      metadata: asRecord(row.metadata),
      chunking_strategy: row.chunkingStrategy,
      embedding_id: row.embeddingId,
      created_at: row.createdAt,
    }));
  }
}

export const prismaKnowledgeDal = new PrismaKnowledgeDal();
