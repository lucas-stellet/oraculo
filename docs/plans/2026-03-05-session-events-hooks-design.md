# Session Events & Hook Purpose — Design

## Problem

Four hooks (Stop, SessionEnd, TaskCompleted, TeammateIdle) only broadcast empty WebSocket events. They serve no real purpose. Additionally, HTTP hooks fail with `ECONNREFUSED` when the server is offline, which surfaces as user-visible errors — particularly `SessionEnd`, which fires after the server may have already shut down via idle timeout.

## Decision

Give each hook a concrete purpose: record structured events in a new `session_events` table for implementation logging and future analytics. Apply the hybrid approach — critical lifecycle hooks (SessionStart, SessionEnd) use resilient command hooks that write directly to SQLite; in-session hooks (Stop, TaskCompleted, TeammateIdle) remain HTTP since the server is reliably online during active sessions.

## Schema Changes

### New table: `session_events`

```sql
CREATE TABLE session_events (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    session_id TEXT NOT NULL REFERENCES claude_sessions(id),
    event_type TEXT NOT NULL
               CHECK (event_type IN ('task_completed','stop','teammate_idle','session_end')),
    payload    TEXT DEFAULT '{}',
    created_at TEXT DEFAULT (datetime('now'))
);
CREATE INDEX idx_session_events_session ON session_events(session_id);
```

Append-only. One row per event. Payload is freeform JSON specific to each event type.

### Altered table: `claude_sessions`

- PK becomes the `session_id` from Claude Code (read from stdin), replacing the internally generated UUID.
- Add column `ended_at TEXT`.
- Use `INSERT OR IGNORE` on first start; `UPDATE metadata` on resume/clear/compact.

## Hook Behavior

| Hook            | Type    | Writes to                                              | Decision control |
|:----------------|:--------|:-------------------------------------------------------|:-----------------|
| `SessionStart`  | command | `claude_sessions` (upsert)                             | None             |
| `SessionEnd`    | command | `claude_sessions.ended_at` + `session_events`          | None             |
| `Stop`          | HTTP    | `session_events` (type=stop)                           | None             |
| `TaskCompleted` | HTTP    | `session_events` (type=task_completed)                 | None             |
| `TeammateIdle`  | HTTP    | `session_events` (type=teammate_idle)                  | None             |

All hooks also broadcast via WebSocket for the UI when the server is online.

### SessionStart (command hook — changed)

Reads `session_id`, `source`, and `model` from stdin JSON (Claude Code input). Replaces the internally generated UUID with Claude Code's `session_id`.

- First invocation (`source=startup`): `INSERT OR IGNORE` into `claude_sessions`.
- Subsequent invocations (`resume`, `clear`, `compact`): `UPDATE metadata` with latest timestamp and source.
- Auto-starts HTTP server if offline (existing behavior, unchanged).
- Notifies server via POST if online (existing behavior, unchanged).

### SessionEnd (command hook — changed from HTTP)

Reads `session_id` and `reason` from stdin JSON.

- Updates `ended_at` in `claude_sessions`.
- Inserts into `session_events` with `event_type=session_end` and `payload={"reason":"..."}`.
- Notifies server via POST if online (for WebSocket broadcast to UI).
- Fails silently if server is offline — the SQLite write already captured the data.

### Stop (HTTP hook — changed handler)

Server handler now writes to `session_events` with `event_type=stop`. Payload includes `last_assistant_message` (truncated if large). Maintains existing WebSocket broadcast.

If server is offline, event is lost. Acceptable since Stop fires during active sessions when the server is almost always running.

### TaskCompleted (HTTP hook — changed handler)

Server handler now writes to `session_events` with `event_type=task_completed`. Payload includes `task_name` and `status`. Maintains existing WebSocket broadcast.

Same offline trade-off as Stop.

### TeammateIdle (HTTP hook — changed handler)

Server handler now writes to `session_events` with `event_type=teammate_idle`. Payload includes `teammate_name` and `team_name`. Maintains existing WebSocket broadcast.

Same offline trade-off as Stop.

## Unchanged Hooks

- `SubagentStart` — already writes to `agents` table via HTTP. No change.
- `SubagentStop` — already writes to `agents` table via HTTP. No change.
- `PostToolUse` — already writes to `tool_events` table via HTTP. No change.

## Event Payloads

```jsonc
// task_completed
{"task_name": "implement-auth", "status": "completed"}

// stop
{"last_assistant_message": "I've completed..."}

// teammate_idle
{"teammate_name": "researcher", "team_name": "my-project"}

// session_end
{"reason": "prompt_input_exit"}
```

## Migration

New migration adds:
1. `ended_at` column to `claude_sessions`.
2. `session_events` table with index.

Existing `claude_sessions` rows keep their old UUIDs. New sessions will use Claude Code's `session_id`.
