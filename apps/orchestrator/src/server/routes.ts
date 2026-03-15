import { Hono } from "hono";
import { streamSSE } from "hono/streaming";
import type { EpicStore } from "../db/stores/epic-store.ts";
import type { StoryStore } from "../db/stores/story-store.ts";
import type { TaskStore } from "../db/stores/task-store.ts";
import type { ApprovalStore } from "../db/stores/approval-store.ts";
import type { VersionStore } from "../db/stores/version-store.ts";
import type { ReviewStore } from "../db/stores/review-store.ts";
import type { ValidationStore } from "../db/stores/validation-store.ts";
import type { Verdict } from "../domain/enums.ts";
import { VERDICTS } from "../domain/enums.ts";
import { DomainError } from "../domain/errors.ts";
import type { Broadcaster } from "./broadcaster.ts";

export interface APIDeps {
  epics: EpicStore;
  stories: StoryStore;
  tasks: TaskStore;
  approvals: ApprovalStore;
  versions: VersionStore;
  reviews: ReviewStore;
  validations: ValidationStore;
  broadcaster: Broadcaster;
  version: string;
}

/** Maps DomainError codes to HTTP status codes. */
function domainStatus(err: DomainError): number {
  switch (err.code) {
    case "not_found":
      return 404;
    case "already_exists":
      return 409;
    case "approval_decided":
      return 409;
    case "invalid_transition":
    case "cyclic_dependency":
    case "missing_required":
    case "invalid_phase":
      return 400;
    default:
      return 500;
  }
}

export function apiRoutes(deps: APIDeps): Hono {
  const app = new Hono();
  const startedAt = new Date().toISOString();

  // --- System ---

  // GET /api/system/status
  app.get("/system/status", (c) => {
    return c.json({
      version: deps.version,
      started_at: startedAt,
    });
  });

  // --- Epics ---

  // GET /api/epics
  app.get("/epics", (c) => {
    try {
      const summaries = deps.epics.listSummaries();
      return c.json(summaries);
    } catch (err) {
      if (err instanceof DomainError)
        return c.json({ error: err.message }, domainStatus(err));
      return c.json(
        { error: err instanceof Error ? err.message : String(err) },
        500,
      );
    }
  });

  // POST /api/epics
  app.post("/epics", async (c) => {
    const body = await c.req
      .json<{ name?: string; description?: string }>()
      .catch(() => null);
    if (!body) return c.json({ error: "invalid JSON body" }, 400);
    if (!body.name) return c.json({ error: "name is required" }, 400);

    try {
      const epic = deps.epics.create(body.name, body.description ?? "");
      return c.json(epic, 201);
    } catch (err) {
      if (err instanceof DomainError)
        return c.json({ error: err.message }, domainStatus(err));
      return c.json(
        { error: err instanceof Error ? err.message : String(err) },
        500,
      );
    }
  });

  // GET /api/epics/:epicName
  app.get("/epics/:epicName", (c) => {
    const name = c.req.param("epicName");
    try {
      const epic = deps.epics.getByName(name);
      return c.json(epic);
    } catch (err) {
      if (err instanceof DomainError)
        return c.json({ error: err.message }, domainStatus(err));
      return c.json(
        { error: err instanceof Error ? err.message : String(err) },
        500,
      );
    }
  });

  // PUT /api/epics/:epicName
  app.put("/epics/:epicName", async (c) => {
    const name = c.req.param("epicName");
    const body = await c.req
      .json<{ description?: string }>()
      .catch(() => null);
    if (!body) return c.json({ error: "invalid JSON body" }, 400);

    try {
      const epic = deps.epics.update(name, body.description ?? "");
      return c.json(epic);
    } catch (err) {
      if (err instanceof DomainError)
        return c.json({ error: err.message }, domainStatus(err));
      return c.json(
        { error: err instanceof Error ? err.message : String(err) },
        500,
      );
    }
  });

  // DELETE /api/epics/:epicName
  app.delete("/epics/:epicName", (c) => {
    const name = c.req.param("epicName");
    try {
      deps.epics.delete(name);
      return c.body(null, 204);
    } catch (err) {
      if (err instanceof DomainError)
        return c.json({ error: err.message }, domainStatus(err));
      return c.json(
        { error: err instanceof Error ? err.message : String(err) },
        500,
      );
    }
  });

  // --- Stories ---

  // GET /api/epics/:epicName/stories
  app.get("/epics/:epicName/stories", (c) => {
    const epicName = c.req.param("epicName");
    try {
      const epic = deps.epics.getByName(epicName);
      const summaries = deps.stories.listSummaries(epic.id);
      return c.json(summaries);
    } catch (err) {
      if (err instanceof DomainError)
        return c.json({ error: err.message }, domainStatus(err));
      return c.json(
        { error: err instanceof Error ? err.message : String(err) },
        500,
      );
    }
  });

  // POST /api/epics/:epicName/stories
  app.post("/epics/:epicName/stories", async (c) => {
    const epicName = c.req.param("epicName");
    const body = await c.req
      .json<{ name?: string; description?: string }>()
      .catch(() => null);
    if (!body) return c.json({ error: "invalid JSON body" }, 400);
    if (!body.name) return c.json({ error: "name is required" }, 400);

    try {
      const epic = deps.epics.getByName(epicName);
      const story = deps.stories.create(
        epic.id,
        body.name,
        body.description ?? "",
      );
      return c.json(story, 201);
    } catch (err) {
      if (err instanceof DomainError)
        return c.json({ error: err.message }, domainStatus(err));
      return c.json(
        { error: err instanceof Error ? err.message : String(err) },
        500,
      );
    }
  });

  // --- Tasks ---

  // GET /api/epics/:epicName/stories/:storyName/tasks
  app.get("/epics/:epicName/stories/:storyName/tasks", (c) => {
    const epicName = c.req.param("epicName");
    const storyName = c.req.param("storyName");
    try {
      const epic = deps.epics.getByName(epicName);
      const story = deps.stories.getByName(epic.id, storyName);
      const enriched = deps.tasks.listEnriched(story.id);
      return c.json(enriched);
    } catch (err) {
      if (err instanceof DomainError)
        return c.json({ error: err.message }, domainStatus(err));
      return c.json(
        { error: err instanceof Error ? err.message : String(err) },
        500,
      );
    }
  });

  // --- Versions ---

  // GET /api/epics/:epicName/stories/:storyName/versions
  app.get("/epics/:epicName/stories/:storyName/versions", (c) => {
    const epicName = c.req.param("epicName");
    const storyName = c.req.param("storyName");
    try {
      const epic = deps.epics.getByName(epicName);
      const story = deps.stories.getByName(epic.id, storyName);
      const versions = deps.versions.ListStoryVersions(story.id);
      return c.json(versions);
    } catch (err) {
      if (err instanceof DomainError)
        return c.json({ error: err.message }, domainStatus(err));
      return c.json(
        { error: err instanceof Error ? err.message : String(err) },
        500,
      );
    }
  });

  // --- Reviews ---

  // GET /api/epics/:epicName/stories/:storyName/reviews
  app.get("/epics/:epicName/stories/:storyName/reviews", (c) => {
    const epicName = c.req.param("epicName");
    const storyName = c.req.param("storyName");
    try {
      const epic = deps.epics.getByName(epicName);
      const story = deps.stories.getByName(epic.id, storyName);
      const reviews = deps.reviews.ListByStory(story.id);
      return c.json(reviews);
    } catch (err) {
      if (err instanceof DomainError)
        return c.json({ error: err.message }, domainStatus(err));
      return c.json(
        { error: err instanceof Error ? err.message : String(err) },
        500,
      );
    }
  });

  // --- Validations ---

  // GET /api/epics/:epicName/stories/:storyName/validations
  app.get("/epics/:epicName/stories/:storyName/validations", (c) => {
    const epicName = c.req.param("epicName");
    const storyName = c.req.param("storyName");
    try {
      const epic = deps.epics.getByName(epicName);
      const story = deps.stories.getByName(epic.id, storyName);
      const validations = deps.validations.ListByStory(story.id);
      return c.json(validations);
    } catch (err) {
      if (err instanceof DomainError)
        return c.json({ error: err.message }, domainStatus(err));
      return c.json(
        { error: err instanceof Error ? err.message : String(err) },
        500,
      );
    }
  });

  // --- Approvals ---

  // GET /api/approvals
  app.get("/approvals", (c) => {
    const status = c.req.query("status");
    const epicIdStr = c.req.query("epic_id");

    try {
      if (epicIdStr) {
        const epicId = parseInt(epicIdStr, 10);
        if (isNaN(epicId)) {
          return c.json({ error: "invalid epic_id" }, 400);
        }
        const approvals = deps.approvals.List(epicId);
        if (status === "pending") {
          return c.json(approvals.filter((a) => a.status === "pending"));
        }
        return c.json(approvals);
      }

      if (status === "pending") {
        return c.json(deps.approvals.ListPending());
      }
      const approvals = deps.approvals.List();
      return c.json(approvals);
    } catch (err) {
      if (err instanceof DomainError)
        return c.json({ error: err.message }, domainStatus(err));
      return c.json(
        { error: err instanceof Error ? err.message : String(err) },
        500,
      );
    }
  });

  // GET /api/approvals/:id
  app.get("/approvals/:id", (c) => {
    const id = c.req.param("id");
    try {
      const approval = deps.approvals.Get(id);
      return c.json(approval);
    } catch (err) {
      if (err instanceof DomainError)
        return c.json({ error: err.message }, domainStatus(err));
      return c.json(
        { error: err instanceof Error ? err.message : String(err) },
        500,
      );
    }
  });

  // POST /api/approvals/:id/verdict
  app.post("/approvals/:id/verdict", async (c) => {
    const id = c.req.param("id");
    const body = await c.req
      .json<{ verdict?: string; comment?: string }>()
      .catch(() => null);
    if (!body) return c.json({ error: "invalid JSON body" }, 400);

    const verdict = body.verdict as Verdict;
    if (!VERDICTS.includes(verdict)) {
      return c.json(
        {
          error:
            "invalid verdict: must be approved, rejected, or needs_revision",
        },
        400,
      );
    }

    try {
      deps.approvals.SetVerdict(id, verdict, body.comment ?? "");
      const updated = deps.approvals.Get(id);
      return c.json(updated);
    } catch (err) {
      if (err instanceof DomainError)
        return c.json({ error: err.message }, domainStatus(err));
      return c.json(
        { error: err instanceof Error ? err.message : String(err) },
        500,
      );
    }
  });

  // POST /api/approvals/:id/comments
  app.post("/approvals/:id/comments", async (c) => {
    const id = c.req.param("id");
    const body = await c.req
      .json<{ selected_text?: string; comment?: string }>()
      .catch(() => null);
    if (!body) return c.json({ error: "invalid JSON body" }, 400);

    if (!body.selected_text || !body.comment) {
      return c.json(
        { error: "selected_text and comment are required" },
        400,
      );
    }

    try {
      const comment = deps.approvals.AddComment(
        id,
        body.selected_text,
        body.comment,
      );
      return c.json(comment, 201);
    } catch (err) {
      if (err instanceof DomainError)
        return c.json({ error: err.message }, domainStatus(err));
      return c.json(
        { error: err instanceof Error ? err.message : String(err) },
        500,
      );
    }
  });

  // GET /api/approvals/:id/comments
  app.get("/approvals/:id/comments", (c) => {
    const id = c.req.param("id");
    try {
      const comments = deps.approvals.ListComments(id);
      return c.json(comments);
    } catch (err) {
      if (err instanceof DomainError)
        return c.json({ error: err.message }, domainStatus(err));
      return c.json(
        { error: err instanceof Error ? err.message : String(err) },
        500,
      );
    }
  });

  // DELETE /api/approvals/:id/comments/:commentId
  // Note: The Go API deletes a single comment by commentId.
  // The TS ApprovalStore only has DeleteCommentsByApproval (batch).
  // This route deletes all comments for the approval as a best-effort match.
  app.delete("/approvals/:id/comments/:commentId", (c) => {
    const commentIdStr = c.req.param("commentId");
    const commentId = parseInt(commentIdStr, 10);
    if (isNaN(commentId)) {
      return c.json({ error: "invalid comment id" }, 400);
    }

    try {
      deps.approvals.DeleteCommentsByApproval(c.req.param("id"));
      return c.body(null, 204);
    } catch (err) {
      if (err instanceof DomainError)
        return c.json({ error: err.message }, domainStatus(err));
      return c.json(
        { error: err instanceof Error ? err.message : String(err) },
        500,
      );
    }
  });

  // --- SSE Log Stream ---

  // GET /api/logs
  app.get("/logs", (c) => {
    return streamSSE(c, async (stream) => {
      // Replay buffered entries
      for (const entry of deps.broadcaster.recent()) {
        await stream.writeSSE({ data: entry });
      }

      // Live stream
      let closed = false;
      const unsub = deps.broadcaster.subscribe((data) => {
        if (!closed) {
          stream.writeSSE({ data }).catch(() => {
            closed = true;
          });
        }
      });

      // Wait until the client disconnects
      stream.onAbort(() => {
        closed = true;
        unsub();
      });

      // Keep the stream open
      await new Promise<void>((resolve) => {
        const interval = setInterval(() => {
          if (closed) {
            clearInterval(interval);
            resolve();
          }
        }, 500);
      });
    });
  });

  return app;
}
