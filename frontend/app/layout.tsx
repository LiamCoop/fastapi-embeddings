import type { Metadata } from "next";
import { ClerkProvider } from "@clerk/nextjs";
import "./globals.css";

export const metadata: Metadata = {
  title: "Ragtime — AI Reliability Platform",
  description:
    "The platform for teams who need their AI to be right. Ingest, inspect, evaluate, and monitor your AI knowledge with confidence.",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en" className="dark">
      <body className="font-sans antialiased">
        <ClerkProvider>{children}</ClerkProvider>
      </body>
    </html>
  );
}
