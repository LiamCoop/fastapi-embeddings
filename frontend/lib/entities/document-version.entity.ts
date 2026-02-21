import { Entity, PrimaryColumn, Column, CreateDateColumn } from "typeorm";

@Entity("document_versions")
export class DocumentVersion {
  @PrimaryColumn({ type: "uuid" })
  id!: string;

  @Column({ type: "uuid", name: "document_id" })
  documentId!: string;

  @Column({ type: "uuid", name: "kb_id" })
  kbId!: string;

  @Column({ type: "integer", name: "version_number" })
  versionNumber!: number;

  @Column({ type: "text", name: "raw_content_uri" })
  rawContentUri!: string;

  @Column({ type: "text", name: "extracted_content", nullable: true })
  extractedContent!: string | null;

  @Column({ type: "text", name: "extracted_content_hash", nullable: true })
  extractedContentHash!: string | null;

  @Column({ type: "text", name: "processing_status" })
  processingStatus!: string;

  @Column({ type: "text", name: "error_message", nullable: true })
  errorMessage!: string | null;

  @Column({ type: "boolean", name: "is_active" })
  isActive!: boolean;

  @CreateDateColumn({ type: "timestamptz", name: "created_at" })
  createdAt!: Date;
}
