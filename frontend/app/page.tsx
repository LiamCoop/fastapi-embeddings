import { auth } from "@clerk/nextjs/server";
import { redirect } from "next/navigation";
import { Navigation } from "@/components/navigation";
import { Hero } from "@/components/hero";
import { PlatformLayers } from "@/components/platform-layers";
import { WorkflowSection } from "@/components/workflow-section";
import { CtaSection } from "@/components/cta-section";
import { Footer } from "@/components/footer";

export default async function Home() {
  const { userId } = await auth();
  if (userId) redirect("/dashboard");

  return (
    <main className="min-h-screen bg-background">
      <Navigation />
      <Hero />
      <PlatformLayers />
      <WorkflowSection />
      <CtaSection />
      <Footer />
    </main>
  );
}
