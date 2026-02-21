"use client";

import {
  retrieveOrgKnowledge,
  type KnowledgeRetrievalRequest,
  type KnowledgeRetrievalResponse,
} from "@/app/lib/org-knowledge";
import { Button } from "@/components/ui/button";
import { useState } from "react";

type RetrievalQueryClientProps = {
  slug: string;
  kbId: string;
};

function toErrorMessage(value: unknown): string {
  if (!(value instanceof Error)) {
    return "Retrieval failed";
  }

  try {
    const parsed = JSON.parse(value.message) as { error?: string };
    if (parsed?.error) {
      return parsed.error;
    }
  } catch {
    // Not a JSON error payload; use the original message below.
  }

  return value.message || "Retrieval failed";
}

function formatScore(value: number): string {
  return value.toFixed(3);
}

export function RetrievalQueryClient({ slug, kbId }: RetrievalQueryClientProps) {
  const [query, setQuery] = useState("");
  const [topK, setTopK] = useState(5);
  const [hybridWeight, setHybridWeight] = useState(0.7);
  const [pathPrefix, setPathPrefix] = useState("");
  const [documentType, setDocumentType] = useState("");
  const [source, setSource] = useState("");
  const [tagsInput, setTagsInput] = useState("");
  const [createdAfter, setCreatedAfter] = useState("");
  const [createdBefore, setCreatedBefore] = useState("");
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [response, setResponse] = useState<KnowledgeRetrievalResponse | null>(null);

  const runQuery = async () => {
    const nextQuery = query.trim();
    if (!nextQuery || isLoading) {
      return;
    }

    setIsLoading(true);
    setError(null);
    setResponse(null);
    try {
      const tags = tagsInput
        .split(",")
        .map((tag) => tag.trim())
        .filter((tag) => tag.length > 0);
      const filters: KnowledgeRetrievalRequest["filters"] = {
        path_prefix: pathPrefix.trim() || undefined,
        document_type: documentType.trim() || undefined,
        source: source.trim() || undefined,
        tags: tags.length > 0 ? tags : undefined,
        created_after: createdAfter.trim() || undefined,
        created_before: createdBefore.trim() || undefined,
      };
      const hasFilters = Object.values(filters).some((value) => {
        if (Array.isArray(value)) {
          return value.length > 0;
        }
        return typeof value === "string" ? value.trim().length > 0 : Boolean(value);
      });

      const payload: KnowledgeRetrievalRequest = {
        query: nextQuery,
        top_k: topK,
        hybrid_weight: hybridWeight,
        filters: hasFilters ? filters : undefined,
      };
      const result = await retrieveOrgKnowledge(slug, kbId, payload, {
        init: { cache: "no-store" },
      });
      setResponse(result);
    } catch (err) {
      setError(toErrorMessage(err));
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="grid grid-cols-1 gap-6 xl:grid-cols-[minmax(0,1.15fr)_minmax(0,1fr)] xl:items-start">
      <div className="rounded-xl border border-border bg-card p-6 xl:sticky xl:top-6">
        <p className="text-xs uppercase tracking-[0.2em] text-muted-foreground">Query</p>
        <p className="mt-2 text-sm text-muted-foreground">
          Enter a query and run hybrid retrieval against active document versions in this knowledge
          base.
        </p>
        <div className="mt-3 rounded-lg border border-dashed border-border bg-background/30 p-3 text-xs text-muted-foreground">
          <p className="font-medium text-foreground">How this works</p>
          <p className="mt-1">
            We convert your query into an embedding, combine semantic and keyword search, apply your
            filters, and return ranked chunks with citations.
          </p>
        </div>

        <div className="mt-4 grid grid-cols-1 gap-3 md:grid-cols-2">
          <label className="rounded-lg border border-border bg-background px-3 py-2 text-xs text-muted-foreground md:col-span-2">
            Query
            <input
              value={query}
              onChange={(event) => setQuery(event.target.value)}
              onKeyDown={(event) => {
                if (event.key === "Enter") {
                  event.preventDefault();
                  void runQuery();
                }
              }}
              placeholder="What does this knowledge base say about...?"
              className="mt-1 block w-full bg-transparent text-sm text-foreground outline-none"
            />
            <p className="mt-1 text-[11px] text-muted-foreground/70">
              Ask in natural language. Example: &quot;How do we handle activation failures?&quot;
            </p>
          </label>

          <label className="rounded-lg border border-border bg-background px-3 py-2 text-xs text-muted-foreground">
            Top-K
            <input
              type="number"
              min={1}
              max={50}
              value={topK}
              onChange={(event) => {
                const value = Number(event.target.value);
                if (!Number.isNaN(value)) {
                  setTopK(Math.max(1, Math.min(50, value)));
                }
              }}
              className="mt-1 block w-full bg-transparent text-sm text-foreground outline-none"
            />
            <p className="mt-1 text-[11px] text-muted-foreground/70">
              Number of chunks to return. Higher values give broader context.
            </p>
          </label>

          <label className="rounded-lg border border-border bg-background px-3 py-2 text-xs text-muted-foreground">
            Hybrid Weight
            <input
              type="number"
              min={0}
              max={1}
              step={0.1}
              value={hybridWeight}
              onChange={(event) => {
                const value = Number(event.target.value);
                if (!Number.isNaN(value)) {
                  setHybridWeight(Math.max(0, Math.min(1, value)));
                }
              }}
              className="mt-1 block w-full bg-transparent text-sm text-foreground outline-none"
            />
            <p className="mt-1 text-[11px] text-muted-foreground/70">
              1.0 favors semantic meaning, 0.0 favors exact keyword matches.
            </p>
          </label>

          <label className="rounded-lg border border-border bg-background px-3 py-2 text-xs text-muted-foreground">
            Path Prefix
            <input
              value={pathPrefix}
              onChange={(event) => setPathPrefix(event.target.value)}
              placeholder="docs/architecture/"
              className="mt-1 block w-full bg-transparent text-sm text-foreground outline-none"
            />
            <p className="mt-1 text-[11px] text-muted-foreground/70">
              Only search documents whose path starts with this value.
            </p>
          </label>

          <label className="rounded-lg border border-border bg-background px-3 py-2 text-xs text-muted-foreground">
            Document Type
            <input
              value={documentType}
              onChange={(event) => setDocumentType(event.target.value)}
              placeholder="text/markdown"
              className="mt-1 block w-full bg-transparent text-sm text-foreground outline-none"
            />
            <p className="mt-1 text-[11px] text-muted-foreground/70">
              Restrict by content type, like markdown or pdf.
            </p>
          </label>
          <label className="rounded-lg border border-border bg-background px-3 py-2 text-xs text-muted-foreground">
            Source
            <input
              value={source}
              onChange={(event) => setSource(event.target.value)}
              placeholder="github"
              className="mt-1 block w-full bg-transparent text-sm text-foreground outline-none"
            />
            <p className="mt-1 text-[11px] text-muted-foreground/70">
              Match a source label from document metadata.
            </p>
          </label>
          <label className="rounded-lg border border-border bg-background px-3 py-2 text-xs text-muted-foreground md:col-span-2">
            Tags
            <input
              value={tagsInput}
              onChange={(event) => setTagsInput(event.target.value)}
              placeholder="payments, incidents, onboarding"
              className="mt-1 block w-full bg-transparent text-sm text-foreground outline-none"
            />
            <p className="mt-1 text-[11px] text-muted-foreground/70">
              Comma-separated. Use tags to target a topic or document slice.
            </p>
          </label>
          <label className="rounded-lg border border-border bg-background px-3 py-2 text-xs text-muted-foreground">
            Created After
            <input
              value={createdAfter}
              onChange={(event) => setCreatedAfter(event.target.value)}
              placeholder="2026-01-01T00:00:00Z"
              className="mt-1 block w-full bg-transparent text-sm text-foreground outline-none"
            />
            <p className="mt-1 text-[11px] text-muted-foreground/70">
              Only include versions created at or after this timestamp (RFC3339).
            </p>
          </label>
          <label className="rounded-lg border border-border bg-background px-3 py-2 text-xs text-muted-foreground">
            Created Before
            <input
              value={createdBefore}
              onChange={(event) => setCreatedBefore(event.target.value)}
              placeholder="2026-02-21T00:00:00Z"
              className="mt-1 block w-full bg-transparent text-sm text-foreground outline-none"
            />
            <p className="mt-1 text-[11px] text-muted-foreground/70">
              Only include versions created at or before this timestamp (RFC3339).
            </p>
          </label>
        </div>

        <div className="mt-3 flex items-center gap-3">
          <Button onClick={() => void runQuery()} disabled={isLoading || query.trim().length === 0}>
            {isLoading ? "Running..." : "Run Retrieval"}
          </Button>
          {response ? (
            <p className="text-xs text-muted-foreground">
              request_id: <span className="font-mono">{response.request_id}</span> ·{" "}
              {response.result_count} results · {response.latency_ms} ms
            </p>
          ) : null}
        </div>

        {error ? <p className="mt-3 text-xs text-destructive">{error}</p> : null}
      </div>

      <div className="rounded-xl border border-border bg-card p-6">
        <p className="text-xs uppercase tracking-[0.2em] text-muted-foreground">Results</p>
        {!response ? (
          <div className="mt-4 rounded-lg border border-dashed border-border py-12 text-center">
            <p className="text-sm text-muted-foreground/60">Run a query to see retrieved chunks</p>
          </div>
        ) : response.results.length === 0 ? (
          <div className="mt-4 rounded-lg border border-dashed border-border py-12 text-center">
            <p className="text-sm text-muted-foreground/60">No chunks matched this query</p>
          </div>
        ) : (
          <div className="mt-4 space-y-3">
            {response.results.map((result, index) => (
              <div key={result.chunk_id} className="rounded-lg border border-border bg-background p-4">
                <div className="flex flex-wrap items-center gap-2 text-xs">
                  <span className="rounded border border-border px-2 py-1 font-mono">#{index + 1}</span>
                  <span className="rounded border border-border px-2 py-1 font-mono">
                    final {formatScore(result.scores.final)}
                  </span>
                  <span className="rounded border border-border px-2 py-1 font-mono">
                    semantic {formatScore(result.scores.semantic)}
                  </span>
                  <span className="rounded border border-border px-2 py-1 font-mono">
                    lexical {formatScore(result.scores.lexical)}
                  </span>
                </div>

                <p className="mt-3 text-sm text-foreground">{result.content}</p>

                <div className="mt-3 space-y-1 text-xs text-muted-foreground">
                  <p>
                    {result.document_title ?? result.citation.path} · {result.document_type}
                  </p>
                  <p className="font-mono">
                    doc {result.citation.document_id} · version {result.citation.version_number} ·
                    chunk {result.citation.chunk_sequence}
                  </p>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
