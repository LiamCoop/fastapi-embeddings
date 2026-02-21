import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  output: "standalone",
  serverExternalPackages: ["typeorm", "pg", "reflect-metadata", "prisma", "@prisma/client"],
  async redirects() {
    return [
      {
        source: "/workspace",
        destination: "/",
        permanent: false,
      },
      {
        source: "/knowledge",
        destination: "/",
        permanent: false,
      },
      {
        source: "/knowledge/:id",
        destination: "/",
        permanent: false,
      },
      {
        source: "/playground",
        destination: "/",
        permanent: false,
      },
      {
        source: "/evaluation",
        destination: "/",
        permanent: false,
      },
      {
        source: "/health",
        destination: "/",
        permanent: false,
      },
      {
        source: "/projects",
        destination: "/",
        permanent: false,
      },
      {
        source: "/projects/:id",
        destination: "/",
        permanent: false,
      },
      {
        source: "/integrations",
        destination: "/",
        permanent: false,
      },
      {
        source: "/settings",
        destination: "/",
        permanent: false,
      },
    ];
  },
};

export default nextConfig;
