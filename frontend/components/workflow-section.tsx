import { ArrowDown } from "lucide-react";

const steps = [
  {
    number: "01",
    title: "Ingest",
    description:
      "Bring in documents, URLs, and data sources. Ragtime parses, chunks, and indexes your knowledge automatically.",
  },
  {
    number: "02",
    title: "Inspect",
    description:
      "Query your knowledge and see exactly what gets retrieved, with relevance scores and source attribution.",
  },
  {
    number: "03",
    title: "Evaluate",
    description:
      "Run test suites against your knowledge. Measure relevance, faithfulness, and completeness across versions.",
  },
  {
    number: "04",
    title: "Monitor",
    description:
      "Track production query behavior, detect drift, and get alerted before quality degrades for your users.",
  },
];

export function WorkflowSection() {
  return (
    <section id="workflow" className="relative px-6 py-32">
      <div className="mx-auto max-w-6xl">
        <div className="mb-16 text-center">
          <p className="mb-3 text-sm font-medium uppercase tracking-widest text-primary">
            How it works
          </p>
          <h2 className="text-balance text-3xl font-bold tracking-tight text-foreground md:text-4xl">
            A continuous loop of trust
          </h2>
          <p className="mx-auto mt-4 max-w-2xl text-pretty text-lg leading-relaxed text-muted-foreground">
            Reliability is not a one-time setup. Ragtime helps your team
            iterate continuously across ingestion, retrieval, evaluation, and
            production monitoring.
          </p>
        </div>

        <div className="mx-auto flex max-w-xl flex-col gap-0">
          {steps.map((step, index) => (
            <div key={step.number} className="relative flex flex-col items-center">
              <div className="flex w-full items-start gap-6 rounded-xl border border-border bg-card p-6 transition-colors hover:border-primary/30">
                <span className="shrink-0 font-mono text-3xl font-bold text-primary/30">
                  {step.number}
                </span>
                <div>
                  <h3 className="text-lg font-semibold text-foreground">
                    {step.title}
                  </h3>
                  <p className="mt-1 text-sm leading-relaxed text-muted-foreground">
                    {step.description}
                  </p>
                </div>
              </div>
              {index < steps.length - 1 && (
                <div className="flex h-10 items-center justify-center">
                  <ArrowDown className="size-4 text-muted-foreground/30" />
                </div>
              )}
            </div>
          ))}
        </div>

        {/* Loop-back visual */}
        <div className="mx-auto mt-6 flex max-w-xl items-center justify-center">
          <div className="flex items-center gap-3 rounded-full border border-dashed border-primary/20 bg-primary/5 px-6 py-2.5">
            <div className="size-1.5 animate-pulse rounded-full bg-primary" />
            <span className="font-mono text-xs text-primary/70">
              iterate and improve continuously
            </span>
          </div>
        </div>
      </div>
    </section>
  );
}
