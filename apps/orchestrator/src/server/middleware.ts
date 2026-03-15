import { cors } from "hono/cors";

/**
 * CORS middleware matching the Go corsMiddleware:
 * - Allow all origins
 * - Allow GET, POST, PUT, DELETE, OPTIONS
 * - Allow Content-Type header
 */
export const corsMiddleware = cors({
  origin: "*",
  allowMethods: ["GET", "POST", "PUT", "DELETE", "OPTIONS"],
  allowHeaders: ["Content-Type"],
});
