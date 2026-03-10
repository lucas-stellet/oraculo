import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  async rewrites() {
    return [
      { source: "/api/:path*", destination: "http://localhost:6077/api/:path*" },
      { source: "/ws", destination: "http://localhost:6077/ws" },
    ];
  },
};

export default nextConfig;
