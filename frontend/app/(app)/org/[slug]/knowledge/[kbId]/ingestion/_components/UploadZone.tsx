"use client";

import { knowledgeDocumentsApiPath } from "@/app/lib/org-knowledge";
import { useRouter } from "next/navigation";
import { useRef, useState } from "react";
import type { ChangeEvent, DragEvent, InputHTMLAttributes } from "react";

type UploadZoneProps = {
  slug: string;
  kbId: string;
};

export function UploadZone({ slug, kbId }: UploadZoneProps) {
  const router = useRouter();
  const fileInputRef = useRef<HTMLInputElement | null>(null);
  const directoryInputRef = useRef<HTMLInputElement | null>(null);
  const [isFormOpen, setIsFormOpen] = useState(false);
  const [dragActive, setDragActive] = useState(false);
  const [selectedFile, setSelectedFile] = useState<File | null>(null);
  const [pathValue, setPathValue] = useState("");
  const [isUploading, setIsUploading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [successMessage, setSuccessMessage] = useState<string | null>(null);

  const directoryUploadInputProps = {
    webkitdirectory: "",
    directory: "",
  } as unknown as InputHTMLAttributes<HTMLInputElement>;

  function isMarkdownFile(file: File): boolean {
    const lower = file.name.toLowerCase();
    return lower.endsWith(".md") || lower.endsWith(".mdx");
  }

  function directoryRelativePath(file: File): string {
    const raw = (file as File & { webkitRelativePath?: string }).webkitRelativePath ?? file.name;
    const normalized = raw.replaceAll("\\", "/").trim();
    if (!normalized) {
      return file.name;
    }

    const segments = normalized.split("/").filter(Boolean);
    if (segments.length <= 1) {
      return segments[0] ?? file.name;
    }
    return segments.slice(1).join("/");
  }

  async function uploadOne(file: File, requestedPath: string): Promise<void> {
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
  }

  async function uploadFile(file: File, requestedPath: string) {
    setIsUploading(true);
    setError(null);
    setSuccessMessage(null);

    try {
      await uploadOne(file, requestedPath);
      router.refresh();
      setSuccessMessage(`Uploaded and stored ${requestedPath.trim() || file.name}`);
      setSelectedFile(null);
      setPathValue("");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Upload failed");
    } finally {
      setIsUploading(false);
    }
  }

  async function uploadDirectory(files: FileList) {
    setIsUploading(true);
    setError(null);
    setSuccessMessage(null);

    const allFiles = Array.from(files);
    const markdownFiles = allFiles.filter(isMarkdownFile);
    const skippedCount = allFiles.length - markdownFiles.length;

    if (markdownFiles.length === 0) {
      setIsUploading(false);
      setError("No markdown files found in selected directory.");
      return;
    }

    let uploadedCount = 0;
    const failed: string[] = [];

    for (const file of markdownFiles) {
      const relativePath = directoryRelativePath(file);
      try {
        await uploadOne(file, relativePath);
        uploadedCount++;
      } catch (err) {
        const reason = err instanceof Error ? err.message : "Upload failed";
        failed.push(`${relativePath} (${reason})`);
      }
    }

    router.refresh();

    if (failed.length > 0) {
      const preview = failed.slice(0, 3).join("; ");
      const extra = failed.length > 3 ? ` (+${failed.length - 3} more)` : "";
      setError(`Failed to upload ${failed.length} markdown file(s): ${preview}${extra}`);
    }

    setSuccessMessage(
      `Uploaded and stored ${uploadedCount} markdown file(s). Skipped ${skippedCount} non-markdown file(s).`,
    );
    setIsUploading(false);
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

  function onDirectoryInputChange(event: ChangeEvent<HTMLInputElement>) {
    const files = event.target.files;
    if (!files || files.length === 0 || isUploading) {
      return;
    }
    void uploadDirectory(files);
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
        ref={fileInputRef}
        type="file"
        className="hidden"
        onChange={onInputChange}
        disabled={isUploading}
      />
      <input
        ref={directoryInputRef}
        type="file"
        multiple
        className="hidden"
        onChange={onDirectoryInputChange}
        disabled={isUploading}
        {...directoryUploadInputProps}
      />

      <p className="text-xs uppercase tracking-[0.2em] text-muted-foreground">Upload Sources</p>

      {!isFormOpen ? (
        <div className="mt-4 space-y-3">
          <p className="text-sm text-muted-foreground/70">
            Add documents and set their ingestion path for retrieval filtering. Chunking runs manually from the
            Chunks page.
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
                fileInputRef.current?.click();
              }
            }}
            onKeyDown={(event) => {
              if ((event.key === "Enter" || event.key === " ") && !isUploading) {
                event.preventDefault();
                fileInputRef.current?.click();
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
              onClick={() => fileInputRef.current?.click()}
              disabled={isUploading}
              className="rounded-md border border-border px-3 py-2 text-xs font-medium uppercase tracking-[0.14em] text-foreground transition hover:bg-secondary disabled:cursor-not-allowed disabled:opacity-60"
            >
              Choose File
            </button>
            <button
              type="button"
              onClick={() => directoryInputRef.current?.click()}
              disabled={isUploading}
              className="rounded-md border border-border px-3 py-2 text-xs font-medium uppercase tracking-[0.14em] text-foreground transition hover:bg-secondary disabled:cursor-not-allowed disabled:opacity-60"
            >
              Choose Folder
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
