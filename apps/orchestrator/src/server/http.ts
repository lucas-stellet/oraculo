import { Hono } from "hono";
import type { Database } from "bun:sqlite";
import { EpicStore } from "../db/stores/epic-store.ts";
import { StoryStore } from "../db/stores/story-store.ts";
import { TaskStore } from "../db/stores/task-store.ts";
import { ApprovalStore } from "../db/stores/approval-store.ts";
import { VersionStore } from "../db/stores/version-store.ts";
import { ReviewStore } from "../db/stores/review-store.ts";
import { ValidationStore } from "../db/stores/validation-store.ts";
import { AgentStore } from "../db/stores/agent-store.ts";
import { ToolEventStore } from "../db/stores/tool-event-store.ts";
import { SessionEventStore } from "../db/stores/session-event-store.ts";
import { WebSocketHub } from "./ws.ts";
import { Broadcaster } from "./broadcaster.ts";
import { corsMiddleware } from "./middleware.ts";
import { hookRoutes } from "./hooks.ts";
import { apiRoutes } from "./routes.ts";
import { VERSION } from "../version.ts";

export interface CreateAppOptions {
  version?: string;
}

/**
 * Creates a fully wired Hono application with all routes.
 *
 * Accepts a bun:sqlite Database (already migrated) and returns the Hono app
 * along with the WebSocket hub and broadcaster for external use.
 */
export function createApp(db: Database, opts?: CreateAppOptions) {
  const app = new Hono();
  const hub = new WebSocketHub();
  const broadcaster = new Broadcaster();
  const version = opts?.version ?? VERSION;

  // Instantiate stores
  const epics = new EpicStore(db);
  const stories = new StoryStore(db);
  const tasks = new TaskStore(db);
  const approvals = new ApprovalStore(db);
  const versions = new VersionStore(db);
  const reviews = new ReviewStore(db);
  const validations = new ValidationStore(db);
  const agents = new AgentStore(db);
  const toolEvents = new ToolEventStore(db);
  const sessionEvents = new SessionEventStore(db);

  // CORS
  app.use("*", corsMiddleware);

  // Health
  app.get("/health", (c) => c.json({ status: "ok" }));

  // Hook endpoints
  app.route(
    "/hooks",
    hookRoutes({
      epics,
      stories,
      tasks,
      agents,
      toolEvents,
      sessionEvents,
      hub,
      broadcaster,
    }),
  );

  // API endpoints
  app.route(
    "/api",
    apiRoutes({
      epics,
      stories,
      tasks,
      approvals,
      versions,
      reviews,
      validations,
      broadcaster,
      version,
    }),
  );

  return { app, hub, broadcaster };
}
