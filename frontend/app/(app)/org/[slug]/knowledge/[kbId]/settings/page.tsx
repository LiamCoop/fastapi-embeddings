export default async function KbSettingsPage({
  params,
}: {
  params: Promise<{ slug: string; kbId: string }>;
}) {
  const { kbId } = await params;

  return (
    <div className="px-8 py-10 space-y-8">
      <div>
        <p className="text-xs uppercase tracking-[0.2em] text-muted-foreground">Settings</p>
        <h1 className="mt-1 text-2xl font-semibold text-foreground">Corpus Settings</h1>
        <p className="mt-0.5 font-mono text-xs text-muted-foreground/60">{kbId}</p>
      </div>

      <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
        <div className="rounded-xl border border-border bg-card p-6">
          <p className="text-xs uppercase tracking-[0.2em] text-muted-foreground">Display Name</p>
          <div className="mt-4 rounded-lg border border-dashed border-border py-8 text-center">
            <p className="text-sm text-muted-foreground/60">Not configured</p>
          </div>
        </div>

        <div className="rounded-xl border border-border bg-card p-6">
          <p className="text-xs uppercase tracking-[0.2em] text-muted-foreground">Visibility</p>
          <div className="mt-4 rounded-lg border border-dashed border-border py-8 text-center">
            <p className="text-sm text-muted-foreground/60">Visible</p>
          </div>
        </div>

        <div className="rounded-xl border border-border bg-card p-6 md:col-span-2">
          <p className="text-xs uppercase tracking-[0.2em] text-muted-foreground">Danger Zone</p>
          <div className="mt-4 rounded-lg border border-dashed border-destructive/40 py-8 text-center">
            <p className="text-sm text-muted-foreground/60">Delete corpus</p>
            <p className="mt-1 text-xs text-muted-foreground/40">
              This action cannot be undone.
            </p>
          </div>
        </div>
      </div>
    </div>
  );
}
