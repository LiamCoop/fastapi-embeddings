import { SignUpButton } from "@clerk/nextjs";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { ArrowRight } from "lucide-react";

export function Hero() {
  return (
    <section className="relative flex min-h-screen flex-col items-center justify-center overflow-hidden px-6 pt-20">
      {/* Subtle grid background */}
      <div
        className="pointer-events-none absolute inset-0 opacity-[0.03]"
        style={{
          backgroundImage:
            "linear-gradient(to right, currentColor 1px, transparent 1px), linear-gradient(to bottom, currentColor 1px, transparent 1px)",
          backgroundSize: "64px 64px",
        }}
      />

      {/* Accent glow */}
      <div className="pointer-events-none absolute top-1/4 left-1/2 h-[500px] w-[800px] -translate-x-1/2 -translate-y-1/2 rounded-full bg-primary/8 blur-[120px]" />

      <div className="relative z-10 mx-auto flex max-w-4xl flex-col items-center text-center">
        <Badge
          variant="outline"
          className="mb-8 border-border bg-secondary/50 px-4 py-1.5 text-xs font-medium text-muted-foreground"
        >
          <span className="mr-2 inline-block size-1.5 rounded-full bg-primary" />
          Now in early access
        </Badge>

        <h1 className="text-balance text-4xl font-bold leading-[1.1] tracking-tight text-foreground md:text-6xl lg:text-7xl">
          The platform for AI
          <br />
          <span className="text-primary">you can trust</span>
        </h1>

        <p className="mt-6 max-w-2xl text-pretty text-lg leading-relaxed text-muted-foreground md:text-xl">
          Ingest, inspect, evaluate, and monitor your AI knowledge.
          Ragtime gives your team a shared workspace to understand why
          your AI knows what it knows — and fix what it doesn{"'"}t.
        </p>

        <div className="mt-10 flex flex-col items-center gap-4 sm:flex-row">
          <SignUpButton>
            <Button size="lg" className="gap-2 px-8 text-sm font-medium">
              Start building
              <ArrowRight className="size-4" />
            </Button>
          </SignUpButton>
          <Button
            variant="outline"
            size="lg"
            className="px-8 text-sm font-medium"
            asChild
          >
            <a href="#layers">Explore the platform</a>
          </Button>
        </div>

        {/* Trust signals */}
        <div className="mt-20 flex flex-col items-center gap-4">
          <p className="text-xs font-medium uppercase tracking-widest text-muted-foreground/60">
            Trusted by AI teams at
          </p>
          <div className="flex flex-wrap items-center justify-center gap-x-10 gap-y-4">
            {["Acme Corp", "Meridian", "Lattice AI", "NovaSoft", "Helix"].map((company) => (
              <span
                key={company}
                className="text-sm font-medium tracking-wide text-muted-foreground/40"
              >
                {company}
              </span>
            ))}
          </div>
        </div>
      </div>

      {/* Scroll indicator */}
      <div className="absolute bottom-8 left-1/2 flex -translate-x-1/2 flex-col items-center gap-2">
        <span className="text-xs text-muted-foreground/40">Scroll</span>
        <div className="h-8 w-px bg-gradient-to-b from-muted-foreground/30 to-transparent" />
      </div>
    </section>
  );
}
