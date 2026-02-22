import { apiBaseUrl } from "@/app/lib/api";

type HydratePayload = {
  chunk_ids?: string[];
  adjacent_before?: number;
  adjacent_after?: number;
};

export async function POST(
  request: Request,
  { params }: { params: Promise<{ slug: string; kbId: string }> },
) {
  const { kbId } = await params;
  if (!kbId) {
    return Response.json({ error: "kb id is required" }, { status: 400 });
  }

  const payload = (await request.json().catch(() => null)) as HydratePayload | null;
  const chunkIDs = payload?.chunk_ids?.filter((id) => id.trim().length > 0) ?? [];
  if (chunkIDs.length === 0) {
    return Response.json({ error: "chunk_ids is required" }, { status: 400 });
  }

  const upstreamRes = await fetch(`${apiBaseUrl()}/kb/${kbId}/hydrate`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({
      chunk_ids: chunkIDs,
      adjacent_before: payload?.adjacent_before ?? 0,
      adjacent_after: payload?.adjacent_after ?? 0,
    }),
    cache: "no-store",
  });

  const body = await upstreamRes.text();
  if (!upstreamRes.ok) {
    return Response.json(
      { error: body || `Hydrate request failed with status ${upstreamRes.status}` },
      { status: upstreamRes.status },
    );
  }

  if (!body) {
    return new Response(null, { status: upstreamRes.status });
  }

  return new Response(body, {
    status: upstreamRes.status,
    headers: {
      "Content-Type": upstreamRes.headers.get("content-type") ?? "application/json",
    },
  });
}
