"use client";

import { useEffect, useState } from "react";
import { useRouter, useParams } from "next/navigation";
import { useUser } from "@clerk/nextjs";
import { Button } from "@/components/ui/button";
import { generateDicewareName } from "@/app/lib/diceware";
import {
  fetchOrgKnowledgeBases,
  knowledgeApiPath,
  type KnowledgeBase,
} from "@/app/lib/org-knowledge";

export default function OrgKnowledgeBasesPage() {
  const router = useRouter();
  const params = useParams();
  const slug = params?.slug as string;
  const { user, isLoaded } = useUser();
  const [items, setItems] = useState<KnowledgeBase[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [isCreating, setIsCreating] = useState(false);
  const [isCreateOpen, setIsCreateOpen] = useState(false);
  const [newName, setNewName] = useState("");
  const [createError, setCreateError] = useState<string | null>(null);
  const userId = user?.id ?? null;

  useEffect(() => {
    if (!isLoaded || !userId) return;

    let isActive = true;
    setIsLoading(true);
    setError(null);

    (async () => {
      try {
        const data = await fetchOrgKnowledgeBases(slug);
        if (!isActive) return;
        setItems(data.knowledge_bases ?? []);
      } catch (err) {
        if (!isActive) return;
        setError(err instanceof Error ? err.message : "Failed to load knowledge bases");
      } finally {
        if (!isActive) return;
        setIsLoading(false);
      }
    })();

    return () => {
      isActive = false;
    };
  }, [isLoaded, userId, slug]);

  const openCreateModal = () => {
    if (!userId) return;
    setCreateError(null);
    setNewName(generateDicewareName());
    setIsCreateOpen(true);
  };

  const closeCreateModal = () => {
    if (isCreating) return;
    setIsCreateOpen(false);
  };

  const handleCreate = async () => {
    if (!userId) return;
    setIsCreating(true);
    setCreateError(null);
    try {
      const name = newName.trim();
      if (!name) {
        setCreateError("Please enter a name for the corpus.");
        return;
      }
      const res = await fetch(knowledgeApiPath(slug), {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          name,
          metadata: { user_id: userId, system_name: name },
        }),
      });
      if (!res.ok) {
        const message = await res.text();
        throw new Error(message || `Request failed with status ${res.status}`);
      }
      const created = (await res.json()) as KnowledgeBase;
      router.push(`/org/${slug}/knowledge/${created.id}/ingestion`);
    } catch (err) {
      setCreateError(err instanceof Error ? err.message : "Failed to create knowledge base");
      return;
    } finally {
      setIsCreating(false);
    }
  };

  return (
    <div className="px-8 py-10 space-y-8">
      <div className="flex flex-wrap items-center justify-between gap-4">
        <div>
          <p className="text-xs uppercase tracking-[0.2em] text-muted-foreground">Knowledge</p>
          <h1 className="mt-1 text-2xl font-semibold text-foreground">Corpus Library</h1>
          <p className="mt-2 text-sm text-muted-foreground">
            Browse and manage your versioned knowledge bases.
          </p>
        </div>
        <Button onClick={openCreateModal} disabled={isCreating}>
          New Corpus
        </Button>
      </div>

      {error ? (
        <div className="rounded-xl border border-destructive/40 bg-destructive/10 px-4 py-3 text-sm text-destructive-foreground">
          {error}
        </div>
      ) : null}

      {isLoading ? (
        <div className="rounded-xl border border-border bg-card px-6 py-12 text-center text-sm text-muted-foreground">
          Loading corpuses...
        </div>
      ) : items.length === 0 ? (
        <div className="rounded-xl border border-dashed border-border px-6 py-16 text-center">
          <p className="text-sm text-muted-foreground">No corpuses yet</p>
          <p className="mt-1 text-xs text-muted-foreground/60">
            Create one to start building your knowledge base.
          </p>
        </div>
      ) : (
        <div className="grid grid-cols-1 gap-4 md:grid-cols-2 xl:grid-cols-3">
          {items.map((kb) => (
            <button
              key={kb.id}
              type="button"
              onClick={() => router.push(`/org/${slug}/knowledge/${kb.id}`)}
              className="group flex flex-col gap-3 rounded-xl border border-border bg-card px-6 py-5 text-left transition hover:border-primary/40 hover:bg-card/80"
            >
              <div className="flex items-center justify-between">
                <h2 className="text-base font-semibold text-foreground">{kb.name}</h2>
                <span className="rounded-full border border-border px-2 py-0.5 font-mono text-[10px] uppercase tracking-[0.2em] text-muted-foreground">
                  Active
                </span>
              </div>
              <p className="text-xs text-muted-foreground">
                Updated {new Date(kb.updated_at).toLocaleString()}
              </p>
              <span className="mt-auto text-xs font-semibold uppercase tracking-[0.2em] text-primary">
                Open
              </span>
            </button>
          ))}
        </div>
      )}

      {isCreateOpen ? (
        <div className="fixed inset-0 z-50 flex items-center justify-center px-4">
          <div
            className="absolute inset-0 bg-background/80 backdrop-blur-sm"
            onClick={closeCreateModal}
            aria-hidden="true"
          />
          <div className="relative w-full max-w-lg rounded-2xl border border-border bg-card p-6 shadow-2xl">
            <div className="flex items-start justify-between gap-4">
              <div>
                <p className="text-xs uppercase tracking-[0.2em] text-muted-foreground">
                  New Corpus
                </p>
                <h2 className="mt-2 text-xl font-semibold text-foreground">
                  Name your knowledge base
                </h2>
                <p className="mt-2 text-sm text-muted-foreground">
                  Give this corpus a clear, human-friendly name. You can change it later.
                </p>
              </div>
              <button
                type="button"
                onClick={closeCreateModal}
                className="rounded-full border border-border px-2 py-1 text-xs uppercase tracking-[0.2em] text-muted-foreground transition hover:border-primary/40 hover:text-foreground"
              >
                Close
              </button>
            </div>

            <div className="mt-6 space-y-2">
              <label className="text-xs font-semibold uppercase tracking-[0.2em] text-muted-foreground">
                Corpus name
              </label>
              <input
                value={newName}
                onChange={(event) => setNewName(event.target.value)}
                placeholder="e.g. Product docs, Support playbooks"
                className="h-11 w-full rounded-xl border border-input bg-background px-4 text-sm text-foreground outline-none transition focus:border-primary/60 focus:ring-2 focus:ring-primary/20"
              />
            </div>

            {createError ? (
              <div className="mt-4 rounded-xl border border-destructive/40 bg-destructive/10 px-4 py-3 text-sm text-destructive-foreground">
                {createError}
              </div>
            ) : null}

            <div className="mt-6 flex flex-wrap items-center justify-end gap-3">
              <Button variant="secondary" onClick={closeCreateModal} disabled={isCreating}>
                Cancel
              </Button>
              <Button onClick={handleCreate} disabled={isCreating}>
                {isCreating ? "Creating..." : "Create corpus"}
              </Button>
            </div>
          </div>
        </div>
      ) : null}
    </div>
  );
}
