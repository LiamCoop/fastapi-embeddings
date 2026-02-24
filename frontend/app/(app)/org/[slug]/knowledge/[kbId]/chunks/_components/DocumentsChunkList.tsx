"use client";

import type { IngestionDocument } from "@/app/lib/org-knowledge";
import { cn } from "@/lib/utils";

type DocumentsChunkListProps = {
  documents: IngestionDocument[];
  chunkCountsByDocumentId: Record<string, number>;
  selectedDocumentId: string | null;
  onSelectDocument: (documentId: string) => void;
};

export function DocumentsChunkList({
  documents,
  chunkCountsByDocumentId,
  selectedDocumentId,
  onSelectDocument,
}: DocumentsChunkListProps) {
  return (
    <>
      <div className="mt-4 overflow-hidden divide-y divide-border rounded-lg border border-border">
        {documents.map((doc) => {
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
              </div>
            </div>
          );
        })}
      </div>
    </>
  );
}
