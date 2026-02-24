"use client";

import { knowledgeDocumentApiPath, type IngestionDocument } from "@/app/lib/org-knowledge";
import { formatUtcDateTime } from "@/app/lib/datetime";
import { Button } from "@/components/ui/button";
import { useRouter } from "next/navigation";
import { useState } from "react";

type DocumentsListProps = {
  slug: string;
  kbId: string;
  documents: IngestionDocument[];
};

export function DocumentsList({ slug, kbId, documents }: DocumentsListProps) {
  const router = useRouter();
  const [deletingId, setDeletingId] = useState<string | null>(null);
  const [confirmDeleteId, setConfirmDeleteId] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);

  async function handleDelete(documentId: string) {
    setDeletingId(documentId);
    setConfirmDeleteId(null);
    setError(null);

    try {
      const res = await fetch(knowledgeDocumentApiPath(slug, kbId, documentId), {
        method: "DELETE",
      });

      if (!res.ok) {
        const body = (await res.json().catch(() => null)) as { error?: string } | null;
        throw new Error(body?.error ?? `Delete failed with status ${res.status}`);
      }

      router.refresh();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Delete failed");
    } finally {
      setDeletingId(null);
    }
  }

  if (documents.length === 0) {
    return (
      <div className="mt-4 rounded-lg border border-dashed border-border py-10 text-center">
        <p className="text-sm text-muted-foreground/60">No documents ingested yet</p>
        <p className="mt-1 text-xs text-muted-foreground/40">
          Upload stores the document version. Run manual chunking from the Chunks page via Rechunk.
        </p>
      </div>
    );
  }

  return (
    <div className="mt-4 space-y-3">
      {error ? <p className="text-xs text-destructive">{error}</p> : null}
      <div className="divide-y divide-border rounded-lg border border-border">
        {documents.map((doc) => {
          const isDeleting = deletingId === doc.id;
          const isConfirming = confirmDeleteId === doc.id;

          return (
            <div key={doc.id} className="flex flex-wrap items-start justify-between gap-3 px-4 py-3">
              <div className="min-w-0">
                <p className="truncate text-sm font-medium text-foreground">{doc.title?.trim() || doc.path}</p>
                <p className="truncate text-xs text-muted-foreground">{doc.path}</p>
              </div>
              <div className="flex items-center gap-3">
                <div className="text-right">
                  <p className="font-mono text-[10px] uppercase tracking-[0.2em] text-muted-foreground">
                    {doc.processing_status ?? "UNKNOWN"}
                  </p>
                  <p className="text-xs text-muted-foreground">
                    Updated {formatUtcDateTime(doc.updated_at)}
                  </p>
                </div>
                {isConfirming ? (
                  <div className="flex items-center gap-2">
                    <Button
                      type="button"
                      size="sm"
                      variant="outline"
                      disabled={Boolean(deletingId)}
                      onClick={() => {
                        setConfirmDeleteId(null);
                      }}
                    >
                      Cancel
                    </Button>
                    <Button
                      type="button"
                      size="sm"
                      variant="destructive"
                      disabled={Boolean(deletingId)}
                      onClick={() => {
                        void handleDelete(doc.id);
                      }}
                    >
                      {isDeleting ? "Deleting..." : "Confirm"}
                    </Button>
                  </div>
                ) : (
                  <Button
                    type="button"
                    size="sm"
                    variant="destructive"
                    disabled={Boolean(deletingId)}
                    onClick={() => {
                      setConfirmDeleteId(doc.id);
                    }}
                  >
                    Delete
                  </Button>
                )}
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
}
