import { loadOrgKnowledgeChunksForDocument } from "@/app/lib/org-knowledge.server";

export async function GET(
  _request: Request,
  { params }: { params: Promise<{ slug: string; kbId: string; documentId: string }> },
) {
  const { slug, kbId, documentId } = await params;
  if (!slug || !kbId || !documentId) {
    return Response.json({ error: "Organization slug, kb id, and document id are required" }, { status: 400 });
  }

  try {
    const data = await loadOrgKnowledgeChunksForDocument(slug, kbId, documentId);
    if (!data) {
      return Response.json({ error: "Document not found" }, { status: 404 });
    }
    return Response.json(data);
  } catch (error) {
    console.error("Failed to load document chunks", {
      slug,
      kbId,
      documentId,
      error,
    });
    return Response.json({ error: "Failed to load document chunks" }, { status: 500 });
  }
}
