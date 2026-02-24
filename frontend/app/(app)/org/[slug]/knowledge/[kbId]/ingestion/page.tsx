import { loadOrgKnowledgeDocuments } from "@/app/lib/org-knowledge.server";
import { formatUtcDateTime } from "@/app/lib/datetime";
import { DocumentsList } from "./_components/DocumentsList";
import { UploadZone } from "./_components/UploadZone";

export default async function KbIngestionPage({
  params,
}: {
  params: Promise<{ slug: string; kbId: string }>;
}) {
  const { slug, kbId } = await params;

  let data = null;
  let error: string | null = null;

  try {
    data = await loadOrgKnowledgeDocuments(slug, kbId);
  } catch (err) {
    error = err instanceof Error ? err.message : "Failed to load ingestion data";
  }

  if (error || !data) {
    return <div className="px-8 py-10 text-sm text-destructive">{error ?? "Corpus not found."}</div>;
  }

  const kb = data.knowledge_base;
  const documents = data.documents ?? [];

  return (
    <div className="px-8 py-8 space-y-8">
      <div className="flex items-start justify-between gap-4">
        <div>
          <p className="text-xs uppercase tracking-[0.2em] text-muted-foreground">Ingestion</p>
          <h2 className="mt-1 text-lg font-semibold text-foreground">{kb.name}</h2>
          <p className="mt-0.5 text-xs text-muted-foreground">
            Last updated {formatUtcDateTime(kb.updated_at)}
          </p>
        </div>
      </div>

      <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
        <div className="rounded-xl border border-border bg-card p-6">
          <UploadZone slug={slug} kbId={kbId} />
        </div>

        <div className="rounded-xl border border-border bg-card p-6">
          <p className="text-xs uppercase tracking-[0.2em] text-muted-foreground">Sync Configuration</p>
          <div className="mt-4 rounded-lg border border-dashed border-border py-10 text-center">
            <p className="text-sm text-muted-foreground/60">No sync schedules configured</p>
          </div>
        </div>

        <div className="rounded-xl border border-border bg-card p-6 md:col-span-2">
          <div className="flex items-center justify-between gap-3">
            <p className="text-xs uppercase tracking-[0.2em] text-muted-foreground">Documents</p>
            <p className="font-mono text-xs text-muted-foreground">{documents.length}</p>
          </div>

          <DocumentsList slug={slug} kbId={kbId} documents={documents} />
        </div>
      </div>
    </div>
  );
}
