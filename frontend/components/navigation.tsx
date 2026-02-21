"use client";

import { useState } from "react";
import { SignInButton, SignUpButton } from "@clerk/nextjs";
import { Button } from "@/components/ui/button";
import { Menu, X } from "lucide-react";

export function Navigation() {
  const [mobileOpen, setMobileOpen] = useState(false);

  return (
    <header className="fixed top-0 left-0 right-0 z-50 border-b border-border/50 bg-background/80 backdrop-blur-xl">
      <nav className="mx-auto flex max-w-6xl items-center justify-between px-6 py-4">
        <div className="flex items-center gap-2">
          <div className="flex size-8 items-center justify-center rounded-md bg-primary">
            <svg
              width="18"
              height="18"
              viewBox="0 0 18 18"
              fill="none"
              className="text-primary-foreground"
            >
              <path
                d="M3 3h4.5v4.5H3V3zm7.5 0H15v4.5h-4.5V3zM3 10.5h4.5V15H3v-4.5zm7.5 0H15V15h-4.5v-4.5z"
                fill="currentColor"
                opacity="0.9"
              />
              <path
                d="M5.25 5.25h-.001m7.5 0h-.001m-7.5 7.5h-.001m7.5 0h-.001"
                stroke="currentColor"
                strokeWidth="1.5"
                strokeLinecap="round"
              />
            </svg>
          </div>
          <span className="text-lg font-semibold tracking-tight text-foreground">
            Ragtime
          </span>
        </div>

        <div className="hidden items-center gap-8 md:flex">
          <a
            href="#layers"
            className="text-sm text-muted-foreground transition-colors hover:text-foreground"
          >
            Platform
          </a>
          <a
            href="#workflow"
            className="text-sm text-muted-foreground transition-colors hover:text-foreground"
          >
            How it works
          </a>
          <a
            href="#"
            className="text-sm text-muted-foreground transition-colors hover:text-foreground"
          >
            Docs
          </a>
          <a
            href="#"
            className="text-sm text-muted-foreground transition-colors hover:text-foreground"
          >
            Pricing
          </a>
        </div>

        <div className="hidden items-center gap-3 md:flex">
          <SignInButton>
            <Button variant="ghost" size="sm" className="text-muted-foreground hover:text-foreground">
              Sign in
            </Button>
          </SignInButton>
          <SignUpButton>
            <Button size="sm">Get started</Button>
          </SignUpButton>
        </div>

        <button
          className="text-muted-foreground md:hidden"
          onClick={() => setMobileOpen(!mobileOpen)}
          aria-label={mobileOpen ? "Close menu" : "Open menu"}
        >
          {mobileOpen ? <X className="size-5" /> : <Menu className="size-5" />}
        </button>
      </nav>

      {mobileOpen && (
        <div className="border-t border-border/50 bg-background px-6 pb-6 pt-4 md:hidden">
          <div className="flex flex-col gap-4">
            <a
              href="#layers"
              className="text-sm text-muted-foreground transition-colors hover:text-foreground"
              onClick={() => setMobileOpen(false)}
            >
              Platform
            </a>
            <a
              href="#workflow"
              className="text-sm text-muted-foreground transition-colors hover:text-foreground"
              onClick={() => setMobileOpen(false)}
            >
              How it works
            </a>
            <a
              href="#"
              className="text-sm text-muted-foreground transition-colors hover:text-foreground"
            >
              Docs
            </a>
            <a
              href="#"
              className="text-sm text-muted-foreground transition-colors hover:text-foreground"
            >
              Pricing
            </a>
            <div className="flex gap-3 pt-2">
              <SignInButton>
                <Button variant="ghost" size="sm" className="text-muted-foreground hover:text-foreground">
                  Sign in
                </Button>
              </SignInButton>
              <SignUpButton>
                <Button size="sm">Get started</Button>
              </SignUpButton>
            </div>
          </div>
        </div>
      )}
    </header>
  );
}
