"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { cn } from "@/lib/utils";

const KB_NAV_ITEMS = [
  { label: "Overview", segment: "" },
  { label: "Ingestion", segment: "/ingestion" },
  { label: "Chunks", segment: "/chunks" },
  { label: "Retrieval", segment: "/retrieval" },
  { label: "Playground", segment: "/playground" },
  { label: "Evaluation", segment: "/evaluation" },
  { label: "Inspector", segment: "/inspector" },
  { label: "Settings", segment: "/settings" },
];

export function KbNav({ slug, kbId }: { slug: string; kbId: string }) {
  const pathname = usePathname();
  const kbBase = `/org/${slug}/knowledge/${kbId}`;

  return (
    <div className="mt-3 border-t border-border pt-3">
      <Link
        href={`/org/${slug}/knowledge`}
        className="flex items-center gap-1 px-3 py-1.5 text-xs text-muted-foreground hover:text-foreground transition"
      >
        ← All Knowledge Bases
      </Link>
      <div className="mt-2 px-3 pb-1">
        <p className="font-mono text-[10px] uppercase tracking-[0.2em] text-muted-foreground/60">
          {kbId.slice(0, 8)}
        </p>
      </div>
      <div className="space-y-0.5">
        {KB_NAV_ITEMS.map((item) => {
          const href = `${kbBase}${item.segment}`;
          const isActive =
            item.segment === ""
              ? pathname === kbBase
              : pathname.startsWith(href);
          return (
            <Link
              key={item.label}
              href={href}
              className={cn(
                "flex items-center rounded-lg px-3 py-2 text-sm transition",
                isActive
                  ? "bg-secondary font-medium text-foreground"
                  : "text-muted-foreground hover:bg-secondary/60 hover:text-foreground"
              )}
            >
              {item.label}
            </Link>
          );
        })}
      </div>
    </div>
  );
}
