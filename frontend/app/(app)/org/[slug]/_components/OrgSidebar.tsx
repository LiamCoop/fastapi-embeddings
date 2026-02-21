"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { UserButton, SignOutButton } from "@clerk/nextjs";
import { cn } from "@/lib/utils";
import { KbNav } from "./KbNav";

const ORG_NAV = [
  { label: "Overview", segment: "" },
  { label: "Knowledge Bases", segment: "/knowledge" },
  { label: "Integrations", segment: "/integrations" },
];

const ORG_SETTINGS_NAV = [
  { label: "Settings", segment: "/settings/general" },
  { label: "Members", segment: "/settings/members" },
  { label: "Billing", segment: "/settings/billing" },
];

export function OrgSidebar({ slug }: { slug: string }) {
  const pathname = usePathname();
  const orgBase = `/org/${slug}`;

  // Detect if we're inside a specific KB
  const kbMatch = pathname.match(new RegExp(`^/org/${slug}/knowledge/([^/]+)`));
  const kbId = kbMatch?.[1] ?? null;

  const isNavActive = (segment: string) => {
    const href = `${orgBase}${segment}`;
    if (segment === "") {
      // Overview is active only on exact org root
      return pathname === orgBase;
    }
    return pathname === href || pathname.startsWith(href + "/");
  };

  const isSettingsActive = (segment: string) => {
    const href = `${orgBase}${segment}`;
    return pathname === href || pathname.startsWith(href);
  };

  return (
    <aside className="flex h-full w-52 shrink-0 flex-col border-r border-border bg-card/50">
      <div className="border-b border-border px-5 py-4">
        <p className="text-[10px] uppercase tracking-[0.3em] text-muted-foreground">Ragtime</p>
        <p className="mt-0.5 text-sm font-semibold text-foreground">{slug}</p>
      </div>

      <nav className="flex-1 overflow-y-auto px-2 py-3 space-y-0.5">
        {ORG_NAV.map((item) => (
          <Link
            key={item.label}
            href={`${orgBase}${item.segment}`}
            className={cn(
              "flex items-center rounded-lg px-3 py-2 text-sm transition",
              isNavActive(item.segment)
                ? "bg-secondary font-medium text-foreground"
                : "text-muted-foreground hover:bg-secondary/60 hover:text-foreground"
            )}
          >
            {item.label}
          </Link>
        ))}

        <div className="pt-3 pb-1">
          <p className="px-3 text-[10px] uppercase tracking-[0.2em] text-muted-foreground/60">
            Organization
          </p>
        </div>

        {ORG_SETTINGS_NAV.map((item) => (
          <Link
            key={item.label}
            href={`${orgBase}${item.segment}`}
            className={cn(
              "flex items-center rounded-lg px-3 py-2 text-sm transition",
              isSettingsActive(item.segment)
                ? "bg-secondary font-medium text-foreground"
                : "text-muted-foreground hover:bg-secondary/60 hover:text-foreground"
            )}
          >
            {item.label}
          </Link>
        ))}

        {kbId && <KbNav slug={slug} kbId={kbId} />}
      </nav>

      <div className="border-t border-border px-4 py-3 flex items-center justify-between gap-2">
        <UserButton />
        <SignOutButton>
          <button className="text-xs text-muted-foreground hover:text-foreground transition">
            Sign out
          </button>
        </SignOutButton>
      </div>
    </aside>
  );
}
