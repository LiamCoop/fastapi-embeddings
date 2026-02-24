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

export interface KnowledgeDal {
  ping(): Promise<void>;
  getOrganizationBySlug(slug: string): Promise<OrgRecord | null>;
  getOrganizationSlugById(id: string): Promise<string | null>;
  listKnowledgeBasesForOrg(orgId: string, orgSlug: string): Promise<KnowledgeBaseRecord[]>;
  createKnowledgeBase(name: string, metadata: Record<string, unknown>): Promise<KnowledgeBaseRecord>;
  getKnowledgeBaseById(id: string): Promise<KnowledgeBaseRecord | null>;
  getDocumentIdByKbPath(kbId: string, path: string): Promise<string | null>;
  createOrUpdateDocumentAndInsertVersion(input: CreateOrUpdateDocumentInput): Promise<CreatedDocumentResult>;
  deleteDocumentById(kbId: string, documentId: string): Promise<boolean>;
  listDocumentsForKb(kbId: string): Promise<DocumentRecord[]>;
  listLatestVersionsForDocuments(documentIds: string[]): Promise<LatestDocumentVersionRecord[]>;
  listActiveChunksForKb(kbId: string): Promise<RawChunkRow[]>;
  listChunkCountsForKb(kbId: string): Promise<ChunkCountRow[]>;
  listActiveChunksForDocument(kbId: string, documentId: string): Promise<RawChunkRow[]>;
}
