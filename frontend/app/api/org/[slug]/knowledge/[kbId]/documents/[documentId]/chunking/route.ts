import { apiBaseUrl } from "@/app/lib/api";
import {
  knowledgeKbChunkCountsTag,
  knowledgeKbDocumentChunksTag,
  knowledgeKbDocumentsTag,
  knowledgeKbChunksTag,
  knowledgeKbTag,
} from "@/app/lib/knowledge-cache";
import { revalidateTag } from "next/cache";

export async function POST(
  request: Request,
  { params }: { params: Promise<{ slug: string; kbId: string; documentId: string }> },
) {
  const { kbId, documentId } = await params;
  if (!kbId || !documentId) {
    return Response.json({ error: "kb id and document id are required" }, { status: 400 });
  }

  const payload = (await request.json().catch(() => null)) as {
    strategy?: string;
    max_runes?: number;
    overlap_runes?: number;
    separators?: string[];
    language_hints?: string[];
  } | null;

  const strategy = payload?.strategy?.trim() || "fixed";
  const upstreamRes = await fetch(`${apiBaseUrl()}/kb/${kbId}/documents/${documentId}/chunking`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({
      strategy,
      max_runes: payload?.max_runes ?? 0,
      overlap_runes: payload?.overlap_runes ?? 0,
      separators: payload?.separators ?? [],
      language_hints: payload?.language_hints ?? [],
    }),
    cache: "no-store",
  });

  const body = await upstreamRes.text();
  if (!upstreamRes.ok) {
    return Response.json(
      { error: body || `Chunking request failed with status ${upstreamRes.status}` },
      { status: upstreamRes.status },
    );
  }

  if (!body) {
    revalidateTag(knowledgeKbTag(kbId), "max");
    revalidateTag(knowledgeKbDocumentsTag(kbId), "max");
    revalidateTag(knowledgeKbChunksTag(kbId), "max");
    revalidateTag(knowledgeKbChunkCountsTag(kbId), "max");
    revalidateTag(knowledgeKbDocumentChunksTag(kbId, documentId), "max");
    return new Response(null, { status: upstreamRes.status });
  }

  revalidateTag(knowledgeKbTag(kbId), "max");
  revalidateTag(knowledgeKbDocumentsTag(kbId), "max");
  revalidateTag(knowledgeKbChunksTag(kbId), "max");
  revalidateTag(knowledgeKbChunkCountsTag(kbId), "max");
  revalidateTag(knowledgeKbDocumentChunksTag(kbId, documentId), "max");

  return new Response(body, {
    status: upstreamRes.status,
    headers: {
      "Content-Type": upstreamRes.headers.get("content-type") ?? "application/json",
    },
  });
}
