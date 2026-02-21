import { RetrievalQueryClient } from "./_components/RetrievalQueryClient";

export default async function KbRetrievalPage({
  params,
}: {
  params: Promise<{ slug: string; kbId: string }>;
}) {
  const { slug, kbId } = await params;

  return (
    <div className="px-8 py-8 space-y-8">
      <div>
        <p className="text-xs uppercase tracking-[0.2em] text-muted-foreground">Retrieval</p>
        <p className="mt-2 text-sm text-muted-foreground">
          Configure embedding models, hybrid retrieval weights, metadata indexing, and reranker
          settings.
        </p>
      </div>

      <RetrievalQueryClient slug={slug} kbId={kbId} />

      <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
        <div className="rounded-xl border border-border bg-card p-6">
          <p className="text-xs uppercase tracking-[0.2em] text-muted-foreground">Embedding Model</p>
          <div className="mt-4 rounded-lg border border-dashed border-border py-10 text-center">
            <p className="text-sm text-muted-foreground/60">text-embedding-3-small</p>
            <p className="mt-1 text-xs text-muted-foreground/40">Global default</p>
          </div>
        </div>

        <div className="rounded-xl border border-border bg-card p-6">
          <p className="text-xs uppercase tracking-[0.2em] text-muted-foreground">
            Hybrid Retrieval
          </p>
          <div className="mt-4 rounded-lg border border-dashed border-border py-10 text-center">
            <p className="text-sm text-muted-foreground/60">Semantic + Lexical blend</p>
            <p className="mt-1 text-xs text-muted-foreground/40">Configurable weight</p>
          </div>
        </div>

        <div className="rounded-xl border border-border bg-card p-6">
          <p className="text-xs uppercase tracking-[0.2em] text-muted-foreground">
            Metadata Indexing
          </p>
          <div className="mt-4 rounded-lg border border-dashed border-border py-10 text-center">
            <p className="text-sm text-muted-foreground/60">No metadata filters configured</p>
          </div>
        </div>

        <div className="rounded-xl border border-border bg-card p-6">
          <p className="text-xs uppercase tracking-[0.2em] text-muted-foreground">Reranker</p>
          <div className="mt-4 rounded-lg border border-dashed border-border py-10 text-center">
            <p className="text-sm text-muted-foreground/60">Not configured</p>
          </div>
        </div>
      </div>
    </div>
  );
}
