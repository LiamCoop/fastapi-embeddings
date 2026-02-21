"use client";

import {
  embedOrgKnowledgeChunk,
  fetchOrgKnowledgeChunksForDocument,
  type IngestionDocument,
  type KnowledgeChunk,
} from "@/app/lib/org-knowledge";
import { Badge } from "@/components/ui/badge";
import { usePathname, useRouter, useSearchParams } from "next/navigation";
import { useEffect, useMemo, useState } from "react";
import { DocumentsChunkList } from "./DocumentsChunkList";

type ChunkExplorerClientProps = {
  slug: string;
  kbId: string;
  documents: IngestionDocument[];
  chunkCountsByDocumentId: Record<string, number>;
};

export function ChunkExplorerClient({ slug, kbId, documents, chunkCountsByDocumentId }: ChunkExplorerClientProps) {
  const router = useRouter();
  const pathname = usePathname();
  const searchParams = useSearchParams();
  const [selectedChunks, setSelectedChunks] = useState<KnowledgeChunk[]>([]);
  const [isLoadingChunks, setIsLoadingChunks] = useState(false);
  const [chunksError, setChunksError] = useState<string | null>(null);
  const [countsByDocumentId, setCountsByDocumentId] = useState<Record<string, number>>(chunkCountsByDocumentId);
  const [embeddingInFlight, setEmbeddingInFlight] = useState<Record<string, boolean>>({});

  useEffect(() => {
    setCountsByDocumentId(chunkCountsByDocumentId);
  }, [chunkCountsByDocumentId]);

  const selectedFromUrl = searchParams.get("documentId");
  const selectedDocumentId = selectedFromUrl && documents.some((document) => document.id === selectedFromUrl)
    ? selectedFromUrl
    : (documents[0]?.id ?? null);

  const selectedDocument = useMemo(
    () => documents.find((document) => document.id === selectedDocumentId) ?? null,
    [documents, selectedDocumentId],
  );

  useEffect(() => {
    if (!selectedDocument) {
      setSelectedChunks([]);
      setChunksError(null);
      setIsLoadingChunks(false);
      return;
    }

    const controller = new AbortController();
    setIsLoadingChunks(true);
    setChunksError(null);

    void fetchOrgKnowledgeChunksForDocument(slug, kbId, selectedDocument.id, {
      init: { cache: "no-store", signal: controller.signal },
    })
      .then((res) => {
        setSelectedChunks(res.chunks);
        setCountsByDocumentId((prev) => ({
          ...prev,
          [selectedDocument.id]: res.chunk_count,
        }));
      })
      .catch((err: unknown) => {
        if (controller.signal.aborted) {
          return;
        }
        setSelectedChunks([]);
        setChunksError(err instanceof Error ? err.message : "Failed to load chunks");
      })
      .finally(() => {
        if (!controller.signal.aborted) {
          setIsLoadingChunks(false);
        }
      });

    return () => {
      controller.abort();
    };
  }, [kbId, selectedDocument, slug]);

  const handleSelectDocument = (documentId: string) => {
    const nextParams = new URLSearchParams(searchParams.toString());
    nextParams.set("documentId", documentId);
    const nextUrl = `${pathname}?${nextParams.toString()}`;
    router.replace(nextUrl, { scroll: false });
  };

  const handleEmbedChunk = async (chunkId: string) => {
    if (embeddingInFlight[chunkId]) {
      return;
    }
    setEmbeddingInFlight((prev) => ({ ...prev, [chunkId]: true }));
    setChunksError(null);
    try {
      const res = await embedOrgKnowledgeChunk(slug, kbId, chunkId);
      setSelectedChunks((prev) =>
        prev.map((chunk) =>
          chunk.id === chunkId
            ? {
                ...chunk,
                embedding_id: res.embedding_id,
              }
            : chunk,
        ),
      );
    } catch (err) {
      setChunksError(err instanceof Error ? err.message : "Failed to embed chunk");
    } finally {
      setEmbeddingInFlight((prev) => ({ ...prev, [chunkId]: false }));
    }
  };

  return (
    <div className="grid grid-cols-1 gap-4 lg:grid-cols-[280px_1fr]">
      <div className="rounded-xl border border-border bg-card p-5">
        <div className="flex items-center justify-between gap-3">
          <p className="text-xs uppercase tracking-[0.2em] text-muted-foreground">Documents</p>
          <p className="font-mono text-xs text-muted-foreground">{documents.length}</p>
        </div>

        {documents.length === 0 ? (
          <div className="mt-4 rounded-lg border border-dashed border-border py-10 text-center">
            <p className="text-xs text-muted-foreground/60">No documents</p>
          </div>
        ) : (
          <DocumentsChunkList
            slug={slug}
            kbId={kbId}
            documents={documents}
            chunkCountsByDocumentId={countsByDocumentId}
            selectedDocumentId={selectedDocumentId}
            onSelectDocument={handleSelectDocument}
          />
        )}
      </div>

      <div className="space-y-4">
        <div className="rounded-xl border border-border bg-card p-6">
          <div className="flex items-center justify-between gap-3">
            <p className="text-xs uppercase tracking-[0.2em] text-muted-foreground">Chunk Visualization</p>
            {selectedDocument ? (
              <Badge variant="outline">{countsByDocumentId[selectedDocument.id] ?? 0} chunks</Badge>
            ) : null}
          </div>
          {!selectedDocument ? (
            <div className="mt-4 rounded-lg border border-dashed border-border py-16 text-center">
              <p className="text-sm text-muted-foreground/60">Select a document to explore its chunks</p>
            </div>
          ) : isLoadingChunks ? (
            <div className="mt-4 rounded-lg border border-dashed border-border py-16 text-center">
              <p className="text-sm text-muted-foreground/60">Loading chunks...</p>
            </div>
          ) : chunksError ? (
            <div className="mt-4 rounded-lg border border-dashed border-destructive/40 py-8 text-center">
              <p className="text-sm text-destructive">{chunksError}</p>
            </div>
          ) : selectedChunks.length === 0 ? (
            <div className="mt-4 rounded-lg border border-dashed border-border py-16 text-center">
              <p className="text-sm text-muted-foreground/60">No chunks found for this document</p>
            </div>
          ) : (
            <div className="mt-4 space-y-3">
              {selectedChunks.map((chunk) => (
                <div key={chunk.id} className="rounded-lg border border-border bg-background p-4">
                  <div className="flex items-center justify-between gap-3">
                    <div className="flex items-center gap-2">
                      <p className="font-mono text-[11px] text-muted-foreground">Chunk #{chunk.sequence_number}</p>
                      <Badge variant="outline" className="font-mono text-[10px] uppercase">
                        {chunk.chunking_strategy}
                      </Badge>
                    </div>
                    <button
                      type="button"
                      onClick={() => void handleEmbedChunk(chunk.id)}
                      title="embed?"
                      disabled={embeddingInFlight[chunk.id]}
                      className={`inline-flex h-6 min-w-16 items-center justify-center gap-1 rounded-md border px-2 text-[10px] font-mono transition-colors ${
                        chunk.embedding_id
                          ? "border-emerald-500/60 bg-emerald-500/15 text-emerald-700"
                          : "border-border bg-muted text-muted-foreground hover:bg-muted/80"
                      } ${embeddingInFlight[chunk.id] ? "opacity-60" : ""}`}
                    >
                      {embeddingInFlight[chunk.id] ? "..." : chunk.embedding_id ? "✓ EMBEDDED" : "EMBED"}
                    </button>
                  </div>
                  <pre className="mt-3 whitespace-pre-wrap break-words text-xs text-foreground">{chunk.content}</pre>
                </div>
              ))}
            </div>
          )}
        </div>

        <div className="rounded-xl border border-border bg-card p-6">
          <p className="text-xs uppercase tracking-[0.2em] text-muted-foreground">Metadata</p>
          {!selectedDocument ? (
            <div className="mt-4 rounded-lg border border-dashed border-border py-10 text-center">
              <p className="text-sm text-muted-foreground/60">No document selected</p>
            </div>
          ) : (
            <div className="mt-4 space-y-2 text-xs text-muted-foreground">
              <p className="truncate">Path: {selectedDocument.path}</p>
              <p>Status: {selectedDocument.processing_status ?? "UNKNOWN"}</p>
              <p>Version: {selectedDocument.version_number ?? "n/a"}</p>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
