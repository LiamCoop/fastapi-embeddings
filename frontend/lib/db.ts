import "reflect-metadata";
import { DataSource } from "typeorm";
import { User } from "./entities/user.entity";
import { Organization } from "./entities/organization.entity";
import { OrganizationMembership } from "./entities/organization-membership.entity";
import { KnowledgeBaseSettings } from "./entities/knowledge-base-settings.entity";
import { KnowledgeBase } from "./entities/knowledge-base.entity";
import { Document } from "./entities/document.entity";
import { DocumentVersion } from "./entities/document-version.entity";

const AppDataSource = new DataSource({
  type: "postgres",
  url: process.env.DATABASE_URL,
  entities: [
    User,
    Organization,
    OrganizationMembership,
    KnowledgeBaseSettings,
    KnowledgeBase,
    Document,
    DocumentVersion,
  ],
  synchronize: process.env.NODE_ENV === "development",
  ssl: process.env.NODE_ENV === "production" ? { rejectUnauthorized: false } : false,
});

let initPromise: Promise<DataSource> | null = null;

export function getDb(): Promise<DataSource> {
  if (!initPromise) {
    initPromise = AppDataSource.initialize().catch((err) => {
      initPromise = null;
      throw err;
    });
  }
  return initPromise;
}
