export const KNOWLEDGE_CACHE_REVALIDATE_SECONDS = 30;

export function knowledgeOrgTag(slug: string): string {
  return `knowledge:org:${slug}`;
}

export function knowledgeKbTag(kbId: string): string {
  return `knowledge:kb:${kbId}`;
}

export function knowledgeKbDocumentsTag(kbId: string): string {
  return `knowledge:kb:${kbId}:documents`;
}

export function knowledgeKbChunksTag(kbId: string): string {
  return `knowledge:kb:${kbId}:chunks`;
}

export function knowledgeKbChunkCountsTag(kbId: string): string {
  return `knowledge:kb:${kbId}:chunk-counts`;
}

export function knowledgeKbDocumentChunksTag(kbId: string, documentId: string): string {
  return `knowledge:kb:${kbId}:document:${documentId}:chunks`;
}
