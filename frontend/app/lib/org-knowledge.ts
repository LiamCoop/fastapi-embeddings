export type KnowledgeBase = {
  id: string;
  name: string;
  metadata: Record<string, unknown>;
  created_at: string;
  updated_at: string;
};

export type KnowledgeBaseListResponse = {
  knowledge_bases: KnowledgeBase[];
};

export type IngestionDocument = {
  id: string;
  path: string;
  title: string | null;
  document_type: string;
  source_metadata: Record<string, unknown>;
  active_version_id: string | null;
  processing_status: string | null;
  version_number: number | null;
  created_at: string;
  updated_at: string;
};

export type IngestionResponse = {
  knowledge_base: KnowledgeBase;
  documents: IngestionDocument[];
};

export type KnowledgeChunk = {
  id: string;
  document_id: string;
  document_version_id: string;
  sequence_number: number;
  content: string;
  content_hash: string;
  metadata: Record<string, unknown>;
  chunking_strategy: string;
  embedding_id: string | null;
  created_at: string;
};

export type IngestionWithChunksResponse = IngestionResponse & {
  chunks_by_document_id: Record<string, KnowledgeChunk[]>;
};

export type KnowledgeChunksByDocumentResponse = {
  kb_id: string;
  chunks_by_document_id: Record<string, KnowledgeChunk[]>;
  total_chunks: number;
};

export type KnowledgeChunkCountsResponse = {
  kb_id: string;
  chunk_counts_by_document_id: Record<string, number>;
  total_chunks: number;
};

export type KnowledgeDocumentChunksResponse = {
  kb_id: string;
  document_id: string;
  chunks: KnowledgeChunk[];
  chunk_count: number;
};

export type KnowledgeChunkEmbedResponse = {
  chunk_id: string;
  embedding_id: string;
  reused: boolean;
};

export type KnowledgeRetrievalRequestFilters = {
  project_id?: string;
  path_prefix?: string;
  document_type?: string;
  source?: string;
  tags?: string[];
  updated_after?: string;
  created_after?: string;
  created_before?: string;
};

export type KnowledgeRetrievalRequest = {
  query: string;
  top_k?: number;
  retrieval_profile?: "auto" | "exact" | "balanced" | "semantic";
  semantic_weight?: number;
  hybrid_weight?: number;
  debug?: boolean;
  filters?: KnowledgeRetrievalRequestFilters;
};

export type KnowledgeRetrievalScore = {
  semantic: number;
  lexical: number;
  final: number;
};

export type KnowledgeRetrievalCitation = {
  document_id: string;
  document_version_id: string;
  path: string;
  title?: string | null;
  version_number: number;
  chunk_sequence: number;
  start_rune?: number;
  end_rune?: number;
  rune_length?: number;
};

export type KnowledgeRetrievalResult = {
  chunk_id: string;
  document_id: string;
  document_version_id: string;
  document_path: string;
  document_title?: string | null;
  document_type: string;
  content: string;
  metadata: Record<string, unknown>;
  scores: KnowledgeRetrievalScore;
  citation: KnowledgeRetrievalCitation;
};

export type KnowledgeRetrievalResponse = {
  request_id: string;
  query_id?: string;
  index_version?: string;
  kb_id: string;
  query: string;
  top_k: number;
  hybrid_weight: number;
  result_count: number;
  latency_ms: number;
  results: KnowledgeRetrievalResult[];
  passages?: KnowledgeRetrievalResult[];
  debug?: {
    retrieval_profile_effective: string;
    semantic_weight_effective: number;
    auto_signals_detected?: string[];
    lexical_candidates: number;
    semantic_candidates: number;
    reranker_applied: boolean;
    filters_applied?: Record<string, unknown>;
  };
};

export type KnowledgeHydrateRequest = {
  chunk_ids: string[];
  adjacent_before?: number;
  adjacent_after?: number;
};

export type KnowledgeHydrateResponse = {
  kb_id: string;
  chunk_count: number;
  chunks: KnowledgeRetrievalResult[];
};

async function fetchJson<T>(input: string, init?: RequestInit): Promise<T> {
  const res = await fetch(input, {
    ...init,
    headers: {
      "Content-Type": "application/json",
      ...(init?.headers ?? {}),
    },
  });

  if (!res.ok) {
    const message = await res.text();
    throw new Error(message || `Request failed with status ${res.status}`);
  }

  const contentType = res.headers.get("content-type") ?? "";
  if (!contentType.includes("application/json")) {
    const body = await res.text();
    const preview = body.slice(0, 120).replace(/\s+/g, " ").trim();
    throw new Error(
      `Expected JSON from ${input}, received ${contentType || "unknown content type"} (${preview})`,
    );
  }

  return (await res.json()) as T;
}

export function knowledgeApiPath(slug: string): string {
  return `/api/org/${slug}/knowledge`;
}

export function knowledgeDocumentsApiPath(slug: string, kbId: string): string {
  return `/api/org/${slug}/knowledge/${kbId}/documents`;
}

export function knowledgeDocumentChunkingApiPath(
  slug: string,
  kbId: string,
  documentId: string,
): string {
  return `/api/org/${slug}/knowledge/${kbId}/documents/${documentId}/chunking`;
}

export function knowledgeDocumentApiPath(slug: string, kbId: string, documentId: string): string {
  return `/api/org/${slug}/knowledge/${kbId}/documents/${documentId}`;
}

export function knowledgeChunksApiPath(slug: string, kbId: string): string {
  return `/api/org/${slug}/knowledge/${kbId}/chunks`;
}

export function knowledgeChunkCountsApiPath(slug: string, kbId: string): string {
  return `/api/org/${slug}/knowledge/${kbId}/chunks/counts`;
}

export function knowledgeDocumentChunksApiPath(slug: string, kbId: string, documentId: string): string {
  return `/api/org/${slug}/knowledge/${kbId}/documents/${documentId}/chunks`;
}

export function knowledgeChunkEmbedApiPath(slug: string, kbId: string, chunkId: string): string {
  return `/api/org/${slug}/knowledge/${kbId}/chunks/${chunkId}/embed`;
}

export function knowledgeRetrieveApiPath(slug: string, kbId: string): string {
  return `/api/org/${slug}/knowledge/${kbId}/retrieve`;
}

export function knowledgeHydrateApiPath(slug: string, kbId: string): string {
  return `/api/org/${slug}/knowledge/${kbId}/hydrate`;
}

export async function fetchOrgKnowledgeBases(
  slug: string,
  init?: RequestInit,
): Promise<KnowledgeBaseListResponse> {
  return fetchJson<KnowledgeBaseListResponse>(knowledgeApiPath(slug), init);
}

export async function fetchOrgKnowledgeDocuments(
  slug: string,
  kbId: string,
  options?: { baseUrl?: string; init?: RequestInit },
): Promise<IngestionResponse> {
  const path = knowledgeDocumentsApiPath(slug, kbId);
  const url = options?.baseUrl ? `${options.baseUrl}${path}` : path;
  return fetchJson<IngestionResponse>(url, options?.init);
}

export async function fetchOrgKnowledgeChunksByDocument(
  slug: string,
  kbId: string,
  options?: { baseUrl?: string; init?: RequestInit },
): Promise<KnowledgeChunksByDocumentResponse> {
  const path = knowledgeChunksApiPath(slug, kbId);
  const url = options?.baseUrl ? `${options.baseUrl}${path}` : path;
  return fetchJson<KnowledgeChunksByDocumentResponse>(url, options?.init);
}

export async function fetchOrgKnowledgeChunkCounts(
  slug: string,
  kbId: string,
  options?: { baseUrl?: string; init?: RequestInit },
): Promise<KnowledgeChunkCountsResponse> {
  const path = knowledgeChunkCountsApiPath(slug, kbId);
  const url = options?.baseUrl ? `${options.baseUrl}${path}` : path;
  return fetchJson<KnowledgeChunkCountsResponse>(url, options?.init);
}

export async function fetchOrgKnowledgeChunksForDocument(
  slug: string,
  kbId: string,
  documentId: string,
  options?: { baseUrl?: string; init?: RequestInit },
): Promise<KnowledgeDocumentChunksResponse> {
  const path = knowledgeDocumentChunksApiPath(slug, kbId, documentId);
  const url = options?.baseUrl ? `${options.baseUrl}${path}` : path;
  return fetchJson<KnowledgeDocumentChunksResponse>(url, options?.init);
}

export async function embedOrgKnowledgeChunk(
  slug: string,
  kbId: string,
  chunkId: string,
  options?: { baseUrl?: string; init?: RequestInit },
): Promise<KnowledgeChunkEmbedResponse> {
  const path = knowledgeChunkEmbedApiPath(slug, kbId, chunkId);
  const url = options?.baseUrl ? `${options.baseUrl}${path}` : path;
  return fetchJson<KnowledgeChunkEmbedResponse>(url, {
    method: "POST",
    ...(options?.init ?? {}),
  });
}

export async function retrieveOrgKnowledge(
  slug: string,
  kbId: string,
  payload: KnowledgeRetrievalRequest,
  options?: { baseUrl?: string; init?: RequestInit },
): Promise<KnowledgeRetrievalResponse> {
  const path = knowledgeRetrieveApiPath(slug, kbId);
  const url = options?.baseUrl ? `${options.baseUrl}${path}` : path;
  return fetchJson<KnowledgeRetrievalResponse>(url, {
    method: "POST",
    body: JSON.stringify(payload),
    ...(options?.init ?? {}),
  });
}

export async function hydrateOrgKnowledge(
  slug: string,
  kbId: string,
  payload: KnowledgeHydrateRequest,
  options?: { baseUrl?: string; init?: RequestInit },
): Promise<KnowledgeHydrateResponse> {
  const path = knowledgeHydrateApiPath(slug, kbId);
  const url = options?.baseUrl ? `${options.baseUrl}${path}` : path;
  return fetchJson<KnowledgeHydrateResponse>(url, {
    method: "POST",
    body: JSON.stringify(payload),
    ...(options?.init ?? {}),
  });
}
