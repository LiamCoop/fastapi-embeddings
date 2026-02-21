import {
  fetchOrgKnowledgeChunkCounts,
  fetchOrgKnowledgeDocuments,
} from "@/app/lib/org-knowledge";
import {
  KNOWLEDGE_CACHE_REVALIDATE_SECONDS,
  knowledgeKbChunkCountsTag,
  knowledgeKbDocumentsTag,
} from "@/app/lib/knowledge-cache";
import { ChunkExplorerClient } from "./_components/ChunkExplorerClient";
import { headers } from "next/headers";

async function getRequestContext() {
  const headerList = await headers();
  const host = headerList.get("x-forwarded-host") ?? headerList.get("host");

  if (!host) {
    throw new Error("Unable to determine request host");
  }

  const proto =
    headerList.get("x-forwarded-proto") ?? (host.includes("localhost") || host.startsWith("127.0.0.1") ? "http" : "https");

  return {
    baseUrl: `${proto}://${host}`,
    cookie: headerList.get("cookie") ?? "",
  };
}

export default async function KbChunksPage({
  params,
}: {
  params: Promise<{ slug: string; kbId: string }>;
}) {
  const { slug, kbId } = await params;

  let data = null;
  let chunkCountsByDocumentId: Record<string, number> = {};
  let error: string | null = null;

  try {
    const { baseUrl, cookie } = await getRequestContext();
    const [documentsRes, countsRes] = await Promise.all([
      fetchOrgKnowledgeDocuments(slug, kbId, {
        baseUrl,
        init: {
          cache: "force-cache",
          headers: { cookie },
          next: {
            revalidate: KNOWLEDGE_CACHE_REVALIDATE_SECONDS,
            tags: [knowledgeKbDocumentsTag(kbId)],
          },
        },
      }),
      fetchOrgKnowledgeChunkCounts(slug, kbId, {
        baseUrl,
        init: {
          cache: "force-cache",
          headers: { cookie },
          next: {
            revalidate: KNOWLEDGE_CACHE_REVALIDATE_SECONDS,
            tags: [knowledgeKbChunkCountsTag(kbId)],
          },
        },
      }),
    ]);
    data = documentsRes;
    chunkCountsByDocumentId = countsRes.chunk_counts_by_document_id;
  } catch (err) {
    error = err instanceof Error ? err.message : "Failed to load documents";
  }

  if (error || !data) {
    return <div className="px-8 py-10 text-sm text-destructive">{error ?? "Corpus not found."}</div>;
  }

  const documents = data.documents ?? [];
  return (
    <div className="px-8 py-8 space-y-8">
      <div>
        <p className="text-xs uppercase tracking-[0.2em] text-muted-foreground">Chunk Explorer</p>
        <p className="mt-2 text-sm text-muted-foreground">
          Browse documents, inspect chunks, and view metadata. Diff chunks across versions.
        </p>
      </div>

      <ChunkExplorerClient
        slug={slug}
        kbId={kbId}
        documents={documents}
        chunkCountsByDocumentId={chunkCountsByDocumentId}
      />
    </div>
  );
}
