import {
  Entity,
  PrimaryGeneratedColumn,
  Column,
  CreateDateColumn,
  ManyToOne,
  JoinColumn,
  Unique,
} from "typeorm";
import { Organization } from "./organization.entity";

@Entity("organization_memberships")
@Unique(["organizationId", "userClerkId"])
export class OrganizationMembership {
  @PrimaryGeneratedColumn("uuid")
  id!: string;

  @Column({ type: "uuid" })
  organizationId!: string;

  @Column({ type: "varchar" })
  userClerkId!: string;

  @Column({ type: "varchar", default: "member" })
  role!: "owner" | "admin" | "member" | "viewer";

  @Column({ type: "timestamptz", nullable: true })
  acceptedAt!: Date | null;

  @Column({ type: "varchar", nullable: true })
  invitedByClerkId!: string | null;

  @CreateDateColumn({ type: "timestamptz" })
  createdAt!: Date;

  @ManyToOne(() => Organization)
  @JoinColumn({ name: "organizationId" })
  organization!: Organization;
}
