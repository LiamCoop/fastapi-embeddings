import type { KnowledgeDal } from "@/lib/dal/contracts";
import { prismaKnowledgeDal } from "@/lib/dal/prisma-knowledge-dal";
import { typeormKnowledgeDal } from "@/lib/dal/typeorm-knowledge-dal";

const DAL_TYPEORM = "typeorm";
const DAL_PRISMA = "prisma";

function resolveDalMode(): string {
  const raw = process.env.FRONTEND_DAL?.trim().toLowerCase();
  if (raw === DAL_PRISMA) {
    return DAL_PRISMA;
  }

  return DAL_TYPEORM;
}

export function getKnowledgeDal(): KnowledgeDal {
  return resolveDalMode() === DAL_PRISMA ? prismaKnowledgeDal : typeormKnowledgeDal;
}

export function getKnowledgeDalMode(): "typeorm" | "prisma" {
  return resolveDalMode() === DAL_PRISMA ? DAL_PRISMA : DAL_TYPEORM;
}
