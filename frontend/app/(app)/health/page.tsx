export default function HealthPage() {
  return (
    <div className="px-8 py-10 space-y-8">
      <div>
        <p className="text-xs uppercase tracking-[0.2em] text-muted-foreground">Observability</p>
        <h1 className="mt-1 text-2xl font-semibold text-foreground">AI Health</h1>
        <p className="mt-2 text-sm text-muted-foreground">
          Production monitoring: query patterns, failed retrievals, latency, and drift indicators.
        </p>
      </div>

      <div className="grid grid-cols-2 gap-4 md:grid-cols-4">
        {[
          { label: "Total Queries", value: "—" },
          { label: "Failed Retrievals", value: "—" },
          { label: "Avg Latency", value: "—" },
          { label: "Unanswered", value: "—" },
        ].map((stat) => (
          <div key={stat.label} className="rounded-xl border border-border bg-card p-5">
            <p className="text-xs uppercase tracking-[0.2em] text-muted-foreground">{stat.label}</p>
            <p className="mt-2 font-mono text-2xl font-bold text-foreground/70">{stat.value}</p>
          </div>
        ))}
      </div>

      <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
        <div className="rounded-xl border border-border bg-card p-6">
          <p className="text-xs uppercase tracking-[0.2em] text-muted-foreground">Common Queries</p>
          <div className="mt-4 rounded-lg border border-dashed border-border py-12 text-center">
            <p className="text-sm text-muted-foreground/60">No query data yet</p>
          </div>
        </div>

        <div className="rounded-xl border border-border bg-card p-6">
          <p className="text-xs uppercase tracking-[0.2em] text-muted-foreground">Drift Indicators</p>
          <div className="mt-4 rounded-lg border border-dashed border-border py-12 text-center">
            <p className="text-sm text-muted-foreground/60">No drift detected</p>
          </div>
        </div>
      </div>
    </div>
  );
}
