import { Hono } from "hono";
import type { EpicStore } from "../db/stores/epic-store.ts";
import type { StoryStore } from "../db/stores/story-store.ts";
import type { TaskStore } from "../db/stores/task-store.ts";
import type { AgentStore } from "../db/stores/agent-store.ts";
import type { ToolEventStore } from "../db/stores/tool-event-store.ts";
import type { SessionEventStore } from "../db/stores/session-event-store.ts";
import type { WebSocketHub } from "./ws.ts";
import type { Broadcaster } from "./broadcaster.ts";

export interface HookDeps {
  epics: EpicStore;
  stories: StoryStore;
  tasks: TaskStore;
  agents: AgentStore;
  toolEvents: ToolEventStore;
  sessionEvents: SessionEventStore;
  hub: WebSocketHub;
  broadcaster: Broadcaster;
}

function broadcast(deps: HookDeps, event: string, payload: unknown) {
  deps.hub.broadcast({ event, data: payload });
  deps.broadcaster.push({ event, data: payload });
}

export function hookRoutes(deps: HookDeps): Hono {
  const app = new Hono();

  // POST /hooks/session-start
  app.post("/session-start", async (c) => {
    const body = await c.req.json<{ session_id?: string }>().catch(() => null);
    if (!body) return c.json({ error: "invalid JSON body" }, 400);

    broadcast(deps, "session_started", body);
    return c.json({ status: "ok" });
  });

  // POST /hooks/session-end
  app.post("/session-end", async (c) => {
    const body = await c.req.json<{ session_id?: string }>().catch(() => null);
    if (!body) return c.json({ error: "invalid JSON body" }, 400);

    broadcast(deps, "session_ended", body);
    return c.json({ status: "ok" });
  });

  // POST /hooks/agent-start
  app.post("/agent-start", async (c) => {
    const body = await c.req
      .json<{
        session_id?: string;
        agent_name?: string;
        agent_type?: string;
        task_name?: string;
        story_name?: string;
        epic_name?: string;
      }>()
      .catch(() => null);
    if (!body) return c.json({ error: "invalid JSON body" }, 400);

    if (!body.task_name || !body.story_name || !body.epic_name) {
      return c.json(
        { error: "task_name, story_name, and epic_name are required" },
        400,
      );
    }

    let epic;
    try {
      epic = deps.epics.getByName(body.epic_name);
    } catch {
      return c.json({ error: `epic not found: ${body.epic_name}` }, 404);
    }

    let story;
    try {
      story = deps.stories.getByName(epic.id, body.story_name);
    } catch {
      return c.json({ error: `story not found: ${body.story_name}` }, 404);
    }

    let task;
    try {
      task = deps.tasks.getByName(story.id, body.task_name);
    } catch {
      return c.json({ error: `task not found: ${body.task_name}` }, 404);
    }

    try {
      const agent = deps.agents.Create(
        body.session_id ?? "",
        body.agent_name ?? "",
        body.agent_type ?? "",
        task.id,
      );
      broadcast(deps, "agent_started", agent);
      return c.json(agent);
    } catch (err) {
      return c.json(
        { error: err instanceof Error ? err.message : String(err) },
        500,
      );
    }
  });

  // POST /hooks/agent-stop
  app.post("/agent-stop", async (c) => {
    const body = await c.req
      .json<{ agent_id?: number; status?: string }>()
      .catch(() => null);
    if (!body) return c.json({ error: "invalid JSON body" }, 400);

    try {
      deps.agents.Stop(body.agent_id ?? 0);
      broadcast(deps, "agent_stopped", body);
      return c.json({ status: "ok" });
    } catch (err) {
      return c.json(
        { error: err instanceof Error ? err.message : String(err) },
        500,
      );
    }
  });

  // POST /hooks/tool-used
  app.post("/tool-used", async (c) => {
    const body = await c.req
      .json<{
        session_id?: string;
        tool_name?: string;
        file_path?: string;
      }>()
      .catch(() => null);
    if (!body) return c.json({ error: "invalid JSON body" }, 400);

    try {
      deps.toolEvents.Create(
        body.session_id ?? "",
        body.tool_name ?? "",
        body.file_path ?? "",
      );
      broadcast(deps, "tool_used", body);
      return c.json({ status: "ok" });
    } catch (err) {
      return c.json(
        { error: err instanceof Error ? err.message : String(err) },
        500,
      );
    }
  });

  // POST /hooks/task-completed
  app.post("/task-completed", async (c) => {
    const body = await c.req
      .json<{
        session_id?: string;
        task_name?: string;
        status?: string;
      }>()
      .catch(() => null);
    if (!body) return c.json({ error: "invalid JSON body" }, 400);

    const payload = JSON.stringify({
      task_name: body.task_name ?? "",
      status: body.status ?? "",
    });

    try {
      deps.sessionEvents.Create(
        body.session_id ?? "",
        "task_completed",
        payload,
      );
    } catch {
      // Non-fatal: log recording failure
    }

    broadcast(deps, "task_completed", body);
    return c.json({ status: "ok" });
  });

  // POST /hooks/task-started
  app.post("/task-started", async (c) => {
    const body = await c.req
      .json<{
        session_id?: string;
        task_name?: string;
        story_name?: string;
        epic_name?: string;
      }>()
      .catch(() => null);
    if (!body) return c.json({ error: "invalid JSON body" }, 400);

    if (!body.task_name || !body.story_name || !body.epic_name) {
      return c.json(
        { error: "task_name, story_name, and epic_name are required" },
        400,
      );
    }

    broadcast(deps, "task_started", body);
    return c.json({ status: "ok" });
  });

  // POST /hooks/stop
  app.post("/stop", async (c) => {
    const body = await c.req
      .json<{
        session_id?: string;
        last_assistant_message?: string;
      }>()
      .catch(() => null);
    if (!body) return c.json({ error: "invalid JSON body" }, 400);

    const payload = JSON.stringify({
      last_assistant_message: body.last_assistant_message ?? "",
    });

    try {
      deps.sessionEvents.Create(body.session_id ?? "", "stop", payload);
    } catch {
      // Non-fatal
    }

    broadcast(deps, "stop", body);
    return c.json({ status: "ok" });
  });

  // POST /hooks/teammate-idle
  app.post("/teammate-idle", async (c) => {
    const body = await c.req
      .json<{
        session_id?: string;
        teammate_name?: string;
        team_name?: string;
      }>()
      .catch(() => null);
    if (!body) return c.json({ error: "invalid JSON body" }, 400);

    const payload = JSON.stringify({
      teammate_name: body.teammate_name ?? "",
      team_name: body.team_name ?? "",
    });

    try {
      deps.sessionEvents.Create(
        body.session_id ?? "",
        "teammate_idle",
        payload,
      );
    } catch {
      // Non-fatal
    }

    broadcast(deps, "teammate_idle", body);
    return c.json({ status: "ok" });
  });

  return app;
}
