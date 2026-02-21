import {
  Entity,
  PrimaryColumn,
  Column,
  CreateDateColumn,
  UpdateDateColumn,
} from "typeorm";

@Entity("documents")
export class Document {
  @PrimaryColumn({ type: "uuid" })
  id!: string;

  @Column({ type: "uuid", name: "kb_id" })
  kbId!: string;

  @Column({ type: "text" })
  path!: string;

  @Column({ type: "text", nullable: true })
  title!: string | null;

  @Column({ type: "text", name: "document_type" })
  documentType!: string;

  @Column({ type: "jsonb", name: "source_metadata", default: {} })
  sourceMetadata!: Record<string, unknown>;

  @Column({ type: "uuid", name: "active_version_id", nullable: true })
  activeVersionId!: string | null;

  @CreateDateColumn({ type: "timestamptz", name: "created_at" })
  createdAt!: Date;

  @UpdateDateColumn({ type: "timestamptz", name: "updated_at" })
  updatedAt!: Date;
}
