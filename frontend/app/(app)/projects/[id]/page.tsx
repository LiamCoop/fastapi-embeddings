export default async function ProjectDetailPage({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const { id } = await params;

  return (
    <div className="px-8 py-10 space-y-8">
      <div>
        <p className="text-xs uppercase tracking-[0.2em] text-muted-foreground">Project</p>
        <h1 className="mt-1 text-2xl font-semibold text-foreground">Project Overview</h1>
        <p className="mt-1 font-mono text-xs text-muted-foreground/60">{id}</p>
      </div>

      <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
        <div className="rounded-xl border border-border bg-card p-6">
          <p className="text-xs uppercase tracking-[0.2em] text-muted-foreground">Corpuses</p>
          <div className="mt-4 rounded-lg border border-dashed border-border py-10 text-center">
            <p className="text-sm text-muted-foreground/60">No corpuses in this project</p>
          </div>
        </div>

        <div className="rounded-xl border border-border bg-card p-6">
          <p className="text-xs uppercase tracking-[0.2em] text-muted-foreground">Evaluation Summary</p>
          <div className="mt-4 rounded-lg border border-dashed border-border py-10 text-center">
            <p className="text-sm text-muted-foreground/60">No evaluations run yet</p>
          </div>
        </div>

        <div className="rounded-xl border border-border bg-card p-6">
          <p className="text-xs uppercase tracking-[0.2em] text-muted-foreground">MCP Integrations</p>
          <div className="mt-4 rounded-lg border border-dashed border-border py-10 text-center">
            <p className="text-sm text-muted-foreground/60">No integrations connected</p>
          </div>
        </div>

        <div className="rounded-xl border border-border bg-card p-6">
          <p className="text-xs uppercase tracking-[0.2em] text-muted-foreground">Query Analytics</p>
          <div className="mt-4 rounded-lg border border-dashed border-border py-10 text-center">
            <p className="text-sm text-muted-foreground/60">No queries recorded</p>
          </div>
        </div>
      </div>
    </div>
  );
}
