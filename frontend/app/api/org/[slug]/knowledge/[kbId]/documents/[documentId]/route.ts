import { getKnowledgeDal } from "@/lib/dal";
import {
  knowledgeKbChunkCountsTag,
  knowledgeKbDocumentChunksTag,
  knowledgeKbDocumentsTag,
  knowledgeKbChunksTag,
  knowledgeKbTag,
  knowledgeOrgTag,
} from "@/app/lib/knowledge-cache";
import { revalidateTag } from "next/cache";

export async function DELETE(
  _request: Request,
  { params }: { params: Promise<{ slug: string; kbId: string; documentId: string }> },
) {
  const { slug, kbId, documentId } = await params;
  if (!slug || !kbId || !documentId) {
    return Response.json({ error: "Organization slug, kb id, and document id are required" }, { status: 400 });
  }

  try {
    const dal = getKnowledgeDal();
    const org = await dal.getOrganizationBySlug(slug);
    if (!org) {
      return Response.json({ error: "Organization not found" }, { status: 404 });
    }

    const kb = await dal.getKnowledgeBaseById(kbId);
    if (!kb) {
      return Response.json({ error: "Knowledge base not found" }, { status: 404 });
    }

    const kbOrgId = kb.metadata?.organization_id;
    const kbOrgSlug = kb.metadata?.org_slug;
    if (kbOrgId !== org.id && kbOrgSlug !== org.slug) {
      return Response.json({ error: "Knowledge base not found" }, { status: 404 });
    }

    const deleted = await dal.deleteDocumentById(kbId, documentId);
    if (!deleted) {
      return Response.json({ error: "Document not found" }, { status: 404 });
    }

    revalidateTag(knowledgeOrgTag(slug), "max");
    revalidateTag(knowledgeKbTag(kbId), "max");
    revalidateTag(knowledgeKbDocumentsTag(kbId), "max");
    revalidateTag(knowledgeKbChunksTag(kbId), "max");
    revalidateTag(knowledgeKbChunkCountsTag(kbId), "max");
    revalidateTag(knowledgeKbDocumentChunksTag(kbId, documentId), "max");

    return Response.json({ id: documentId, deleted: true });
  } catch (error) {
    console.error("Failed to delete document", {
      slug,
      kbId,
      documentId,
      error,
    });
    return Response.json({ error: "Failed to delete document" }, { status: 500 });
  }
}
