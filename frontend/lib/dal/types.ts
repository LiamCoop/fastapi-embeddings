export type OrgRecord = {
  id: string;
  slug: string;
};

export type KnowledgeBaseRecord = {
  id: string;
  name: string;
  metadata: Record<string, unknown>;
  createdAt: Date;
  updatedAt: Date;
};

export type DocumentRecord = {
  id: string;
  kbId: string;
  path: string;
  title: string | null;
  documentType: string;
  sourceMetadata: Record<string, unknown>;
  activeVersionId: string | null;
  createdAt: Date;
  updatedAt: Date;
};

export type LatestDocumentVersionRecord = {
  documentId: string;
  processingStatus: string;
  versionNumber: number;
};

export type RawChunkRow = {
  id: string;
  document_id: string;
  document_version_id: string;
  sequence_number: number;
  content: string;
  content_hash: string;
  metadata: Record<string, unknown> | null;
  chunking_strategy: string;
  embedding_id: string | null;
  created_at: Date;
};

export type ChunkCountRow = {
  document_id: string;
  chunk_count: number;
};

export type CreatedDocumentResult = {
  id: string;
  path: string;
  title: string | null;
  documentType: string;
  sourceMetadata: Record<string, unknown>;
  activeVersionId: string | null;
  processingStatus: string;
  versionNumber: number;
  createdAt: Date;
  updatedAt: Date;
};

export type CreateOrUpdateDocumentInput = {
  kbId: string;
  documentPath: string;
  documentTitle: string | null;
  documentType: string;
  originalFilename: string;
  sizeBytes: number;
  rawContentUri: string;
  documentId: string;
};
