import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  reactStrictMode: true,
  output: "standalone",
  async rewrites() {
    if (process.env.NODE_ENV === "production" && !process.env.API_INTERNAL_URL) {
      return [];
    }
    const backend = (process.env.API_INTERNAL_URL || "http://127.0.0.1:8080").replace(/\/$/, "");
    return [{ source: "/api/:path*", destination: `${backend}/api/:path*` }];
  },
};

export default nextConfig;
