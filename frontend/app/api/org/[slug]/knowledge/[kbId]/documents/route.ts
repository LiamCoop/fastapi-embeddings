import { loadOrgKnowledgeDocuments } from "@/app/lib/org-knowledge.server";
import { uploadFileToStorage } from "@/app/lib/storage.server";
import { getKnowledgeDal } from "@/lib/dal";
import {
  knowledgeKbChunkCountsTag,
  knowledgeKbDocumentsTag,
  knowledgeKbChunksTag,
  knowledgeKbTag,
  knowledgeOrgTag,
} from "@/app/lib/knowledge-cache";
import { revalidateTag } from "next/cache";
import path from "node:path";

export async function GET(
  _request: Request,
  { params }: { params: Promise<{ slug: string; kbId: string }> },
) {
  const { slug, kbId } = await params;
  if (!slug || !kbId) {
    return Response.json({ error: "Organization slug and kb id are required" }, { status: 400 });
  }

  try {
    const data = await loadOrgKnowledgeDocuments(slug, kbId);
    if (!data) {
      return Response.json({ error: "Knowledge base not found" }, { status: 404 });
    }
    return Response.json(data);
  } catch (error) {
    console.error("Failed to load knowledge base documents", {
      slug,
      kbId,
      error,
    });
    return Response.json({ error: "Failed to load knowledge base documents" }, { status: 500 });
  }
}

function inferDocumentType(filename: string, contentType: string): string {
  if (contentType?.trim()) {
    return contentType;
  }

  const ext = path.extname(filename).toLowerCase();
  if (ext === ".md" || ext === ".mdx") {
    return "text/markdown";
  }
  if (ext === ".pdf") {
    return "application/pdf";
  }
  if (ext === ".doc") {
    return "application/msword";
  }
  if (ext === ".docx") {
    return "application/vnd.openxmlformats-officedocument.wordprocessingml.document";
  }
  if (ext === ".txt") {
    return "text/plain";
  }

  return "application/octet-stream";
}

function normalizePath(value: string): string {
  return value.replaceAll("\\", "/").trim();
}

export async function POST(
  request: Request,
  { params }: { params: Promise<{ slug: string; kbId: string }> },
) {
  const { slug, kbId } = await params;
  if (!slug || !kbId) {
    return Response.json({ error: "Organization slug and kb id are required" }, { status: 400 });
  }

  try {
    const form = await request.formData();
    const file = form.get("file");
    if (!(file instanceof File)) {
      return Response.json({ error: "file is required" }, { status: 400 });
    }

    const providedPath = typeof form.get("path") === "string" ? (form.get("path") as string) : "";
    const documentPath = normalizePath(providedPath || file.name);
    if (!documentPath) {
      return Response.json({ error: "path cannot be empty" }, { status: 400 });
    }

    const providedTitle = form.get("title");
    const documentTitle =
      typeof providedTitle === "string" && providedTitle.trim().length > 0 ? providedTitle.trim() : null;

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

    const existingDocumentId = await dal.getDocumentIdByKbPath(kbId, documentPath);
    const documentId = existingDocumentId ?? crypto.randomUUID();

    const filename = path.posix.basename(documentPath) || file.name;
    const key = `kb/${kbId}/docs/${documentId}/${filename}`;
    const rawBytes = Buffer.from(await file.arrayBuffer());
    const documentType = inferDocumentType(filename, file.type);
    const rawContentUri = await uploadFileToStorage(key, rawBytes, documentType);

    const created = await dal.createOrUpdateDocumentAndInsertVersion({
      kbId,
      documentPath,
      documentTitle,
      documentType,
      originalFilename: file.name,
      sizeBytes: file.size,
      rawContentUri,
      documentId,
    });

    revalidateTag(knowledgeOrgTag(slug), "max");
    revalidateTag(knowledgeKbTag(kbId), "max");
    revalidateTag(knowledgeKbDocumentsTag(kbId), "max");
    revalidateTag(knowledgeKbChunksTag(kbId), "max");
    revalidateTag(knowledgeKbChunkCountsTag(kbId), "max");

    return Response.json(
      {
        id: created.id,
        path: created.path,
        title: created.title,
        document_type: created.documentType,
        source_metadata: created.sourceMetadata,
        active_version_id: created.activeVersionId,
        processing_status: created.processingStatus,
        version_number: created.versionNumber,
        created_at: created.createdAt.toISOString(),
        updated_at: created.updatedAt.toISOString(),
      },
      { status: 201 },
    );
  } catch (error) {
    console.error("Failed to upload document", {
      slug,
      kbId,
      error,
    });
    return Response.json({ error: "Failed to upload document" }, { status: 500 });
  }
}
