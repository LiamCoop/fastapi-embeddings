import type { DataSource, EntityManager } from "typeorm";
import { getDb } from "@/lib/db";
import { Document } from "@/lib/entities/document.entity";
import { DocumentVersion } from "@/lib/entities/document-version.entity";
import { KnowledgeBase } from "@/lib/entities/knowledge-base.entity";
import { Organization } from "@/lib/entities/organization.entity";
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

function toDate(value: Date | string): Date {
  return value instanceof Date ? value : new Date(value);
}

async function currentVersionNumberForDocument(manager: EntityManager, documentId: string): Promise<number> {
  const row = await manager
    .getRepository(DocumentVersion)
    .createQueryBuilder("dv")
    .select("COALESCE(MAX(dv.versionNumber), 0)", "max_version")
    .where("dv.documentId = :documentId", { documentId })
    .getRawOne<{ max_version: string | number }>();

  return Number(row?.max_version ?? 0);
}

export class TypeormKnowledgeDal implements KnowledgeDal {
  private async db(): Promise<DataSource> {
    return getDb();
  }

  async ping(): Promise<void> {
    const db = await this.db();
    await db.query("SELECT 1");
  }

  async getOrganizationBySlug(slug: string): Promise<OrgRecord | null> {
    const db = await this.db();
    const org = await db.getRepository(Organization).findOne({
      where: { slug },
      select: ["id", "slug"],
    });

    return org ? { id: org.id, slug: org.slug } : null;
  }

  async getOrganizationSlugById(id: string): Promise<string | null> {
    const db = await this.db();
    const org = await db.getRepository(Organization).findOne({ where: { id }, select: ["slug"] });
    return org?.slug ?? null;
  }

  async listKnowledgeBasesForOrg(orgId: string, orgSlug: string): Promise<KnowledgeBaseRecord[]> {
    const db = await this.db();
    const entities = await db
      .getRepository(KnowledgeBase)
      .createQueryBuilder("kb")
      .where("kb.metadata ->> 'organization_id' = :orgId", { orgId })
      .orWhere("kb.metadata ->> 'org_slug' = :slug", { slug: orgSlug })
      .orderBy("kb.updatedAt", "DESC")
      .getMany();

    return entities.map((kb) => ({
      id: kb.id,
      name: kb.name,
      metadata: asRecord(kb.metadata),
      createdAt: kb.createdAt,
      updatedAt: kb.updatedAt,
    }));
  }

  async createKnowledgeBase(name: string, metadata: Record<string, unknown>): Promise<KnowledgeBaseRecord> {
    const db = await this.db();
    const repo = db.getRepository(KnowledgeBase);
    const saved = await repo.save(
      repo.create({
        id: crypto.randomUUID(),
        name,
        metadata,
      }),
    );

    return {
      id: saved.id,
      name: saved.name,
      metadata: asRecord(saved.metadata),
      createdAt: saved.createdAt,
      updatedAt: saved.updatedAt,
    };
  }

  async getKnowledgeBaseById(id: string): Promise<KnowledgeBaseRecord | null> {
    const db = await this.db();
    const kb = await db.getRepository(KnowledgeBase).findOne({
      where: { id },
      select: ["id", "name", "metadata", "createdAt", "updatedAt"],
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
    const db = await this.db();
    const document = await db.getRepository(Document).findOne({ where: { kbId, path }, select: ["id"] });
    return document?.id ?? null;
  }

  async createOrUpdateDocumentAndInsertVersion(input: CreateOrUpdateDocumentInput): Promise<CreatedDocumentResult> {
    const db = await this.db();

    return db.transaction(async (manager) => {
      let document = await manager.getRepository(Document).findOne({
        where: { kbId: input.kbId, path: input.documentPath },
      });

      if (!document) {
        document = manager.getRepository(Document).create({
          id: input.documentId,
          kbId: input.kbId,
          path: input.documentPath,
          title: input.documentTitle,
          documentType: input.documentType,
          sourceMetadata: {
            original_filename: input.originalFilename,
            size_bytes: input.sizeBytes,
          },
          activeVersionId: null,
        });
      } else {
        document.title = input.documentTitle;
        document.documentType = input.documentType;
        document.sourceMetadata = {
          ...asRecord(document.sourceMetadata),
          original_filename: input.originalFilename,
          size_bytes: input.sizeBytes,
        };
      }

      const savedDocument = await manager.getRepository(Document).save(document);
      const versionNumber = (await currentVersionNumberForDocument(manager, savedDocument.id)) + 1;

      await manager.getRepository(DocumentVersion).save(
        manager.getRepository(DocumentVersion).create({
          id: crypto.randomUUID(),
          documentId: savedDocument.id,
          kbId: input.kbId,
          versionNumber,
          rawContentUri: input.rawContentUri,
          processingStatus: "STORED",
          isActive: false,
        }),
      );

      return {
        id: savedDocument.id,
        path: savedDocument.path,
        title: savedDocument.title,
        documentType: savedDocument.documentType,
        sourceMetadata: asRecord(savedDocument.sourceMetadata),
        activeVersionId: savedDocument.activeVersionId,
        processingStatus: "STORED",
        versionNumber,
        createdAt: savedDocument.createdAt,
        updatedAt: savedDocument.updatedAt,
      };
    });
  }

  async deleteDocumentById(kbId: string, documentId: string): Promise<boolean> {
    const db = await this.db();
    const result = await db.getRepository(Document).delete({ id: documentId, kbId });
    return Boolean(result.affected);
  }

  async listDocumentsForKb(kbId: string): Promise<DocumentRecord[]> {
    const db = await this.db();
    const rows = await db.getRepository(Document).find({ where: { kbId }, order: { updatedAt: "DESC" } });

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

    const db = await this.db();
    const rows = await db
      .getRepository(DocumentVersion)
      .createQueryBuilder("dv")
      .distinctOn(["dv.documentId"])
      .select([
        "dv.documentId AS document_id",
        "dv.processingStatus AS processing_status",
        "dv.versionNumber AS version_number",
      ])
      .where("dv.documentId IN (:...documentIds)", { documentIds })
      .orderBy("dv.documentId", "ASC")
      .addOrderBy("dv.versionNumber", "DESC")
      .getRawMany<{
        document_id: string;
        processing_status: string;
        version_number: number;
      }>();

    return rows.map((row) => ({
      documentId: row.document_id,
      processingStatus: row.processing_status,
      versionNumber: Number(row.version_number),
    }));
  }

  async listActiveChunksForKb(kbId: string): Promise<RawChunkRow[]> {
    const db = await this.db();
    const rows = (await db.query(
      `
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
        WHERE c.kb_id = $1
          AND dv.is_active = TRUE
          AND d.kb_id = $1
        ORDER BY dv.document_id ASC, c.sequence_number ASC, c.created_at ASC
      `,
      [kbId],
    )) as Array<Omit<RawChunkRow, "created_at"> & { created_at: Date | string }>;

    return rows.map((row) => ({ ...row, created_at: toDate(row.created_at) }));
  }

  async listChunkCountsForKb(kbId: string): Promise<ChunkCountRow[]> {
    const db = await this.db();
    const rows = (await db.query(
      `
        SELECT
          d.id AS document_id,
          COUNT(c.id)::int AS chunk_count
        FROM documents d
        LEFT JOIN document_versions dv
          ON dv.document_id = d.id
         AND dv.is_active = TRUE
        LEFT JOIN chunks c
          ON c.document_version_id = dv.id
         AND c.kb_id = $1
        WHERE d.kb_id = $1
        GROUP BY d.id
      `,
      [kbId],
    )) as Array<{ document_id: string; chunk_count: number | string }>;

    return rows.map((row) => ({
      document_id: row.document_id,
      chunk_count: Number(row.chunk_count),
    }));
  }

  async listActiveChunksForDocument(kbId: string, documentId: string): Promise<RawChunkRow[]> {
    const db = await this.db();
    const rows = (await db.query(
      `
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
        WHERE c.kb_id = $1
          AND d.kb_id = $1
          AND dv.is_active = TRUE
          AND dv.document_id = $2
        ORDER BY c.sequence_number ASC, c.created_at ASC
      `,
      [kbId, documentId],
    )) as Array<Omit<RawChunkRow, "created_at"> & { created_at: Date | string }>;

    return rows.map((row) => ({ ...row, created_at: toDate(row.created_at) }));
  }
}

export const typeormKnowledgeDal = new TypeormKnowledgeDal();
