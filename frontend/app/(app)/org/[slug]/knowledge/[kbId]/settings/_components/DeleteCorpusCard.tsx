"use client";

import { deleteOrgKnowledgeBase } from "@/app/lib/org-knowledge";
import { Button } from "@/components/ui/button";
import { useRouter } from "next/navigation";
import { useState } from "react";

type DeleteCorpusCardProps = {
  slug: string;
  kbId: string;
  corpusName: string;
};

export function DeleteCorpusCard({ slug, kbId, corpusName }: DeleteCorpusCardProps) {
  const router = useRouter();
  const [isConfirming, setIsConfirming] = useState(false);
  const [isDeleting, setIsDeleting] = useState(false);
  const [confirmationText, setConfirmationText] = useState("");
  const [error, setError] = useState<string | null>(null);

  async function handleDelete() {
    setIsDeleting(true);
    setError(null);

    try {
      await deleteOrgKnowledgeBase(slug, kbId);
      router.push(`/org/${slug}/knowledge`);
      router.refresh();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Delete failed");
      setIsDeleting(false);
    }
  }

  const canConfirmDelete = confirmationText.trim() === corpusName;

  return (
    <div className="rounded-xl border border-border bg-card p-6 md:col-span-2">
      <p className="text-xs uppercase tracking-[0.2em] text-muted-foreground">Danger Zone</p>
      <div className="mt-4 rounded-lg border border-dashed border-destructive/40 p-5">
        <p className="text-sm font-medium text-foreground">
          Delete Corpus <span className="font-mono">"{corpusName}"</span>
        </p>
        <p className="mt-1 text-xs text-muted-foreground/60">
          This permanently deletes the corpus, all documents, chunks, embeddings, and stored source files.
        </p>

        {error ? <p className="mt-3 text-xs text-destructive">{error}</p> : null}

        {isConfirming ? (
          <div className="mt-4 space-y-3">
            <label htmlFor="delete-corpus-confirm" className="block text-xs text-muted-foreground">
              Type "{corpusName}" to confirm deletion.
            </label>
            <input
              id="delete-corpus-confirm"
              type="text"
              value={confirmationText}
              onChange={(event) => {
                setConfirmationText(event.target.value);
              }}
              placeholder={corpusName}
              className="h-9 w-full rounded-md border border-border bg-background px-3 text-sm text-foreground outline-none focus-visible:border-ring focus-visible:ring-ring/50 focus-visible:ring-[3px]"
              disabled={isDeleting}
            />
            <div className="flex items-center gap-2">
              <Button
                type="button"
                size="sm"
                variant="outline"
                disabled={isDeleting}
                onClick={() => {
                  setIsConfirming(false);
                  setConfirmationText("");
                  setError(null);
                }}
              >
                Cancel
              </Button>
              <Button
                type="button"
                size="sm"
                variant="destructive"
                disabled={isDeleting || !canConfirmDelete}
                onClick={() => {
                  void handleDelete();
                }}
              >
                {isDeleting ? "Deleting..." : "Delete corpus"}
              </Button>
            </div>
          </div>
        ) : (
          <div className="mt-4">
            <Button
              type="button"
              size="sm"
              variant="destructive"
              onClick={() => {
                setIsConfirming(true);
                setError(null);
              }}
            >
              Delete corpus
            </Button>
          </div>
        )}
      </div>
    </div>
  );
}
