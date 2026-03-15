import { describe, test, expect, beforeEach } from "bun:test";
import { openMemory } from "../../db/database.ts";
import { createApp } from "../http.ts";
import type { Database } from "bun:sqlite";

let db: Database;
let app: ReturnType<typeof createApp>["app"];

beforeEach(() => {
  db = openMemory();
  const result = createApp(db, { version: "test-0.0.1" });
  app = result.app;
});

async function req(
  method: string,
  path: string,
  body?: unknown,
): Promise<Response> {
  const init: RequestInit = { method };
  if (body !== undefined) {
    init.headers = { "Content-Type": "application/json" };
    init.body = JSON.stringify(body);
  }
  return app.request(path, init);
}

// --- Health ---

describe("GET /health", () => {
  test("returns ok", async () => {
    const res = await req("GET", "/health");
    expect(res.status).toBe(200);
    const data = await res.json();
    expect(data).toEqual({ status: "ok" });
  });
});

// --- System ---

describe("GET /api/system/status", () => {
  test("returns version and started_at", async () => {
    const res = await req("GET", "/api/system/status");
    expect(res.status).toBe(200);
    const data = (await res.json()) as { version: string; started_at: string };
    expect(data.version).toBe("test-0.0.1");
    expect(data.started_at).toBeTruthy();
  });
});

// --- Epics CRUD ---

describe("Epics API", () => {
  test("list epics returns empty array initially", async () => {
    const res = await req("GET", "/api/epics");
    expect(res.status).toBe(200);
    const data = await res.json();
    expect(data).toEqual([]);
  });

  test("create epic", async () => {
    const res = await req("POST", "/api/epics", {
      name: "my-epic",
      description: "A test epic",
    });
    expect(res.status).toBe(201);
    const data = (await res.json()) as { name: string; description: string };
    expect(data.name).toBe("my-epic");
    expect(data.description).toBe("A test epic");
  });

  test("create epic without name returns 400", async () => {
    const res = await req("POST", "/api/epics", { description: "no name" });
    expect(res.status).toBe(400);
  });

  test("create duplicate epic returns 409", async () => {
    await req("POST", "/api/epics", { name: "dup", description: "first" });
    const res = await req("POST", "/api/epics", {
      name: "dup",
      description: "second",
    });
    expect(res.status).toBe(409);
  });

  test("get epic by name", async () => {
    await req("POST", "/api/epics", {
      name: "fetch-me",
      description: "desc",
    });
    const res = await req("GET", "/api/epics/fetch-me");
    expect(res.status).toBe(200);
    const data = (await res.json()) as { name: string };
    expect(data.name).toBe("fetch-me");
  });

  test("get non-existent epic returns 404", async () => {
    const res = await req("GET", "/api/epics/nope");
    expect(res.status).toBe(404);
  });

  test("update epic", async () => {
    await req("POST", "/api/epics", { name: "upd", description: "old" });
    const res = await req("PUT", "/api/epics/upd", {
      description: "new desc",
    });
    expect(res.status).toBe(200);
    const data = (await res.json()) as { description: string };
    expect(data.description).toBe("new desc");
  });

  test("delete epic", async () => {
    await req("POST", "/api/epics", { name: "del-me", description: "bye" });
    const res = await req("DELETE", "/api/epics/del-me");
    expect(res.status).toBe(204);

    const check = await req("GET", "/api/epics/del-me");
    expect(check.status).toBe(404);
  });

  test("delete non-existent epic returns 404", async () => {
    const res = await req("DELETE", "/api/epics/ghost");
    expect(res.status).toBe(404);
  });
});

// --- Stories ---

describe("Stories API", () => {
  test("list stories for epic", async () => {
    await req("POST", "/api/epics", { name: "e1", description: "" });
    const res = await req("GET", "/api/epics/e1/stories");
    expect(res.status).toBe(200);
    expect(await res.json()).toEqual([]);
  });

  test("create story under epic", async () => {
    await req("POST", "/api/epics", { name: "e2", description: "" });
    const res = await req("POST", "/api/epics/e2/stories", {
      name: "s1",
      description: "story one",
    });
    expect(res.status).toBe(201);
    const data = (await res.json()) as { name: string };
    expect(data.name).toBe("s1");
  });

  test("list stories for non-existent epic returns 404", async () => {
    const res = await req("GET", "/api/epics/nope/stories");
    expect(res.status).toBe(404);
  });
});

// --- Approvals ---

describe("Approvals API", () => {
  test("list approvals returns empty initially", async () => {
    const res = await req("GET", "/api/approvals");
    expect(res.status).toBe(200);
    expect(await res.json()).toEqual([]);
  });
});

// --- Hooks ---

describe("Hook endpoints", () => {
  test("POST /hooks/session-start returns ok", async () => {
    const res = await req("POST", "/hooks/session-start", {
      session_id: "sess-1",
    });
    expect(res.status).toBe(200);
    const data = (await res.json()) as { status: string };
    expect(data.status).toBe("ok");
  });

  test("POST /hooks/session-end returns ok", async () => {
    const res = await req("POST", "/hooks/session-end", {
      session_id: "sess-1",
    });
    expect(res.status).toBe(200);
    const data = (await res.json()) as { status: string };
    expect(data.status).toBe("ok");
  });

  test("POST /hooks/task-started with missing fields returns 400", async () => {
    const res = await req("POST", "/hooks/task-started", {
      session_id: "s1",
    });
    expect(res.status).toBe(400);
  });

  test("POST /hooks/task-started with all fields returns ok", async () => {
    const res = await req("POST", "/hooks/task-started", {
      session_id: "s1",
      task_name: "t1",
      story_name: "st1",
      epic_name: "ep1",
    });
    expect(res.status).toBe(200);
  });

  test("POST /hooks/agent-start with missing epic returns 404", async () => {
    const res = await req("POST", "/hooks/agent-start", {
      session_id: "s1",
      agent_name: "coder",
      agent_type: "code",
      task_name: "t1",
      story_name: "st1",
      epic_name: "no-such-epic",
    });
    expect(res.status).toBe(404);
  });

  test("POST /hooks/tool-used returns ok", async () => {
    const res = await req("POST", "/hooks/tool-used", {
      session_id: "s1",
      tool_name: "Read",
      file_path: "/tmp/test.ts",
    });
    expect(res.status).toBe(200);
  });

  test("POST /hooks/stop returns ok", async () => {
    const res = await req("POST", "/hooks/stop", {
      session_id: "s1",
      last_assistant_message: "done",
    });
    expect(res.status).toBe(200);
  });

  test("POST /hooks/teammate-idle returns ok", async () => {
    const res = await req("POST", "/hooks/teammate-idle", {
      session_id: "s1",
      teammate_name: "coder-1",
      team_name: "squad",
    });
    expect(res.status).toBe(200);
  });

  test("POST /hooks/task-completed returns ok", async () => {
    const res = await req("POST", "/hooks/task-completed", {
      session_id: "s1",
      task_name: "t1",
      status: "completed",
    });
    expect(res.status).toBe(200);
  });

  test("POST /hooks/agent-stop returns ok", async () => {
    const res = await req("POST", "/hooks/agent-stop", {
      agent_id: 999,
      status: "completed",
    });
    // agent_id=999 doesn't exist, but Stop just runs UPDATE (no error for missing rows)
    expect(res.status).toBe(200);
  });
});

// --- CORS ---

describe("CORS", () => {
  test("OPTIONS request returns CORS headers", async () => {
    const res = await app.request("/health", {
      method: "OPTIONS",
      headers: { Origin: "http://localhost:3000" },
    });
    // Should include CORS headers
    expect(
      res.headers.get("Access-Control-Allow-Origin"),
    ).toBe("*");
  });

  test("GET response includes CORS headers", async () => {
    const res = await req("GET", "/health");
    expect(
      res.headers.get("Access-Control-Allow-Origin"),
    ).toBe("*");
  });
});
