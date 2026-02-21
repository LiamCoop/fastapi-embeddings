import { Separator } from "@/components/ui/separator";

export function Footer() {
  return (
    <footer className="px-6 pb-12">
      <div className="mx-auto max-w-6xl">
        <Separator className="mb-8" />
        <div className="flex flex-col items-center justify-between gap-6 md:flex-row">
          <div className="flex items-center gap-2">
            <div className="flex size-6 items-center justify-center rounded bg-primary">
              <svg
                width="12"
                height="12"
                viewBox="0 0 18 18"
                fill="none"
                className="text-primary-foreground"
              >
                <path
                  d="M3 3h4.5v4.5H3V3zm7.5 0H15v4.5h-4.5V3zM3 10.5h4.5V15H3v-4.5zm7.5 0H15V15h-4.5v-4.5z"
                  fill="currentColor"
                />
              </svg>
            </div>
            <span className="text-sm font-semibold text-foreground">
              Ragtime
            </span>
          </div>

          <div className="flex flex-wrap items-center justify-center gap-6">
            {["Platform", "Docs", "Pricing", "Blog", "Changelog"].map((link) => (
              <a
                key={link}
                href="#"
                className="text-xs text-muted-foreground transition-colors hover:text-foreground"
              >
                {link}
              </a>
            ))}
          </div>

          <p className="text-xs text-muted-foreground/50">Ragtime</p>
        </div>
      </div>
    </footer>
  );
}
