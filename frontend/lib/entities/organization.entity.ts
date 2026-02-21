import {
  Entity,
  PrimaryGeneratedColumn,
  Column,
  CreateDateColumn,
  UpdateDateColumn,
} from "typeorm";

@Entity("organizations")
export class Organization {
  @PrimaryGeneratedColumn("uuid")
  id!: string;

  @Column({ type: "varchar" })
  name!: string;

  @Column({ type: "varchar", unique: true })
  slug!: string;

  @Column({ type: "varchar", nullable: true })
  stripeCustomerId!: string | null;

  @Column({ type: "varchar", default: "free" })
  plan!: "free" | "pro" | "enterprise";

  @Column({ type: "timestamptz", nullable: true })
  planExpiresAt!: Date | null;

  @Column({ type: "boolean", default: true })
  isActive!: boolean;

  @Column({ type: "jsonb", default: {} })
  settings!: Record<string, unknown>;

  @CreateDateColumn({ type: "timestamptz" })
  createdAt!: Date;

  @UpdateDateColumn({ type: "timestamptz" })
  updatedAt!: Date;
}
