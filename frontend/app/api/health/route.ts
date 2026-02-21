import { getKnowledgeDal, getKnowledgeDalMode } from "@/lib/dal";

export async function GET() {
  try {
    const dal = getKnowledgeDal();
    await dal.ping();
    return Response.json({ status: "ok", db: "connected", dal: getKnowledgeDalMode() });
  } catch (error) {
    console.error("Health check failed", error);
    return Response.json({ status: "error", db: "disconnected", dal: getKnowledgeDalMode() }, { status: 503 });
  }
}
