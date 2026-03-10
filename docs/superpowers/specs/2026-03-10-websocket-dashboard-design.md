# WebSocket Dashboard + Agent↔Task Association

**Date:** 2026-03-10
**Status:** Approved

## Context

The dashboard already has a fully functional WebSocket hub in the Go backend (`apps/backend/src/ws/hub.go`, endpoint `GET /ws`). It broadcasts 10 event types including `approval_requested`, `approval_decided`, `agent_started`, `agent_stopped`, `task_completed`, and others. However, no page in the Next.js dashboard consumes the WebSocket — all data is fetched once on mount with no real-time updates.

Additionally, `Agent` and `Task` are independent DB tables with no association. The execute skill knows which agent handles which task at dispatch time, but this link is never persisted.

## Goals

1. Connect the dashboard to the WebSocket for real-time updates across all relevant pages.
2. Persist the agent↔task association so the Tasks tab can show which agent is executing each task.

---

## Backend Changes

### 1. DB Migration

```sql
ALTER TABLE agents ADD COLUMN task_id INTEGER REFERENCES tasks(id);
```

The column is nullable at the schema level (SQLite constraint for backward compatibility with existing rows). Enforcement is at the application layer: the `POST /hooks/agent-start` handler rejects requests without `task_name`.

### 2. `POST /hooks/agent-start` — Required Fields

New required fields added to the payload:

```json
{
  "session_id": "...",
  "agent_name": "...",
  "agent_type": "...",
  "task_name": "...",
  "story_name": "...",
  "epic_name": "..."
}
```

The handler resolves `epic_name → story_name → task_name` to a `task_id` and persists it on the agent row. Returns `400` if any of the three new fields is absent.

The `agent_started` WS event payload now includes `task_id`.

### 3. New Endpoint: `POST /hooks/task-started`

Mirrors the existing `POST /hooks/task-completed`. Emits a `task_started` WS event:

```json
{ "event": "task_started", "data": { "task_name": "...", "story_name": "...", "epic_name": "..." } }
```

---

## Execute Skill Changes

In `claude-kit/skills/oraculo/execute/phases/01-team-assembly.md`:

- Add `task_name`, `story_name`, `epic_name` to the agent-start hook payload.
- Call `POST /hooks/task-started` before dispatching each agent.

---

## Frontend Architecture

### WebSocket Provider

A single `WebSocketProvider` in `apps/dashboard/src/app/epics/[id]/layout.tsx` opens one connection per epic context. It reconnects automatically on disconnect.

```
layout.tsx
  └── WebSocketProvider (1 connection)
       └── _client.tsx → useWebSocket(handler)
```

A `useWebSocket(handler)` hook lets any `_client.tsx` register a callback for incoming events. The handler receives the parsed event object and decides what to do.

### Per-Page Event Handling

| Page | Event | Action |
|---|---|---|
| `approvals/_client.tsx` | `approval_requested` | Re-fetch `api.getApproval(id)` → prepend to pending list |
| `approvals/_client.tsx` | `approval_decided` | Re-fetch `api.getApproval(id)` → move to resolved list |
| `approvals/[id]/review/_client.tsx` | `approval_decided` | Show "already decided" banner if event ID matches |
| `epics/[id]/_client.tsx` | `task_started`, `task_completed` | Re-fetch `api.listStories(epicName)` → update progress bars |
| `stories/[id]/_client.tsx` | `task_started`, `task_completed` | Re-fetch `api.listTasks(epicName, storyName)` → update task rows |
| `stories/[id]/_client.tsx` | `agent_started` | Local state update: mark task with badge "executing · {agent_name}" |
| `stories/[id]/_client.tsx` | `agent_stopped` | Local state update: remove executing badge from task |

### Re-fetch vs. Local Update Strategy

- **Re-fetch**: used when the event implies DB state has changed (task status, approval status). Simple and consistent with existing patterns.
- **Local update**: used for transient agent activity badges that are not persisted. Avoids a round-trip for ephemeral UI state.

---

## WebSocket URL

- **Dev**: proxied via Next.js rewrite `{ source: "/ws", destination: "http://localhost:6077/ws" }` (already configured in `next.config.ts`)
- **Prod**: browser connects directly to `ws://same-origin/ws` — served by the embedded Go server. No Next.js involvement.

The static export (`output: "export"`) has no effect on WebSocket behavior. Client components run fully in the browser after hydration.

---

## Out of Scope

- Notification UI / toast system for incoming approvals
- Agent list page (no page exists today; can be added later)
- Server-sent events or polling fallback (WS reconnect is sufficient)
