export default function OrgSettingsBillingPage() {
  return (
    <div className="px-8 py-10 space-y-8">
      <div>
        <p className="text-xs uppercase tracking-[0.2em] text-muted-foreground">Settings</p>
        <h1 className="mt-1 text-2xl font-semibold text-foreground">Billing</h1>
        <p className="mt-2 text-sm text-muted-foreground">Plan, usage, and invoices.</p>
      </div>

      <div className="grid grid-cols-1 gap-4 md:grid-cols-3">
        <div className="rounded-xl border border-border bg-card p-6">
          <p className="text-xs uppercase tracking-[0.2em] text-muted-foreground">Current Plan</p>
          <div className="mt-4">
            <p className="text-lg font-semibold text-foreground">Free</p>
            <p className="mt-1 text-xs text-muted-foreground">No credit card required</p>
          </div>
        </div>

        <div className="rounded-xl border border-border bg-card p-6">
          <p className="text-xs uppercase tracking-[0.2em] text-muted-foreground">Usage</p>
          <div className="mt-4 rounded-lg border border-dashed border-border py-8 text-center">
            <p className="text-sm text-muted-foreground/60">No usage data</p>
          </div>
        </div>

        <div className="rounded-xl border border-border bg-card p-6">
          <p className="text-xs uppercase tracking-[0.2em] text-muted-foreground">Invoices</p>
          <div className="mt-4 rounded-lg border border-dashed border-border py-8 text-center">
            <p className="text-sm text-muted-foreground/60">No invoices yet</p>
          </div>
        </div>
      </div>
    </div>
  );
}
