import { SignUpButton } from "@clerk/nextjs";
import { Button } from "@/components/ui/button";
import { ArrowRight } from "lucide-react";

export function CtaSection() {
  return (
    <section className="relative px-6 py-32">
      {/* Accent glow */}
      <div className="pointer-events-none absolute bottom-0 left-1/2 h-[400px] w-[600px] -translate-x-1/2 rounded-full bg-primary/6 blur-[100px]" />

      <div className="relative mx-auto flex max-w-3xl flex-col items-center text-center">
        <h2 className="text-balance text-3xl font-bold tracking-tight text-foreground md:text-5xl">
          Build AI your team can rely on
        </h2>
        <p className="mt-6 max-w-xl text-pretty text-lg leading-relaxed text-muted-foreground">
          Stop guessing whether your AI is giving good answers. Ragtime gives
          you the tools to know — and the workflow to continuously improve.
        </p>
        <div className="mt-10 flex flex-col items-center gap-4 sm:flex-row">
          <SignUpButton>
            <Button size="lg" className="gap-2 px-8 text-sm font-medium">
              Get started free
              <ArrowRight className="size-4" />
            </Button>
          </SignUpButton>
          <Button
            variant="outline"
            size="lg"
            className="px-8 text-sm font-medium"
          >
            Talk to the team
          </Button>
        </div>
      </div>
    </section>
  );
}
