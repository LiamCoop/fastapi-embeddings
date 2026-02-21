import { Button } from "@/components/ui/button";

export default function KbEvaluationPage() {
  return (
    <div className="px-8 py-10 space-y-8">
      <div className="flex items-center justify-between">
        <div>
          <p className="text-xs uppercase tracking-[0.2em] text-muted-foreground">Evaluation</p>
          <h1 className="mt-1 text-2xl font-semibold text-foreground">Evaluation Studio</h1>
          <p className="mt-2 text-sm text-muted-foreground">
            Measure retrieval quality and catch regressions before they reach production.
          </p>
        </div>
        <Button disabled>New Evaluation</Button>
      </div>

      <div className="grid grid-cols-1 gap-4 md:grid-cols-3">
        <div className="rounded-xl border border-border bg-card p-6">
          <p className="text-xs uppercase tracking-[0.2em] text-muted-foreground">
            Golden Question Sets
          </p>
          <div className="mt-4 rounded-lg border border-dashed border-border py-10 text-center">
            <p className="text-sm text-muted-foreground/60">No question sets</p>
          </div>
        </div>

        <div className="rounded-xl border border-border bg-card p-6">
          <p className="text-xs uppercase tracking-[0.2em] text-muted-foreground">
            Retrieval Recall
          </p>
          <div className="mt-4 rounded-lg border border-dashed border-border py-10 text-center">
            <p className="text-sm text-muted-foreground/60">No runs yet</p>
          </div>
        </div>

        <div className="rounded-xl border border-border bg-card p-6">
          <p className="text-xs uppercase tracking-[0.2em] text-muted-foreground">
            Hallucination Indicators
          </p>
          <div className="mt-4 rounded-lg border border-dashed border-border py-10 text-center">
            <p className="text-sm text-muted-foreground/60">No data</p>
          </div>
        </div>
      </div>

      <div className="rounded-xl border border-border bg-card p-6">
        <p className="text-xs uppercase tracking-[0.2em] text-muted-foreground">
          Regression Comparison
        </p>
        <div className="mt-4 rounded-lg border border-dashed border-border py-12 text-center">
          <p className="text-sm text-muted-foreground/60">
            Run evaluations to track quality over time
          </p>
        </div>
      </div>
    </div>
  );
}
