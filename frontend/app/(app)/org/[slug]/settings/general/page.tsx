export default async function OrgSettingsGeneralPage({
  params,
}: {
  params: Promise<{ slug: string }>;
}) {
  const { slug } = await params;

  return (
    <div className="px-8 py-10 space-y-8">
      <div>
        <p className="text-xs uppercase tracking-[0.2em] text-muted-foreground">Settings</p>
        <h1 className="mt-1 text-2xl font-semibold text-foreground">General</h1>
        <p className="mt-2 text-sm text-muted-foreground">
          Organization name, slug, and preferences.
        </p>
      </div>

      <div className="grid grid-cols-1 gap-4 max-w-2xl">
        <div className="rounded-xl border border-border bg-card p-6">
          <p className="text-xs uppercase tracking-[0.2em] text-muted-foreground">
            Organization Name
          </p>
          <div className="mt-4 rounded-lg border border-dashed border-border py-8 text-center">
            <p className="text-sm text-muted-foreground/60">{slug}</p>
          </div>
        </div>

        <div className="rounded-xl border border-border bg-card p-6">
          <p className="text-xs uppercase tracking-[0.2em] text-muted-foreground">Slug</p>
          <div className="mt-4 rounded-lg border border-dashed border-border py-6 px-4">
            <p className="font-mono text-sm text-muted-foreground">{slug}</p>
          </div>
        </div>
      </div>
    </div>
  );
}
