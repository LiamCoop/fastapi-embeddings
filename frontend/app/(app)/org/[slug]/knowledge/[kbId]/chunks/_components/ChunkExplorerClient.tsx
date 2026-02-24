"use client";

import {
  embedOrgKnowledgeChunk,
  fetchOrgKnowledgeChunksForDocument,
  knowledgeDocumentChunkingApiPath,
  type IngestionDocument,
  type KnowledgeChunk,
} from "@/app/lib/org-knowledge";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { usePathname, useRouter, useSearchParams } from "next/navigation";
import { useEffect, useMemo, useState } from "react";
import { DocumentsChunkList } from "./DocumentsChunkList";

type ChunkExplorerClientProps = {
  slug: string;
  kbId: string;
  documents: IngestionDocument[];
  chunkCountsByDocumentId: Record<string, number>;
};

type BatchChunkFailure = {
  path: string;
  reason: string;
};

type BatchChunkProgress = {
  total: number;
  completed: number;
  chunked: number;
  inFlight: number;
  currentPath: string | null;
  failures: BatchChunkFailure[];
};

const BATCH_CHUNK_CONCURRENCY = 4;

export function ChunkExplorerClient({ slug, kbId, documents, chunkCountsByDocumentId }: ChunkExplorerClientProps) {
  const router = useRouter();
  const pathname = usePathname();
  const searchParams = useSearchParams();
  const [selectedChunks, setSelectedChunks] = useState<KnowledgeChunk[]>([]);
  const [isLoadingChunks, setIsLoadingChunks] = useState(false);
  const [chunksError, setChunksError] = useState<string | null>(null);
  const [countsByDocumentId, setCountsByDocumentId] = useState<Record<string, number>>(chunkCountsByDocumentId);
  const [embeddingInFlight, setEmbeddingInFlight] = useState<Record<string, boolean>>({});
  const [isChunkingAll, setIsChunkingAll] = useState(false);
  const [batchChunkError, setBatchChunkError] = useState<string | null>(null);
  const [batchChunkSuccess, setBatchChunkSuccess] = useState<string | null>(null);
  const [batchChunkProgress, setBatchChunkProgress] = useState<BatchChunkProgress | null>(null);

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

  const handleChunkAll = async () => {
    if (isChunkingAll || documents.length === 0) {
      return;
    }

    setIsChunkingAll(true);
    setBatchChunkError(null);
    setBatchChunkSuccess(null);
    setBatchChunkProgress({
      total: documents.length,
      completed: 0,
      chunked: 0,
      inFlight: 0,
      currentPath: null,
      failures: [],
    });

    let nextIndex = 0;
    let chunkedCount = 0;
    const failures: BatchChunkFailure[] = [];
    const workerCount = Math.min(BATCH_CHUNK_CONCURRENCY, documents.length);

    async function worker() {
      while (nextIndex < documents.length) {
        const currentIndex = nextIndex;
        nextIndex += 1;
        const document = documents[currentIndex];
        if (!document) {
          continue;
        }

        const currentPath = document.path || document.title || document.id;
        setBatchChunkProgress((previous) => {
          if (!previous) {
            return previous;
          }
          return {
            ...previous,
            inFlight: previous.inFlight + 1,
            currentPath,
          };
        });

        try {
          const res = await fetch(knowledgeDocumentChunkingApiPath(slug, kbId, document.id), {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({}),
          });
          if (!res.ok) {
            const payload = (await res.json().catch(() => null)) as { error?: string } | null;
            throw new Error(payload?.error ?? `Chunking request failed with status ${res.status}`);
          }
          chunkedCount += 1;
          setBatchChunkProgress((previous) => {
            if (!previous) {
              return previous;
            }
            return {
              ...previous,
              completed: previous.completed + 1,
              chunked: previous.chunked + 1,
              inFlight: Math.max(0, previous.inFlight - 1),
            };
          });
        } catch (err) {
          const reason = err instanceof Error ? err.message : "Chunking failed";
          failures.push({ path: currentPath, reason });
          setBatchChunkProgress((previous) => {
            if (!previous) {
              return previous;
            }
            return {
              ...previous,
              completed: previous.completed + 1,
              inFlight: Math.max(0, previous.inFlight - 1),
              failures: [...previous.failures, { path: currentPath, reason }],
            };
          });
        }
      }
    }

    await Promise.all(Array.from({ length: workerCount }, () => worker()));
    router.refresh();

    if (failures.length > 0) {
      const preview = failures
        .slice(0, 3)
        .map((failure) => `${failure.path} (${failure.reason})`)
        .join("; ");
      const extra = failures.length > 3 ? ` (+${failures.length - 3} more)` : "";
      setBatchChunkError(`Failed to chunk ${failures.length} document(s): ${preview}${extra}`);
    }

    setBatchChunkSuccess(`Chunked ${chunkedCount} of ${documents.length} document(s).`);
    setIsChunkingAll(false);
  };

  return (
    <div className="grid grid-cols-1 gap-4 lg:grid-cols-[280px_1fr]">
      <div className="rounded-xl border border-border bg-card p-5">
        <div className="flex items-center justify-between gap-3">
          <p className="text-xs uppercase tracking-[0.2em] text-muted-foreground">Documents</p>
          <div className="flex items-center gap-2">
            <p className="font-mono text-xs text-muted-foreground">{documents.length}</p>
            <Button
              type="button"
              size="sm"
              variant="outline"
              disabled={isChunkingAll || documents.length === 0}
              onClick={() => void handleChunkAll()}
            >
              {isChunkingAll ? "Chunking..." : "Chunk all"}
            </Button>
          </div>
        </div>
        {batchChunkError ? <p className="mt-3 text-xs text-destructive">{batchChunkError}</p> : null}
        {batchChunkSuccess ? <p className="mt-3 text-xs text-muted-foreground">{batchChunkSuccess}</p> : null}
        {batchChunkProgress ? (
          <div className="mt-3 space-y-2 rounded-lg border border-border bg-secondary/30 p-3">
            <div className="flex items-center justify-between gap-3 text-xs text-muted-foreground">
              <p>
                Chunking {batchChunkProgress.completed}/{batchChunkProgress.total}
              </p>
              <p>{batchChunkProgress.inFlight} in flight</p>
            </div>
            <div className="h-1.5 w-full overflow-hidden rounded bg-muted">
              <div
                className="h-full rounded bg-primary transition-all"
                style={{
                  width: `${batchChunkProgress.total === 0 ? 0 : (batchChunkProgress.completed / batchChunkProgress.total) * 100}%`,
                }}
              />
            </div>
            <p className="text-xs text-muted-foreground">
              {batchChunkProgress.currentPath
                ? `Current: ${batchChunkProgress.currentPath}`
                : "Preparing chunk jobs..."}
            </p>
            {batchChunkProgress.failures.length > 0 ? (
              <p className="text-xs text-destructive">Failures: {batchChunkProgress.failures.length}</p>
            ) : null}
          </div>
        ) : null}

        {documents.length === 0 ? (
          <div className="mt-4 rounded-lg border border-dashed border-border py-10 text-center">
            <p className="text-xs text-muted-foreground/60">No documents</p>
          </div>
        ) : (
          <DocumentsChunkList
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
