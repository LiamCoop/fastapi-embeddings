"use client";

import { knowledgeDocumentChunkingApiPath, type IngestionDocument } from "@/app/lib/org-knowledge";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";
import { useRouter } from "next/navigation";
import { useState } from "react";

type DocumentsChunkListProps = {
  slug: string;
  kbId: string;
  documents: IngestionDocument[];
  chunkCountsByDocumentId: Record<string, number>;
  selectedDocumentId: string | null;
  onSelectDocument: (documentId: string) => void;
};

export function DocumentsChunkList({
  slug,
  kbId,
  documents,
  chunkCountsByDocumentId,
  selectedDocumentId,
  onSelectDocument,
}: DocumentsChunkListProps) {
  const router = useRouter();
  const [rechunkingDocumentId, setRechunkingDocumentId] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);

  async function handleRechunk(documentId: string) {
    setRechunkingDocumentId(documentId);
    setError(null);
    try {
      const res = await fetch(knowledgeDocumentChunkingApiPath(slug, kbId, documentId), {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({}),
      });
      if (!res.ok) {
        const payload = (await res.json().catch(() => null)) as { error?: string } | null;
        throw new Error(payload?.error ?? `Rechunk request failed with status ${res.status}`);
      }
      router.refresh();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to rechunk document");
    } finally {
      setRechunkingDocumentId(null);
    }
  }

  return (
    <>
      {error ? <p className="mt-3 text-xs text-destructive">{error}</p> : null}
      <div className="mt-4 overflow-hidden divide-y divide-border rounded-lg border border-border">
        {documents.map((doc) => {
          const isRechunking = rechunkingDocumentId === doc.id;
          const isSelected = selectedDocumentId === doc.id;
          return (
            <div
              key={doc.id}
              role="button"
              tabIndex={0}
              className={cn(
                "w-full px-3 py-2 text-left transition-colors hover:bg-muted/40",
                isSelected ? "bg-muted/60" : "",
              )}
              onClick={() => {
                onSelectDocument(doc.id);
              }}
              onKeyDown={(event) => {
                if (event.key === "Enter" || event.key === " ") {
                  event.preventDefault();
                  onSelectDocument(doc.id);
                }
              }}
            >
              <div className="grid grid-cols-[minmax(0,1fr)_auto] items-start gap-x-3 gap-y-1">
                <p className="min-w-0 truncate text-sm font-medium text-foreground">{doc.title?.trim() || doc.path}</p>
                <p className="shrink-0 whitespace-nowrap font-mono text-[10px] text-muted-foreground">
                  {chunkCountsByDocumentId[doc.id] ?? 0} chunks
                </p>
                <div className="min-w-0">
                  <p className="truncate text-xs text-muted-foreground">{doc.path}</p>
                  <p className="mt-1 font-mono text-[10px] uppercase tracking-[0.2em] text-muted-foreground/80">
                    {doc.processing_status ?? "UNKNOWN"}
                  </p>
                </div>
                <Button
                  type="button"
                  size="sm"
                  variant="outline"
                  className="shrink-0"
                  disabled={Boolean(rechunkingDocumentId)}
                  onClick={(event) => {
                    event.stopPropagation();
                    onSelectDocument(doc.id);
                    void handleRechunk(doc.id);
                  }}
                >
                  {isRechunking ? "Rechunking..." : "Rechunk"}
                </Button>
              </div>
            </div>
          );
        })}
      </div>
    </>
  );
}
