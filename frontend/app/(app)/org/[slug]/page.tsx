export default async function OrgOverviewPage({
  params,
}: {
  params: Promise<{ slug: string }>;
}) {
  const { slug } = await params;

  return (
    <div className="px-8 py-10 space-y-8">
      <div>
        <p className="text-xs uppercase tracking-[0.2em] text-muted-foreground">Organization</p>
        <h1 className="mt-1 text-2xl font-semibold text-foreground">{slug}</h1>
        <p className="mt-2 text-sm text-muted-foreground">
          High-level health, recent activity, and quick access to your knowledge bases.
        </p>
      </div>

      <div className="grid grid-cols-1 gap-4 md:grid-cols-3">
        <div className="rounded-xl border border-border bg-card p-6">
          <p className="text-xs uppercase tracking-[0.2em] text-muted-foreground">Recent Corpuses</p>
          <div className="mt-4 rounded-lg border border-dashed border-border py-8 text-center">
            <p className="text-sm text-muted-foreground/60">No recent activity</p>
          </div>
        </div>

        <div className="rounded-xl border border-border bg-card p-6">
          <p className="text-xs uppercase tracking-[0.2em] text-muted-foreground">AI Health</p>
          <div className="mt-4 rounded-lg border border-dashed border-border py-8 text-center">
            <p className="text-sm text-muted-foreground/60">No data yet</p>
          </div>
        </div>

        <div className="rounded-xl border border-border bg-card p-6">
          <p className="text-xs uppercase tracking-[0.2em] text-muted-foreground">Quick Actions</p>
          <div className="mt-4 space-y-2">
            <div className="rounded-md border border-dashed border-border px-4 py-3 text-sm text-muted-foreground/60">
              Upload documents
            </div>
            <div className="rounded-md border border-dashed border-border px-4 py-3 text-sm text-muted-foreground/60">
              Test retrieval
            </div>
            <div className="rounded-md border border-dashed border-border px-4 py-3 text-sm text-muted-foreground/60">
              Connect MCP
            </div>
          </div>
        </div>
      </div>

      <div className="rounded-xl border border-border bg-card p-6">
        <p className="text-xs uppercase tracking-[0.2em] text-muted-foreground">Evaluation Regressions</p>
        <div className="mt-4 rounded-lg border border-dashed border-border py-10 text-center">
          <p className="text-sm text-muted-foreground/60">No regressions detected</p>
        </div>
      </div>
    </div>
  );
}
