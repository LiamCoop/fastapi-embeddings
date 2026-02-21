import { Button } from "@/components/ui/button";

export default function ProjectsPage() {
  return (
    <div className="px-8 py-10 space-y-8">
      <div className="flex items-center justify-between">
        <div>
          <p className="text-xs uppercase tracking-[0.2em] text-muted-foreground">Organization</p>
          <h1 className="mt-1 text-2xl font-semibold text-foreground">Projects</h1>
          <p className="mt-2 text-sm text-muted-foreground">
            Logical groupings of corpuses, integrations, and evaluations.
          </p>
        </div>
        <Button disabled>New Project</Button>
      </div>

      <div className="rounded-xl border border-dashed border-border px-8 py-20 text-center">
        <p className="text-sm text-muted-foreground">Projects coming soon</p>
        <p className="mt-1 text-xs text-muted-foreground/60">
          Group your corpuses and integrations into logical knowledge domains.
        </p>
      </div>
    </div>
  );
}
