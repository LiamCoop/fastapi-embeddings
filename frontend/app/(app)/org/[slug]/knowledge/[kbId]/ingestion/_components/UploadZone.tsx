"use client";

import { knowledgeDocumentsApiPath } from "@/app/lib/org-knowledge";
import { knowledgeDocumentChunkingApiPath } from "@/app/lib/org-knowledge";
import { useRouter } from "next/navigation";
import { useRef, useState } from "react";
import type { ChangeEvent, DragEvent } from "react";

type UploadZoneProps = {
  slug: string;
  kbId: string;
};

type UploadedDocumentResponse = {
  id: string;
};

export function UploadZone({ slug, kbId }: UploadZoneProps) {
  const router = useRouter();
  const inputRef = useRef<HTMLInputElement | null>(null);
  const [isFormOpen, setIsFormOpen] = useState(false);
  const [dragActive, setDragActive] = useState(false);
  const [selectedFile, setSelectedFile] = useState<File | null>(null);
  const [pathValue, setPathValue] = useState("");
  const [isUploading, setIsUploading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [successMessage, setSuccessMessage] = useState<string | null>(null);

  async function uploadFile(file: File, requestedPath: string) {
    setIsUploading(true);
    setError(null);
    setSuccessMessage(null);

    try {
      const normalizedPath = requestedPath.trim() || file.name;
      const formData = new FormData();
      formData.append("file", file);
      formData.append("path", normalizedPath);

      const res = await fetch(knowledgeDocumentsApiPath(slug, kbId), {
        method: "POST",
        body: formData,
      });

      if (!res.ok) {
        const body = (await res.json().catch(() => null)) as { error?: string } | null;
        throw new Error(body?.error ?? `Upload failed with status ${res.status}`);
      }

      const uploaded = (await res.json()) as UploadedDocumentResponse;
      const chunkingRes = await fetch(knowledgeDocumentChunkingApiPath(slug, kbId, uploaded.id), {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({}),
      });
      if (!chunkingRes.ok) {
        const body = (await chunkingRes.json().catch(() => null)) as { error?: string } | null;
        throw new Error(body?.error ?? `Chunking failed with status ${chunkingRes.status}`);
      }

      setSuccessMessage(`Uploaded ${normalizedPath}`);
      setSelectedFile(null);
      setPathValue("");
      router.refresh();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Upload failed");
    } finally {
      setIsUploading(false);
    }
  }

  function onFileSelected(file: File) {
    setSelectedFile(file);
    setError(null);
    setSuccessMessage(null);
    setIsFormOpen(true);
    setPathValue((current) => (current.trim().length > 0 ? current : file.name));
  }

  function onInputChange(event: ChangeEvent<HTMLInputElement>) {
    const nextFile = event.target.files?.[0];
    if (!nextFile) {
      return;
    }
    onFileSelected(nextFile);
    event.target.value = "";
  }

  function onDrop(event: DragEvent<HTMLDivElement>) {
    event.preventDefault();
    setDragActive(false);
    if (isUploading) {
      return;
    }
    const nextFile = event.dataTransfer.files?.[0];
    if (!nextFile) {
      return;
    }
    onFileSelected(nextFile);
  }

  function onSubmitUpload() {
    if (isUploading) {
      return;
    }
    if (!selectedFile) {
      setError("Select a file before uploading.");
      return;
    }
    void uploadFile(selectedFile, pathValue);
  }

  return (
    <div>
      <input
        ref={inputRef}
        type="file"
        className="hidden"
        onChange={onInputChange}
        disabled={isUploading}
      />

      <p className="text-xs uppercase tracking-[0.2em] text-muted-foreground">Upload Sources</p>

      {!isFormOpen ? (
        <div className="mt-4 space-y-3">
          <p className="text-sm text-muted-foreground/70">
            Add documents and set their ingestion path for retrieval filtering.
          </p>
          <button
            type="button"
            onClick={() => setIsFormOpen(true)}
            className="rounded-md border border-border px-3 py-2 text-xs font-medium uppercase tracking-[0.14em] text-foreground transition hover:bg-secondary"
          >
            Open Upload Form
          </button>
        </div>
      ) : (
        <div className="mt-4 space-y-4">
          <div className="flex items-center justify-between gap-3">
            <p className="text-xs uppercase tracking-[0.2em] text-muted-foreground">Upload Form</p>
            <button
              type="button"
              onClick={() => setIsFormOpen(false)}
              disabled={isUploading}
              className="text-xs uppercase tracking-[0.16em] text-muted-foreground transition hover:text-foreground disabled:cursor-not-allowed disabled:opacity-50"
            >
              Close
            </button>
          </div>

          <div className="space-y-2">
            <label
              htmlFor="document-path"
              className="text-[11px] font-semibold uppercase tracking-[0.2em] text-muted-foreground"
            >
              File path
            </label>
            <input
              id="document-path"
              value={pathValue}
              onChange={(event) => setPathValue(event.target.value)}
              placeholder="docs/architecture/overview.md"
              disabled={isUploading}
              className="h-10 w-full rounded-lg border border-input bg-background px-3 text-sm text-foreground outline-none transition focus:border-primary/60 focus:ring-2 focus:ring-primary/20 disabled:cursor-not-allowed disabled:opacity-60"
            />
          </div>

          <div
            role="button"
            tabIndex={0}
            onClick={() => {
              if (!isUploading) {
                inputRef.current?.click();
              }
            }}
            onKeyDown={(event) => {
              if ((event.key === "Enter" || event.key === " ") && !isUploading) {
                event.preventDefault();
                inputRef.current?.click();
              }
            }}
            onDragOver={(event) => {
              event.preventDefault();
              if (!isUploading) {
                setDragActive(true);
              }
            }}
            onDragLeave={(event) => {
              event.preventDefault();
              setDragActive(false);
            }}
            onDrop={onDrop}
            className={`rounded-lg border border-dashed py-10 text-center transition-colors ${
              dragActive ? "border-primary bg-primary/5" : "border-border"
            } ${isUploading ? "cursor-not-allowed opacity-70" : "cursor-pointer"}`}
          >
            {isUploading ? (
              <>
                <p className="text-sm text-muted-foreground">Uploading...</p>
                <p className="mt-1 text-xs text-muted-foreground/60">
                  Please wait while your file is stored.
                </p>
              </>
            ) : (
              <>
                <p className="text-sm text-muted-foreground/80">
                  Drag and drop files here, or click to choose a file
                </p>
                <p className="mt-1 text-xs text-muted-foreground/50">Markdown, PDF, and more</p>
                {selectedFile ? (
                  <p className="mt-3 text-xs text-foreground/80">Selected: {selectedFile.name}</p>
                ) : null}
              </>
            )}
          </div>

          <div className="flex items-center justify-end gap-2">
            <button
              type="button"
              onClick={() => inputRef.current?.click()}
              disabled={isUploading}
              className="rounded-md border border-border px-3 py-2 text-xs font-medium uppercase tracking-[0.14em] text-foreground transition hover:bg-secondary disabled:cursor-not-allowed disabled:opacity-60"
            >
              Choose File
            </button>
            <button
              type="button"
              onClick={onSubmitUpload}
              disabled={isUploading || !selectedFile}
              className="rounded-md bg-primary px-3 py-2 text-xs font-semibold uppercase tracking-[0.14em] text-primary-foreground transition hover:opacity-90 disabled:cursor-not-allowed disabled:opacity-60"
            >
              {isUploading ? "Uploading..." : "Upload File"}
            </button>
          </div>
        </div>
      )}

      {error ? <p className="mt-3 text-xs text-destructive">{error}</p> : null}
      {!error && successMessage ? <p className="mt-3 text-xs text-emerald-600">{successMessage}</p> : null}
    </div>
  );
}
