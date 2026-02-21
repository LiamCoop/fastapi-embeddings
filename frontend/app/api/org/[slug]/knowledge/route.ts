import { getKnowledgeDal } from "@/lib/dal";
import { KNOWLEDGE_CACHE_REVALIDATE_SECONDS, knowledgeOrgTag } from "@/app/lib/knowledge-cache";
import { revalidateTag, unstable_cache } from "next/cache";

export async function GET(
  _request: Request,
  { params }: { params: Promise<{ slug: string }> },
) {
  const { slug } = await params;
  if (!slug) {
    return Response.json({ error: "Organization slug is required" }, { status: 400 });
  }

  try {
    const items = await unstable_cache(
      async () => {
        const dal = getKnowledgeDal();

        const org = await dal.getOrganizationBySlug(slug);

        if (!org) {
          return null;
        }

        const entities = await dal.listKnowledgeBasesForOrg(org.id, org.slug);

        return entities.map((kb) => ({
          id: kb.id,
          name: kb.name,
          metadata: kb.metadata ?? {},
          created_at: kb.createdAt,
          updated_at: kb.updatedAt,
        }));
      },
      ["org-knowledge-bases", "v1", slug],
      { revalidate: KNOWLEDGE_CACHE_REVALIDATE_SECONDS, tags: [knowledgeOrgTag(slug)] },
    )();

    if (!items) {
      return Response.json({ error: "Organization not found" }, { status: 404 });
    }

    return Response.json({ knowledge_bases: items });
  } catch (error) {
    console.error("Failed to load organization knowledge bases", {
      slug,
      error,
    });
    return Response.json({ error: "Failed to load knowledge bases" }, { status: 500 });
  }
}

export async function POST(
  request: Request,
  { params }: { params: Promise<{ slug: string }> },
) {
  const { slug } = await params;
  if (!slug) {
    return Response.json({ error: "Organization slug is required" }, { status: 400 });
  }

  let payload: { name?: string; metadata?: Record<string, unknown> };
  try {
    payload = (await request.json()) as {
      name?: string;
      metadata?: Record<string, unknown>;
    };
  } catch {
    return Response.json({ error: "Invalid JSON" }, { status: 400 });
  }

  const name = payload.name?.trim();
  if (!name) {
    return Response.json({ error: "Name is required" }, { status: 400 });
  }

  try {
    const dal = getKnowledgeDal();
    const org = await dal.getOrganizationBySlug(slug);

    if (!org) {
      return Response.json({ error: "Organization not found" }, { status: 404 });
    }

    const saved = await dal.createKnowledgeBase(name, {
      ...(payload.metadata ?? {}),
      org_slug: org.slug,
      organization_id: org.id,
    });
    revalidateTag(knowledgeOrgTag(slug), "max");

    return Response.json({
      id: saved.id,
      name: saved.name,
      metadata: saved.metadata ?? {},
      created_at: saved.createdAt,
      updated_at: saved.updatedAt,
    });
  } catch (error) {
    console.error("Failed to create organization knowledge base", {
      slug,
      error,
    });
    return Response.json({ error: "Failed to create knowledge base" }, { status: 500 });
  }
}
