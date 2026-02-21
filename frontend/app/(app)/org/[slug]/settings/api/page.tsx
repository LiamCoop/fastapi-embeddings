import { Button } from "@/components/ui/button";

export default function OrgSettingsApiPage() {
  return (
    <div className="px-8 py-10 space-y-8">
      <div className="flex items-center justify-between">
        <div>
          <p className="text-xs uppercase tracking-[0.2em] text-muted-foreground">Settings</p>
          <h1 className="mt-1 text-2xl font-semibold text-foreground">API Keys</h1>
          <p className="mt-2 text-sm text-muted-foreground">Manage API keys and access tokens.</p>
        </div>
        <Button disabled>New API Key</Button>
      </div>

      <div className="rounded-xl border border-border bg-card p-6">
        <p className="text-xs uppercase tracking-[0.2em] text-muted-foreground">API Keys</p>
        <div className="mt-4 rounded-lg border border-dashed border-border py-12 text-center">
          <p className="text-sm text-muted-foreground/60">No API keys created yet</p>
          <p className="mt-1 text-xs text-muted-foreground/40">
            Create an API key to authenticate programmatic access to your knowledge bases.
          </p>
        </div>
      </div>
    </div>
  );
}
