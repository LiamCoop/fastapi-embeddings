import { deleteFileFromStorageUri } from "@/app/lib/storage.server";
import {
  knowledgeKbChunkCountsTag,
  knowledgeKbDocumentsTag,
  knowledgeKbChunksTag,
  knowledgeKbTag,
  knowledgeOrgTag,
} from "@/app/lib/knowledge-cache";
import { prisma } from "@/lib/prisma";
import { revalidateTag } from "next/cache";

export async function DELETE(
  _request: Request,
  { params }: { params: Promise<{ slug: string; kbId: string }> },
) {
  const { slug, kbId } = await params;
  if (!slug || !kbId) {
    return Response.json({ error: "Organization slug and kb id are required" }, { status: 400 });
  }

  try {
    const org = await prisma.organization.findUnique({
      where: { slug },
      select: { id: true, slug: true },
    });
    if (!org) {
      return Response.json({ error: "Organization not found" }, { status: 404 });
    }

    const kb = await prisma.knowledgeBase.findUnique({
      where: { id: kbId },
      select: { metadata: true },
    });
    if (!kb) {
      return Response.json({ error: "Knowledge base not found" }, { status: 404 });
    }

    const metadata =
      kb.metadata && typeof kb.metadata === "object" && !Array.isArray(kb.metadata)
        ? (kb.metadata as Record<string, unknown>)
        : {};
    const kbOrgId = metadata.organization_id;
    const kbOrgSlug = metadata.org_slug;
    if (kbOrgId !== org.id && kbOrgSlug !== org.slug) {
      return Response.json({ error: "Knowledge base not found" }, { status: 404 });
    }

    const rawContentUris = (
      await prisma.documentVersion.findMany({
        where: { kbId },
        distinct: ["rawContentUri"],
        select: { rawContentUri: true },
      })
    ).map((row) => row.rawContentUri);
    const uniqueUris = [...new Set(rawContentUris)].filter((value) => value.startsWith("s3://"));

    const deleteResults = await Promise.allSettled(uniqueUris.map((uri) => deleteFileFromStorageUri(uri)));
    const failedDeletes = deleteResults.filter((result) => result.status === "rejected");
    if (failedDeletes.length > 0) {
      throw new Error(`Failed to delete ${failedDeletes.length} object(s) from storage`);
    }

    const deleted = (await prisma.knowledgeBase.deleteMany({ where: { id: kbId } })).count > 0;
    if (!deleted) {
      return Response.json({ error: "Knowledge base not found" }, { status: 404 });
    }

    revalidateTag(knowledgeOrgTag(slug), "max");
    revalidateTag(knowledgeKbTag(kbId), "max");
    revalidateTag(knowledgeKbDocumentsTag(kbId), "max");
    revalidateTag(knowledgeKbChunksTag(kbId), "max");
    revalidateTag(knowledgeKbChunkCountsTag(kbId), "max");

    return Response.json({ id: kbId, deleted: true, deleted_objects: uniqueUris.length });
  } catch (error) {
    console.error("Failed to delete knowledge base", {
      slug,
      kbId,
      error,
    });
    return Response.json({ error: "Failed to delete knowledge base" }, { status: 500 });
  }
}
