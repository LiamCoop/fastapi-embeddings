export default function SettingsPage() {
  return (
    <div className="px-8 py-10 space-y-8">
      <div>
        <p className="text-xs uppercase tracking-[0.2em] text-muted-foreground">Admin</p>
        <h1 className="mt-1 text-2xl font-semibold text-foreground">Settings</h1>
      </div>

      <div className="grid grid-cols-1 gap-4 md:grid-cols-3">
        <div className="rounded-xl border border-border bg-card p-6">
          <p className="text-xs uppercase tracking-[0.2em] text-muted-foreground">Access & Permissions</p>
          <div className="mt-4 space-y-2">
            <div className="rounded-md border border-dashed border-border px-4 py-3 text-sm text-muted-foreground/60">
              Roles
            </div>
            <div className="rounded-md border border-dashed border-border px-4 py-3 text-sm text-muted-foreground/60">
              Corpus access control
            </div>
            <div className="rounded-md border border-dashed border-border px-4 py-3 text-sm text-muted-foreground/60">
              Audit logs
            </div>
          </div>
        </div>

        <div className="rounded-xl border border-border bg-card p-6">
          <p className="text-xs uppercase tracking-[0.2em] text-muted-foreground">Usage & Billing</p>
          <div className="mt-4 space-y-2">
            <div className="rounded-md border border-dashed border-border px-4 py-3 text-sm text-muted-foreground/60">
              Token usage
            </div>
            <div className="rounded-md border border-dashed border-border px-4 py-3 text-sm text-muted-foreground/60">
              Retrieval volume
            </div>
            <div className="rounded-md border border-dashed border-border px-4 py-3 text-sm text-muted-foreground/60">
              Cost attribution
            </div>
          </div>
        </div>

        <div className="rounded-xl border border-border bg-card p-6">
          <p className="text-xs uppercase tracking-[0.2em] text-muted-foreground">Organization</p>
          <div className="mt-4 space-y-2">
            <div className="rounded-md border border-dashed border-border px-4 py-3 text-sm text-muted-foreground/60">
              Environments (dev / prod)
            </div>
            <div className="rounded-md border border-dashed border-border px-4 py-3 text-sm text-muted-foreground/60">
              Global defaults
            </div>
            <div className="rounded-md border border-dashed border-border px-4 py-3 text-sm text-muted-foreground/60">
              Organization preferences
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
