import { loadOrgKnowledgeChunksByDocument } from "@/app/lib/org-knowledge.server";

export async function GET(
  _request: Request,
  { params }: { params: Promise<{ slug: string; kbId: string }> },
) {
  const { slug, kbId } = await params;
  if (!slug || !kbId) {
    return Response.json({ error: "Organization slug and kb id are required" }, { status: 400 });
  }

  try {
    const data = await loadOrgKnowledgeChunksByDocument(slug, kbId);
    if (!data) {
      return Response.json({ error: "Knowledge base not found" }, { status: 404 });
    }
    return Response.json(data);
  } catch (error) {
    console.error("Failed to load knowledge base chunks", {
      slug,
      kbId,
      error,
    });
    return Response.json({ error: "Failed to load knowledge base chunks" }, { status: 500 });
  }
}
