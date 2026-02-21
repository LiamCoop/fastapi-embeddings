import { Button } from "@/components/ui/button";

export default function KbPlaygroundPage() {
  return (
    <div className="flex h-full flex-col px-8 py-8 gap-6">
      <div>
        <p className="text-xs uppercase tracking-[0.2em] text-muted-foreground">Playground</p>
        <h1 className="mt-1 text-2xl font-semibold text-foreground">Answer Playground</h1>
        <p className="mt-2 text-sm text-muted-foreground">
          Ask questions, inspect retrieved chunks, trace grounding, and compare configurations.
        </p>
      </div>

      <div className="flex flex-1 gap-6 min-h-0">
        <div className="flex flex-1 flex-col gap-4">
          <div className="rounded-xl border border-border bg-card p-5">
            <p className="text-xs uppercase tracking-[0.2em] text-muted-foreground">Query</p>
            <div className="mt-3 flex gap-3">
              <div className="flex-1 rounded-lg border border-dashed border-border px-4 py-3 text-sm text-muted-foreground/60">
                Ask a question...
              </div>
              <Button disabled>Ask</Button>
            </div>
          </div>

          <div className="flex-1 rounded-xl border border-border bg-card p-6">
            <p className="text-xs uppercase tracking-[0.2em] text-muted-foreground">Answer</p>
            <div className="mt-4 rounded-lg border border-dashed border-border py-16 text-center">
              <p className="text-sm text-muted-foreground/60">Answer will appear here</p>
              <p className="mt-1 text-xs text-muted-foreground/40">
                With grounding and chunk citations
              </p>
            </div>
          </div>
        </div>

        <div className="w-80 shrink-0 space-y-4">
          <div className="rounded-xl border border-border bg-card p-5">
            <p className="text-xs uppercase tracking-[0.2em] text-muted-foreground">Configuration</p>
            <div className="mt-3 space-y-2">
              <div className="rounded-md border border-dashed border-border px-3 py-2 text-xs text-muted-foreground/60">
                Top-K: 5
              </div>
              <div className="rounded-md border border-dashed border-border px-3 py-2 text-xs text-muted-foreground/60">
                Hybrid weight: 0.7
              </div>
            </div>
          </div>

          <div className="rounded-xl border border-border bg-card p-5">
            <p className="text-xs uppercase tracking-[0.2em] text-muted-foreground">
              Retrieved Chunks
            </p>
            <div className="mt-3 rounded-lg border border-dashed border-border py-12 text-center">
              <p className="text-xs text-muted-foreground/60">No chunks retrieved</p>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
