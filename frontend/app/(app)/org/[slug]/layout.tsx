import { OrgSidebar } from "./_components/OrgSidebar";

export default async function OrgLayout({
  children,
  params,
}: {
  children: React.ReactNode;
  params: Promise<{ slug: string }>;
}) {
  const { slug } = await params;

  return (
    <div className="flex h-full">
      <OrgSidebar slug={slug} />
      <main className="flex-1 overflow-y-auto">{children}</main>
    </div>
  );
}
