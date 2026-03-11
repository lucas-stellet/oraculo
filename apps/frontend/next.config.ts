import type { NextConfig } from "next";

// In production builds, export static files for embedding in the Go binary.
// In dev, use rewrites to proxy API/WS calls to the Go backend at :6077.
const isDev = process.env.NODE_ENV === "development";

const nextConfig: NextConfig = {
  ...(isDev
    ? {
        async rewrites() {
          return [
            { source: "/api/:path*", destination: "http://localhost:6077/api/:path*" },
            { source: "/ws", destination: "http://localhost:6077/ws" },
          ];
        },
      }
    : { output: "export" }),
};

export default nextConfig;
