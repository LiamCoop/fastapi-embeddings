import { Button } from "@/components/ui/button";

export default function OrgIntegrationsPage() {
  return (
    <div className="px-8 py-10 space-y-8">
      <div className="flex items-center justify-between">
        <div>
          <p className="text-xs uppercase tracking-[0.2em] text-muted-foreground">Integrations</p>
          <h1 className="mt-1 text-2xl font-semibold text-foreground">MCP Connections</h1>
          <p className="mt-2 text-sm text-muted-foreground">
            Connect external LLMs and agents via the Model Context Protocol.
          </p>
        </div>
        <Button disabled>New Connection</Button>
      </div>

      <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
        <div className="rounded-xl border border-border bg-card p-6">
          <p className="text-xs uppercase tracking-[0.2em] text-muted-foreground">MCP Servers</p>
          <div className="mt-4 rounded-lg border border-dashed border-border py-12 text-center">
            <p className="text-sm text-muted-foreground/60">No connections yet</p>
            <p className="mt-1 text-xs text-muted-foreground/40">
              Connect an MCP server to expose your corpus to external agents.
            </p>
          </div>
        </div>

        <div className="rounded-xl border border-border bg-card p-6">
          <p className="text-xs uppercase tracking-[0.2em] text-muted-foreground">SDK Quickstart</p>
          <div className="mt-4 rounded-lg border border-dashed border-border py-12 text-center">
            <p className="text-sm text-muted-foreground/60">Code snippets will appear here</p>
          </div>
        </div>

        <div className="rounded-xl border border-border bg-card p-6">
          <p className="text-xs uppercase tracking-[0.2em] text-muted-foreground">
            Usage Analytics
          </p>
          <div className="mt-4 rounded-lg border border-dashed border-border py-10 text-center">
            <p className="text-sm text-muted-foreground/60">No usage data</p>
          </div>
        </div>

        <div className="rounded-xl border border-border bg-card p-6">
          <p className="text-xs uppercase tracking-[0.2em] text-muted-foreground">Rate Limits</p>
          <div className="mt-4 rounded-lg border border-dashed border-border py-10 text-center">
            <p className="text-sm text-muted-foreground/60">Not configured</p>
          </div>
        </div>
      </div>
    </div>
  );
}
