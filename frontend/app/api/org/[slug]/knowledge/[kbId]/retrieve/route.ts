import { apiBaseUrl } from "@/app/lib/api";

type RetrievalPayload = {
  query?: string;
  top_k?: number;
  retrieval_profile?: "auto" | "exact" | "balanced" | "semantic";
  semantic_weight?: number;
  hybrid_weight?: number;
  debug?: boolean;
  filters?: {
    project_id?: string;
    path_prefix?: string;
    document_type?: string;
    source?: string;
    tags?: string[];
    updated_after?: string;
    created_after?: string;
    created_before?: string;
  };
};

export async function POST(
  request: Request,
  { params }: { params: Promise<{ slug: string; kbId: string }> },
) {
  const { kbId } = await params;
  if (!kbId) {
    return Response.json({ error: "kb id is required" }, { status: 400 });
  }

  const payload = (await request.json().catch(() => null)) as RetrievalPayload | null;
  const query = payload?.query?.trim() ?? "";
  if (!query) {
    return Response.json({ error: "query is required" }, { status: 400 });
  }

  const upstreamRes = await fetch(`${apiBaseUrl()}/kb/${kbId}/query`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({
      query,
      top_k: payload?.top_k,
      retrieval_profile: payload?.retrieval_profile,
      semantic_weight: payload?.semantic_weight,
      hybrid_weight: payload?.hybrid_weight,
      debug: payload?.debug,
      filters: payload?.filters,
    }),
    cache: "no-store",
  });

  const body = await upstreamRes.text();
  if (!upstreamRes.ok) {
    return Response.json(
      { error: body || `Retrieval request failed with status ${upstreamRes.status}` },
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
