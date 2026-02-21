import {
  Entity,
  PrimaryColumn,
  Column,
  CreateDateColumn,
  UpdateDateColumn,
} from "typeorm";

@Entity("knowledge_base_settings")
export class KnowledgeBaseSettings {
  @PrimaryColumn({ type: "uuid" })
  kbId!: string;

  @Column({ type: "uuid" })
  organizationId!: string;

  @Column({ type: "varchar", nullable: true })
  displayName!: string | null;

  @Column({ type: "jsonb", default: {} })
  retrievalSettings!: Record<string, unknown>;

  @Column({ type: "jsonb", default: {} })
  evaluationSettings!: Record<string, unknown>;

  @Column({ type: "boolean", default: true })
  isVisible!: boolean;

  @CreateDateColumn({ type: "timestamptz" })
  createdAt!: Date;

  @UpdateDateColumn({ type: "timestamptz" })
  updatedAt!: Date;
}
