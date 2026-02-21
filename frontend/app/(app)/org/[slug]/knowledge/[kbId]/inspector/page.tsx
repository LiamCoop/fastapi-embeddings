import { Button } from "@/components/ui/button";

export default function KbInspectorPage() {
  return (
    <div className="px-8 py-8 space-y-8">
      <div>
        <p className="text-xs uppercase tracking-[0.2em] text-muted-foreground">Inspector</p>
        <p className="mt-2 text-sm text-muted-foreground">
          Debug retrieval behavior. Understand why chunks were or were not retrieved for a given
          query.
        </p>
      </div>

      <div className="rounded-xl border border-border bg-card p-6">
        <p className="text-xs uppercase tracking-[0.2em] text-muted-foreground">Query</p>
        <div className="mt-3 flex gap-3">
          <div className="flex-1 rounded-lg border border-dashed border-border px-4 py-3 text-sm text-muted-foreground/60">
            Enter a query to inspect retrieval...
          </div>
          <Button size="sm" disabled>
            Inspect
          </Button>
        </div>
      </div>

      <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
        <div className="rounded-xl border border-border bg-card p-6">
          <p className="text-xs uppercase tracking-[0.2em] text-muted-foreground">
            Retrieved Chunks
          </p>
          <p className="mt-1 text-xs text-muted-foreground/60">Similarity scores and ranking</p>
          <div className="mt-4 rounded-lg border border-dashed border-border py-12 text-center">
            <p className="text-sm text-muted-foreground/60">Run a query to see results</p>
          </div>
        </div>

        <div className="rounded-xl border border-border bg-card p-6">
          <p className="text-xs uppercase tracking-[0.2em] text-muted-foreground">
            Ranking Explanation
          </p>
          <p className="mt-1 text-xs text-muted-foreground/60">
            Semantic, lexical, and final scores
          </p>
          <div className="mt-4 rounded-lg border border-dashed border-border py-12 text-center">
            <p className="text-sm text-muted-foreground/60">No results yet</p>
          </div>
        </div>
      </div>
    </div>
  );
}
