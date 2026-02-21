import { apiBaseUrl } from "@/app/lib/api";
import {
  knowledgeKbChunkCountsTag,
  knowledgeKbDocumentsTag,
  knowledgeKbChunksTag,
  knowledgeKbTag,
} from "@/app/lib/knowledge-cache";
import { revalidateTag } from "next/cache";

export async function POST(
  _request: Request,
  { params }: { params: Promise<{ slug: string; kbId: string; chunkId: string }> },
) {
  const { kbId, chunkId } = await params;
  if (!kbId || !chunkId) {
    return Response.json({ error: "kb id and chunk id are required" }, { status: 400 });
  }

  const upstreamRes = await fetch(`${apiBaseUrl()}/kb/${kbId}/chunks/${chunkId}/embed`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    cache: "no-store",
  });

  const body = await upstreamRes.text();
  if (!upstreamRes.ok) {
    return Response.json(
      { error: body || `Chunk embed request failed with status ${upstreamRes.status}` },
      { status: upstreamRes.status },
    );
  }

  if (!body) {
    revalidateTag(knowledgeKbTag(kbId), "max");
    revalidateTag(knowledgeKbDocumentsTag(kbId), "max");
    revalidateTag(knowledgeKbChunksTag(kbId), "max");
    revalidateTag(knowledgeKbChunkCountsTag(kbId), "max");
    return new Response(null, { status: upstreamRes.status });
  }

  revalidateTag(knowledgeKbTag(kbId), "max");
  revalidateTag(knowledgeKbDocumentsTag(kbId), "max");
  revalidateTag(knowledgeKbChunksTag(kbId), "max");
  revalidateTag(knowledgeKbChunkCountsTag(kbId), "max");

  return new Response(body, {
    status: upstreamRes.status,
    headers: {
      "Content-Type": upstreamRes.headers.get("content-type") ?? "application/json",
    },
  });
}
