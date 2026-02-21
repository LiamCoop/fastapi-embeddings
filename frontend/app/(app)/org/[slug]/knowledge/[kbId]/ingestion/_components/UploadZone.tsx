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
  const [dragActive, setDragActive] = useState(false);
  const [isUploading, setIsUploading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [successMessage, setSuccessMessage] = useState<string | null>(null);

  async function uploadFile(file: File) {
    setIsUploading(true);
    setError(null);
    setSuccessMessage(null);

    try {
      const formData = new FormData();
      formData.append("file", file);

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
        body: JSON.stringify({ strategy: "fixed" }),
      });
      if (!chunkingRes.ok) {
        const body = (await chunkingRes.json().catch(() => null)) as { error?: string } | null;
        throw new Error(body?.error ?? `Chunking failed with status ${chunkingRes.status}`);
      }

      setSuccessMessage(`Uploaded ${file.name}`);
      router.refresh();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Upload failed");
    } finally {
      setIsUploading(false);
    }
  }

  function onInputChange(event: ChangeEvent<HTMLInputElement>) {
    const nextFile = event.target.files?.[0];
    if (!nextFile) {
      return;
    }
    void uploadFile(nextFile);
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
    void uploadFile(nextFile);
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
        className={`mt-4 rounded-lg border border-dashed py-10 text-center transition-colors ${
          dragActive ? "border-primary bg-primary/5" : "border-border"
        } ${isUploading ? "cursor-not-allowed opacity-70" : "cursor-pointer"}`}
      >
        {isUploading ? (
          <>
            <p className="text-sm text-muted-foreground">Uploading...</p>
            <p className="mt-1 text-xs text-muted-foreground/60">Please wait while your file is stored.</p>
          </>
        ) : (
          <>
            <p className="text-sm text-muted-foreground/80">Drag and drop files here, or click to upload</p>
            <p className="mt-1 text-xs text-muted-foreground/50">Markdown, PDF, and more</p>
          </>
        )}
      </div>

      {error ? <p className="mt-3 text-xs text-destructive">{error}</p> : null}
      {!error && successMessage ? <p className="mt-3 text-xs text-emerald-600">{successMessage}</p> : null}
    </div>
  );
}
