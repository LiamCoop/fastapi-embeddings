"use client";

import { useState } from "react";
import { cn } from "@/lib/utils";
import {
  LayoutDashboard,
  FolderKanban,
  BookOpen,
  Search,
  FlaskConical,
  TestTubeDiagonal,
  Activity,
  Plug,
  Shield,
} from "lucide-react";

const layers = [
  {
    id: "workspace",
    name: "Workspace",
    tagline: "Know the state of your AI at a glance.",
    description:
      "A high-level reliability hub showing recent activity, ingestion health, and quick entry points so users always understand the current state of their AI knowledge.",
    icon: LayoutDashboard,
    visual: WorkspaceVisual,
  },
  {
    id: "projects",
    name: "Projects",
    tagline: "Organize knowledge around what you're building.",
    description:
      "A contextual container that groups corpuses, evaluations, and integrations around a single product, team, or use case.",
    icon: FolderKanban,
    visual: ProjectsVisual,
  },
  {
    id: "knowledge",
    name: "Knowledge",
    tagline: "Turn raw information into reliable knowledge.",
    description:
      "The place where users ingest, structure, and refine their knowledge bases while gaining visibility into how information is parsed and represented.",
    icon: BookOpen,
    visual: KnowledgeVisual,
  },
  {
    id: "retrieval",
    name: "Retrieval Inspector",
    tagline: "See why your AI knows — or doesn't.",
    description:
      "A debugging surface that lets users understand exactly how queries retrieve information and why relevant content was included or missed.",
    icon: Search,
    visual: RetrievalVisual,
  },
  {
    id: "playground",
    name: "Playground",
    tagline: "Experiment safely. Ship confidently.",
    description:
      "A safe experimentation space where users test questions, inspect grounding, and compare configurations before shipping changes.",
    icon: FlaskConical,
    visual: PlaygroundVisual,
  },
  {
    id: "evaluation",
    name: "Evaluation Studio",
    tagline: "Measure trust. Prevent regressions.",
    description:
      "A quality lab that measures retrieval and answer performance, tracks regressions, and helps teams systematically improve AI reliability.",
    icon: TestTubeDiagonal,
    visual: EvaluationVisual,
  },
  {
    id: "health",
    name: "AI Health",
    tagline: "Monitor the reliability of your knowledge in production.",
    description:
      "An observability dashboard providing insight into real-world query behavior, failures, drift, and overall knowledge effectiveness.",
    icon: Activity,
    visual: HealthVisual,
  },
  {
    id: "integrations",
    name: "Integrations",
    tagline: "Connect your AI to knowledge with confidence.",
    description:
      "The connection layer where users obtain credentials, configure access, and enable external LLMs and agents to interact with their knowledge safely.",
    icon: Plug,
    visual: IntegrationsVisual,
  },
  {
    id: "admin",
    name: "Admin & Settings",
    tagline: "Control access, environments, and scale.",
    description:
      "The organizational control surface for permissions, environments, usage tracking, and global configuration across projects.",
    icon: Shield,
    visual: AdminVisual,
  },
];

export function PlatformLayers() {
  const [activeIndex, setActiveIndex] = useState(0);
  const activeLayer = layers[activeIndex];

  return (
    <section id="layers" className="relative px-6 py-32">
      <div className="mx-auto max-w-6xl">
        <div className="mb-16 max-w-2xl">
          <p className="mb-3 text-sm font-medium uppercase tracking-widest text-primary">
            Platform layers
          </p>
          <h2 className="text-balance text-3xl font-bold tracking-tight text-foreground md:text-4xl">
            Dip in and out of every surface your AI reliability needs
          </h2>
          <p className="mt-4 text-pretty text-lg leading-relaxed text-muted-foreground">
            Ragtime is organized into focused layers. Each one gives you a
            different lens on your knowledge system — from high-level health
            to granular retrieval debugging.
          </p>
        </div>

        <div className="grid gap-12 lg:grid-cols-[340px_1fr]">
          {/* Layer selector — vertical tabs */}
          <div className="flex flex-col gap-1" role="tablist" aria-label="Platform layers">
            {layers.map((layer, index) => {
              const Icon = layer.icon;
              const isActive = index === activeIndex;
              return (
                <button
                  key={layer.id}
                  role="tab"
                  aria-selected={isActive}
                  aria-controls={`panel-${layer.id}`}
                  onClick={() => setActiveIndex(index)}
                  className={cn(
                    "group flex items-start gap-3 rounded-lg px-4 py-3 text-left transition-all",
                    isActive
                      ? "bg-secondary text-foreground"
                      : "text-muted-foreground hover:bg-secondary/50 hover:text-foreground"
                  )}
                >
                  <Icon
                    className={cn(
                      "mt-0.5 size-4 shrink-0 transition-colors",
                      isActive ? "text-primary" : "text-muted-foreground group-hover:text-foreground"
                    )}
                  />
                  <div className="min-w-0">
                    <div className="text-sm font-medium">{layer.name}</div>
                    {isActive && (
                      <div className="mt-1 text-xs leading-relaxed text-muted-foreground">
                        {layer.tagline}
                      </div>
                    )}
                  </div>
                </button>
              );
            })}
          </div>

          {/* Layer detail panel */}
          <div
            id={`panel-${activeLayer.id}`}
            role="tabpanel"
            className="flex flex-col gap-6"
          >
            <div className="overflow-hidden rounded-xl border border-border bg-card">
              <div className="flex items-center gap-2 border-b border-border px-5 py-3">
                <div className="flex gap-1.5">
                  <div className="size-2.5 rounded-full bg-muted-foreground/20" />
                  <div className="size-2.5 rounded-full bg-muted-foreground/20" />
                  <div className="size-2.5 rounded-full bg-muted-foreground/20" />
                </div>
                <span className="ml-2 font-mono text-xs text-muted-foreground">
                  ragtime / {activeLayer.name.toLowerCase().replace(/ & /g, "-").replace(/ /g, "-")}
                </span>
              </div>
              <div className="flex min-h-[340px] items-center justify-center p-8">
                <activeLayer.visual />
              </div>
            </div>

            <p className="max-w-xl text-sm leading-relaxed text-muted-foreground">
              {activeLayer.description}
            </p>
          </div>
        </div>
      </div>
    </section>
  );
}

/* -----------------------------------------------------------------------
   Mini visuals for each layer — lightweight, schematic, monochrome+accent
   ----------------------------------------------------------------------- */

function WorkspaceVisual() {
  return (
    <div className="flex w-full max-w-md flex-col gap-4">
      <div className="flex items-center gap-3">
        <div className="h-2 w-2 rounded-full bg-primary" />
        <span className="font-mono text-xs text-muted-foreground">System status: healthy</span>
      </div>
      <div className="grid grid-cols-3 gap-3">
        {["Documents ingested", "Queries (24h)", "Avg. relevance"].map((label, i) => (
          <div key={label} className="rounded-lg border border-border bg-secondary/50 p-4">
            <div className="font-mono text-2xl font-bold text-foreground">
              {["12,847", "3,291", "94.2%"][i]}
            </div>
            <div className="mt-1 text-xs text-muted-foreground">{label}</div>
          </div>
        ))}
      </div>
      <div className="flex gap-2">
        {[85, 92, 78, 95, 88, 91, 87, 93, 90, 86, 94, 89].map((h, i) => (
          <div key={i} className="flex flex-1 flex-col justify-end">
            <div
              className="rounded-sm bg-primary/30"
              style={{ height: `${h * 0.6}px` }}
            />
          </div>
        ))}
      </div>
    </div>
  );
}

function ProjectsVisual() {
  const projects = [
    { name: "Customer Support Bot", corpuses: 4, status: "active" },
    { name: "Internal Docs Search", corpuses: 2, status: "active" },
    { name: "Product FAQ v2", corpuses: 1, status: "draft" },
  ];
  return (
    <div className="flex w-full max-w-md flex-col gap-3">
      {projects.map((p) => (
        <div
          key={p.name}
          className="flex items-center justify-between rounded-lg border border-border bg-secondary/30 px-4 py-3"
        >
          <div className="flex items-center gap-3">
            <FolderKanban className="size-4 text-primary" />
            <div>
              <div className="text-sm font-medium text-foreground">{p.name}</div>
              <div className="text-xs text-muted-foreground">
                {p.corpuses} {p.corpuses === 1 ? "corpus" : "corpuses"}
              </div>
            </div>
          </div>
          <span
            className={cn(
              "rounded-full px-2 py-0.5 font-mono text-[10px] uppercase tracking-wider",
              p.status === "active"
                ? "bg-primary/10 text-primary"
                : "bg-muted text-muted-foreground"
            )}
          >
            {p.status}
          </span>
        </div>
      ))}
    </div>
  );
}

function KnowledgeVisual() {
  const docs = [
    { name: "help-center-articles.pdf", chunks: 847, status: "indexed" },
    { name: "api-reference-v3.md", chunks: 312, status: "processing" },
    { name: "release-notes-2025.html", chunks: 156, status: "indexed" },
  ];
  return (
    <div className="flex w-full max-w-md flex-col gap-3">
      <div className="flex items-center gap-2 text-xs text-muted-foreground">
        <BookOpen className="size-3.5" />
        <span className="font-mono">corpus / customer-support</span>
      </div>
      {docs.map((doc) => (
        <div
          key={doc.name}
          className="flex items-center justify-between rounded-md border border-border px-4 py-2.5"
        >
          <div className="flex items-center gap-2">
            <div className="h-4 w-0.5 rounded-full bg-primary" />
            <span className="font-mono text-xs text-foreground">{doc.name}</span>
          </div>
          <div className="flex items-center gap-3">
            <span className="font-mono text-[10px] text-muted-foreground">
              {doc.chunks} chunks
            </span>
            <span
              className={cn(
                "size-2 rounded-full",
                doc.status === "indexed" ? "bg-primary" : "bg-chart-4 animate-pulse"
              )}
            />
          </div>
        </div>
      ))}
    </div>
  );
}

function RetrievalVisual() {
  const results = [
    { text: "To reset your password, navigate to Settings > Security...", score: 0.94, relevant: true },
    { text: "Our pricing plans start at $29/month for the basic tier...", score: 0.67, relevant: false },
    { text: "Password requirements include at least 8 characters...", score: 0.89, relevant: true },
  ];
  return (
    <div className="flex w-full max-w-md flex-col gap-4">
      <div className="rounded-md border border-primary/30 bg-primary/5 px-4 py-2.5">
        <span className="font-mono text-xs text-muted-foreground">query:</span>
        <span className="ml-2 text-sm text-foreground">
          {'"How do I reset my password?"'}
        </span>
      </div>
      <div className="flex flex-col gap-2">
        {results.map((r, i) => (
          <div
            key={i}
            className={cn(
              "flex items-start gap-3 rounded-md border px-4 py-3",
              r.relevant
                ? "border-primary/20 bg-primary/5"
                : "border-border bg-secondary/20"
            )}
          >
            <span
              className={cn(
                "mt-0.5 shrink-0 font-mono text-xs font-bold",
                r.relevant ? "text-primary" : "text-muted-foreground"
              )}
            >
              {r.score.toFixed(2)}
            </span>
            <span className="text-xs leading-relaxed text-muted-foreground">
              {r.text}
            </span>
          </div>
        ))}
      </div>
    </div>
  );
}

function PlaygroundVisual() {
  return (
    <div className="flex w-full max-w-md flex-col gap-4">
      <div className="flex gap-2">
        {["Config A", "Config B"].map((tab, i) => (
          <div
            key={tab}
            className={cn(
              "rounded-md px-3 py-1.5 font-mono text-xs",
              i === 0
                ? "bg-primary/10 text-primary"
                : "bg-secondary text-muted-foreground"
            )}
          >
            {tab}
          </div>
        ))}
      </div>
      <div className="rounded-md border border-border bg-secondary/30 px-4 py-3">
        <div className="text-xs text-muted-foreground">Question</div>
        <div className="mt-1 text-sm text-foreground">
          {"What's the refund policy for annual plans?"}
        </div>
      </div>
      <div className="rounded-md border border-primary/20 bg-primary/5 px-4 py-3">
        <div className="flex items-center gap-2">
          <div className="size-1.5 rounded-full bg-primary" />
          <span className="text-xs font-medium text-primary">Grounded response</span>
        </div>
        <div className="mt-2 text-xs leading-relaxed text-muted-foreground">
          Annual plans can be refunded within the first 14 days. After that period,
          you may switch to monthly billing at the end of your term.
        </div>
        <div className="mt-2 flex gap-1">
          {["source-1", "source-3"].map((s) => (
            <span key={s} className="rounded bg-secondary px-1.5 py-0.5 font-mono text-[10px] text-muted-foreground">
              {s}
            </span>
          ))}
        </div>
      </div>
    </div>
  );
}

function EvaluationVisual() {
  const metrics = [
    { name: "Relevance", current: 94.2, previous: 91.8, direction: "up" },
    { name: "Faithfulness", current: 97.1, previous: 97.3, direction: "down" },
    { name: "Completeness", current: 88.5, previous: 85.2, direction: "up" },
  ];
  return (
    <div className="flex w-full max-w-md flex-col gap-4">
      <div className="flex items-center gap-2 font-mono text-xs text-muted-foreground">
        <TestTubeDiagonal className="size-3.5" />
        eval-run-2847 — 200 test cases
      </div>
      {metrics.map((m) => (
        <div key={m.name} className="flex items-center justify-between rounded-md border border-border px-4 py-3">
          <span className="text-sm text-foreground">{m.name}</span>
          <div className="flex items-center gap-3">
            <span className="font-mono text-xs text-muted-foreground line-through">
              {m.previous}%
            </span>
            <span
              className={cn(
                "font-mono text-sm font-bold",
                m.direction === "up" ? "text-primary" : "text-chart-1"
              )}
            >
              {m.current}%
            </span>
            <span className={cn("text-xs", m.direction === "up" ? "text-primary" : "text-muted-foreground")}>
              {m.direction === "up" ? "+" : ""}{(m.current - m.previous).toFixed(1)}
            </span>
          </div>
        </div>
      ))}
    </div>
  );
}

function HealthVisual() {
  return (
    <div className="flex w-full max-w-md flex-col gap-4">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <div className="size-2 rounded-full bg-primary" />
          <span className="text-sm font-medium text-foreground">Production health</span>
        </div>
        <span className="font-mono text-xs text-muted-foreground">Last 7 days</span>
      </div>
      <div className="grid grid-cols-7 gap-1.5">
        {Array.from({ length: 7 }).map((_, day) => (
          <div key={day} className="flex flex-col gap-1">
            {Array.from({ length: 4 }).map((_, slot) => {
              const heat = [
                [3, 3, 2, 3, 3, 2, 3],
                [2, 3, 3, 2, 1, 3, 3],
                [3, 2, 1, 3, 3, 3, 2],
                [2, 3, 3, 3, 2, 3, 3],
              ][slot][day];
              return (
                <div
                  key={slot}
                  className={cn(
                    "h-6 rounded-sm",
                    heat === 3
                      ? "bg-primary/40"
                      : heat === 2
                        ? "bg-primary/20"
                        : "bg-destructive/30"
                  )}
                />
              );
            })}
          </div>
        ))}
      </div>
      <div className="flex items-center gap-4 text-xs text-muted-foreground">
        <div className="flex items-center gap-1.5">
          <div className="size-2.5 rounded-sm bg-primary/40" />
          Healthy
        </div>
        <div className="flex items-center gap-1.5">
          <div className="size-2.5 rounded-sm bg-primary/20" />
          Degraded
        </div>
        <div className="flex items-center gap-1.5">
          <div className="size-2.5 rounded-sm bg-destructive/30" />
          Failing
        </div>
      </div>
    </div>
  );
}

function IntegrationsVisual() {
  const integrations = [
    { name: "OpenAI GPT-4o", type: "LLM", connected: true },
    { name: "Claude 3.5 Sonnet", type: "LLM", connected: true },
    { name: "Custom Agent", type: "MCP", connected: false },
  ];
  return (
    <div className="flex w-full max-w-md flex-col gap-3">
      <div className="font-mono text-xs text-muted-foreground">
        Active connections
      </div>
      {integrations.map((int) => (
        <div
          key={int.name}
          className="flex items-center justify-between rounded-md border border-border px-4 py-3"
        >
          <div className="flex items-center gap-3">
            <Plug className="size-4 text-muted-foreground" />
            <div>
              <div className="text-sm font-medium text-foreground">{int.name}</div>
              <div className="font-mono text-[10px] uppercase tracking-wider text-muted-foreground">
                {int.type}
              </div>
            </div>
          </div>
          <span
            className={cn(
              "size-2 rounded-full",
              int.connected ? "bg-primary" : "bg-muted-foreground/30"
            )}
          />
        </div>
      ))}
    </div>
  );
}

function AdminVisual() {
  return (
    <div className="flex w-full max-w-md flex-col gap-4">
      <div className="flex items-center justify-between rounded-md border border-border px-4 py-3">
        <span className="text-sm text-foreground">Team members</span>
        <span className="font-mono text-sm font-bold text-foreground">12</span>
      </div>
      <div className="flex items-center justify-between rounded-md border border-border px-4 py-3">
        <span className="text-sm text-foreground">Environments</span>
        <div className="flex gap-1.5">
          {["prod", "staging", "dev"].map((env) => (
            <span
              key={env}
              className="rounded bg-secondary px-2 py-0.5 font-mono text-[10px] text-muted-foreground"
            >
              {env}
            </span>
          ))}
        </div>
      </div>
      <div className="flex items-center justify-between rounded-md border border-border px-4 py-3">
        <span className="text-sm text-foreground">API usage this month</span>
        <span className="font-mono text-sm text-primary">847K requests</span>
      </div>
    </div>
  );
}
