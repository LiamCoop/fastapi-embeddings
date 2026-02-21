import "server-only";
import { getKnowledgeDal } from "@/lib/dal";
import type { KnowledgeBaseRecord } from "@/lib/dal/types";
import {
  KNOWLEDGE_CACHE_REVALIDATE_SECONDS,
  knowledgeKbChunkCountsTag,
  knowledgeKbDocumentChunksTag,
  knowledgeKbDocumentsTag,
  knowledgeKbChunksTag,
  knowledgeKbTag,
  knowledgeOrgTag,
} from "@/app/lib/knowledge-cache";
import { unstable_cache } from "next/cache";
import type {
  KnowledgeChunksByDocumentResponse,
  KnowledgeChunkCountsResponse,
  KnowledgeDocumentChunksResponse,
  IngestionResponse,
  IngestionWithChunksResponse,
  KnowledgeChunk,
} from "@/app/lib/org-knowledge";

async function loadDocumentsForKb(
  kb: KnowledgeBaseRecord,
): Promise<IngestionResponse> {
  const dal = getKnowledgeDal();
  const rawDocuments = await dal.listDocumentsForKb(kb.id);
  const docIds = rawDocuments.map((doc) => doc.id);
  const latestVersions = await dal.listLatestVersionsForDocuments(docIds);

  const latestVersionByDocumentId = new Map(latestVersions.map((row) => [row.documentId, row]));

  const documents = rawDocuments.map((doc) => ({
    id: doc.id,
    path: doc.path,
    title: doc.title,
    document_type: doc.documentType,
    source_metadata: doc.sourceMetadata ?? {},
    active_version_id: doc.activeVersionId,
    processing_status: latestVersionByDocumentId.get(doc.id)?.processingStatus ?? null,
    version_number: latestVersionByDocumentId.get(doc.id)?.versionNumber ?? null,
    created_at: doc.createdAt.toISOString(),
    updated_at: doc.updatedAt.toISOString(),
  }));

  console.warn("[org-knowledge] loadDocumentsForKb", {
    kbId: kb.id,
    documentCount: documents.length,
  });

  return {
    knowledge_base: {
      id: kb.id,
      name: kb.name,
      metadata: kb.metadata ?? {},
      created_at: kb.createdAt.toISOString(),
      updated_at: kb.updatedAt.toISOString(),
    },
    documents,
  };
}

async function loadOrgKnowledgeDocumentsUncached(
  slug: string,
  kbId: string,
): Promise<IngestionResponse | null> {
  const dal = getKnowledgeDal();

  if (!slug) {
    return null;
  }

  const kb = await dal.getKnowledgeBaseById(kbId);
  if (!kb) {
    return null;
  }

  const kbOrgSlug = kb.metadata?.org_slug;
  if (kbOrgSlug !== slug) {
    return null;
  }

  return loadDocumentsForKb(kb);
}

export async function loadOrgKnowledgeDocuments(
  slug: string,
  kbId: string,
): Promise<IngestionResponse | null> {
  if (!slug || !kbId) {
    return null;
  }

  return unstable_cache(
    () => loadOrgKnowledgeDocumentsUncached(slug, kbId),
    ["org-knowledge-documents", "v1", slug, kbId],
    {
      revalidate: KNOWLEDGE_CACHE_REVALIDATE_SECONDS,
      tags: [knowledgeOrgTag(slug), knowledgeKbTag(kbId), knowledgeKbDocumentsTag(kbId)],
    },
  )();
}

export async function loadKnowledgeDocumentsByKbId(kbId: string): Promise<IngestionResponse | null> {
  console.warn("[org-knowledge] loadKnowledgeDocumentsByKbId:start", { kbId });

  const dal = getKnowledgeDal();
  const kb = await dal.getKnowledgeBaseById(kbId);

  if (!kb) {
    console.warn("[org-knowledge] loadKnowledgeDocumentsByKbId:kb_not_found", { kbId });
    return null;
  }

  const metadata = kb.metadata ?? {};
  const metadataSlug = metadata.org_slug;
  if (typeof metadataSlug === "string" && metadataSlug.trim().length > 0) {
    console.warn("[org-knowledge] loadKnowledgeDocumentsByKbId:using_metadata_slug", {
      kbId,
      metadataSlug,
    });
    const result = await loadOrgKnowledgeDocuments(metadataSlug, kbId);
    console.warn("[org-knowledge] loadKnowledgeDocumentsByKbId:metadata_slug_result", {
      kbId,
      found: Boolean(result),
      documentCount: result?.documents.length ?? 0,
    });
    return result;
  }

  const metadataOrgId = metadata.organization_id;
  if (typeof metadataOrgId === "string" && metadataOrgId.trim().length > 0) {
    console.warn("[org-knowledge] loadKnowledgeDocumentsByKbId:using_metadata_org_id", {
      kbId,
      metadataOrgId,
    });
    const orgSlug = await dal.getOrganizationSlugById(metadataOrgId);
    if (orgSlug) {
      const result = await loadOrgKnowledgeDocuments(orgSlug, kbId);
      console.warn("[org-knowledge] loadKnowledgeDocumentsByKbId:metadata_org_id_result", {
        kbId,
        resolvedSlug: orgSlug,
        found: Boolean(result),
        documentCount: result?.documents.length ?? 0,
      });
      return result;
    }
    console.warn("[org-knowledge] loadKnowledgeDocumentsByKbId:org_not_found_for_metadata_org_id", {
      kbId,
      metadataOrgId,
    });
  }

  console.warn("[org-knowledge] loadKnowledgeDocumentsByKbId:fallback_direct_kb", { kbId });
  return loadDocumentsForKb(kb);
}

export async function loadOrgKnowledgeDocumentsWithChunks(
  slug: string,
  kbId: string,
): Promise<IngestionWithChunksResponse | null> {
  if (!slug || !kbId) {
    return null;
  }

  return unstable_cache(
    async () => {
      const documentsData = await loadOrgKnowledgeDocuments(slug, kbId);
      if (!documentsData) {
        return null;
      }

      const dal = getKnowledgeDal();
      const rows = await dal.listActiveChunksForKb(kbId);

      const chunksByDocumentId: Record<string, KnowledgeChunk[]> = Object.create(null) as Record<
        string,
        KnowledgeChunk[]
      >;

      for (const row of rows) {
        if (!chunksByDocumentId[row.document_id]) {
          chunksByDocumentId[row.document_id] = [];
        }

        chunksByDocumentId[row.document_id].push({
          id: row.id,
          document_id: row.document_id,
          document_version_id: row.document_version_id,
          sequence_number: row.sequence_number,
          content: row.content,
          content_hash: row.content_hash,
          metadata: row.metadata ?? {},
          chunking_strategy: row.chunking_strategy,
          embedding_id: row.embedding_id,
          created_at: row.created_at.toISOString(),
        });
      }

      for (const document of documentsData.documents) {
        if (!chunksByDocumentId[document.id]) {
          chunksByDocumentId[document.id] = [];
        }
      }

      return {
        ...documentsData,
        chunks_by_document_id: chunksByDocumentId,
      };
    },
    ["org-knowledge-documents-with-chunks", "v1", slug, kbId],
    {
      revalidate: KNOWLEDGE_CACHE_REVALIDATE_SECONDS,
      tags: [
        knowledgeOrgTag(slug),
        knowledgeKbTag(kbId),
        knowledgeKbDocumentsTag(kbId),
        knowledgeKbChunksTag(kbId),
      ],
    },
  )();
}

export async function loadOrgKnowledgeChunksByDocument(
  slug: string,
  kbId: string,
): Promise<KnowledgeChunksByDocumentResponse | null> {
  if (!slug || !kbId) {
    return null;
  }

  return unstable_cache(
    async () => {
      const documentsData = await loadOrgKnowledgeDocuments(slug, kbId);
      if (!documentsData) {
        return null;
      }

      const dal = getKnowledgeDal();
      const rows = await dal.listActiveChunksForKb(kbId);

      const chunksByDocumentId: Record<string, KnowledgeChunk[]> = Object.create(null) as Record<
        string,
        KnowledgeChunk[]
      >;

      for (const document of documentsData.documents) {
        chunksByDocumentId[document.id] = [];
      }

      for (const row of rows) {
        const createdAt =
          row.created_at instanceof Date ? row.created_at.toISOString() : new Date(row.created_at).toISOString();

        if (!chunksByDocumentId[row.document_id]) {
          chunksByDocumentId[row.document_id] = [];
        }

        chunksByDocumentId[row.document_id].push({
          id: row.id,
          document_id: row.document_id,
          document_version_id: row.document_version_id,
          sequence_number: row.sequence_number,
          content: row.content,
          content_hash: row.content_hash,
          metadata: row.metadata ?? {},
          chunking_strategy: row.chunking_strategy,
          embedding_id: row.embedding_id,
          created_at: createdAt,
        });
      }

      return {
        kb_id: kbId,
        chunks_by_document_id: chunksByDocumentId,
        total_chunks: rows.length,
      };
    },
    ["org-knowledge-chunks-by-document", "v1", slug, kbId],
    {
      revalidate: KNOWLEDGE_CACHE_REVALIDATE_SECONDS,
      tags: [
        knowledgeOrgTag(slug),
        knowledgeKbTag(kbId),
        knowledgeKbDocumentsTag(kbId),
        knowledgeKbChunksTag(kbId),
      ],
    },
  )();
}

export async function loadOrgKnowledgeChunkCounts(
  slug: string,
  kbId: string,
): Promise<KnowledgeChunkCountsResponse | null> {
  if (!slug || !kbId) {
    return null;
  }

  return unstable_cache(
    async () => {
      const documentsData = await loadOrgKnowledgeDocuments(slug, kbId);
      if (!documentsData) {
        return null;
      }

      const dal = getKnowledgeDal();
      const rows = await dal.listChunkCountsForKb(kbId);

      const chunkCountsByDocumentId: Record<string, number> = Object.create(null) as Record<string, number>;
      for (const document of documentsData.documents) {
        chunkCountsByDocumentId[document.id] = 0;
      }
      for (const row of rows) {
        chunkCountsByDocumentId[row.document_id] = row.chunk_count;
      }

      const totalChunks = Object.values(chunkCountsByDocumentId).reduce((sum, count) => sum + count, 0);
      return {
        kb_id: kbId,
        chunk_counts_by_document_id: chunkCountsByDocumentId,
        total_chunks: totalChunks,
      };
    },
    ["org-knowledge-chunk-counts", "v1", slug, kbId],
    {
      revalidate: KNOWLEDGE_CACHE_REVALIDATE_SECONDS,
      tags: [
        knowledgeOrgTag(slug),
        knowledgeKbTag(kbId),
        knowledgeKbDocumentsTag(kbId),
        knowledgeKbChunkCountsTag(kbId),
      ],
    },
  )();
}

export async function loadOrgKnowledgeChunksForDocument(
  slug: string,
  kbId: string,
  documentId: string,
): Promise<KnowledgeDocumentChunksResponse | null> {
  if (!slug || !kbId || !documentId) {
    return null;
  }

  return unstable_cache(
    async () => {
      const documentsData = await loadOrgKnowledgeDocuments(slug, kbId);
      if (!documentsData) {
        return null;
      }
      const documentExists = documentsData.documents.some((document) => document.id === documentId);
      if (!documentExists) {
        return null;
      }

      const dal = getKnowledgeDal();
      const rows = await dal.listActiveChunksForDocument(kbId, documentId);

      const chunks: KnowledgeChunk[] = rows.map((row) => ({
        id: row.id,
        document_id: row.document_id,
        document_version_id: row.document_version_id,
        sequence_number: row.sequence_number,
        content: row.content,
        content_hash: row.content_hash,
        metadata: row.metadata ?? {},
        chunking_strategy: row.chunking_strategy,
        embedding_id: row.embedding_id,
        created_at:
          row.created_at instanceof Date ? row.created_at.toISOString() : new Date(row.created_at).toISOString(),
      }));

      return {
        kb_id: kbId,
        document_id: documentId,
        chunks,
        chunk_count: chunks.length,
      };
    },
    ["org-knowledge-document-chunks", "v1", slug, kbId, documentId],
    {
      revalidate: KNOWLEDGE_CACHE_REVALIDATE_SECONDS,
      tags: [
        knowledgeOrgTag(slug),
        knowledgeKbTag(kbId),
        knowledgeKbChunksTag(kbId),
        knowledgeKbDocumentChunksTag(kbId, documentId),
      ],
    },
  )();
}
