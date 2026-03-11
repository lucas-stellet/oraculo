# Approval Comments & MCP Migration

## Problem

When a user rejects a document in the dashboard, the agent receives only a flat text string as feedback. The user has no structured way to point at specific parts of the document and explain what's wrong. Additionally, the approval flow uses CLI polling instead of MCP's blocking mechanism, wasting tokens and cycles.

## Solution

Persist inline comments on approvals, enrich the MCP `request_approval` return with structured feedback, and migrate the approval flow from CLI polling to MCP blocking.

## Data Model

New migration **V9** — table `approval_comments`:

```sql
CREATE TABLE approval_comments (
  id          INTEGER PRIMARY KEY AUTOINCREMENT,
  approval_id TEXT NOT NULL REFERENCES approvals(id),
  selected_text TEXT NOT NULL,
  comment     TEXT NOT NULL,
  created_at  TEXT DEFAULT (datetime('now'))
);
CREATE INDEX idx_approval_comments_approval ON approval_comments(approval_id);
```

Note: `approval_id` is TEXT (UUID) to match `approvals.id`.

Rules:
- Approval with verdict `approved` → all comments for that approval are deleted.
- Approval with verdict `rejected` → comments remain in the database.
- Comments can be deleted individually before a verdict is submitted.

## Backend API

### New Endpoints

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/api/approvals/{id}/comments` | Create comment (`{ selected_text, comment }`) |
| `GET` | `/api/approvals/{id}/comments` | List comments for an approval |
| `DELETE` | `/api/approvals/{id}/comments/{commentId}` | Delete a single comment |

### Verdict Handler Update

`POST /api/approvals/{id}/verdict` — existing endpoint, updated behavior:
- If `verdict = approved`: delete all comments for that approval after recording the verdict.
- If `verdict = rejected`: comments remain. Bridge fetches them when unblocking the agent.

## MCP

### `request_approval` — Updated Return

When the agent unblocks, `VerdictResult` includes full context and structured comments:

```json
{
  "id": "uuid",
  "type": "design",
  "epic_id": 1,
  "story_id": 3,
  "status": "rejected",
  "content": "## Original document...",
  "comment": "General rejection reason",
  "comments": [
    { "selected_text": "The system should...", "comment": "Contradicts requirement X" },
    { "selected_text": "Use REST", "comment": "We prefer gRPC here" }
  ]
}
```

- `comment`: general comment from the rejection modal (may be empty if inline comments exist).
- `comments[]`: inline comments tied to specific text selections.
- If `approved`, `comments` is an empty array (already deleted).

Note: `approval_id` is implicit from the request context — the frontend type `InlineComment` does not need an `approval_id` field.

### `approval_status` — Updated Return

Same enriched format, includes current comments (useful for pre-verdict queries).

### Struct Changes

**`VerdictResult`** (bridge.go) — current fields: `ID`, `Verdict`, `Comment`. Updated:

```go
type VerdictResult struct {
	ID       string
	Type     ApprovalType
	EpicID   *int
	StoryID  *int
	Content  string
	Verdict  Verdict
	Comment  string
	Comments []ApprovalComment  // new
}

type ApprovalComment struct {
	ID           int
	SelectedText string
	Comment      string
	CreatedAt    time.Time
}
```

**`requestApprovalOutput`** (mcp/server.go) — current fields: `ID`, `Status`, `Verdict`, `Comment`. Updated to include `Type`, `EpicID`, `StoryID`, `Content`, `Comments`.

**`ApprovalRequest`** (bridge.go) — must wire `EpicID` and `StoryID` (currently passed as `nil` from the MCP handler). The MCP handler already accepts these fields in `requestApprovalInput` but does not forward them.

### Bridge Update

`bridge.Decide()` updated flow, explicit for each path:

**Approved:**
1. Record verdict in DB.
2. Delete all comments for the approval.
3. Send `VerdictResult` (empty comments array) through the channel.

**Rejected:**
1. Record verdict in DB.
2. Fetch comments from DB.
3. Send `VerdictResult` (with comments array) through the channel.

## Frontend

### Comment Persistence

- Comments are saved to the backend immediately when the user submits via the popover (`POST /api/approvals/{id}/comments`).
- On page load, existing comments are restored from the backend (`GET /api/approvals/{id}/comments`).
- No more in-memory-only state.

### Document Highlighting

- Commented text selections are visually highlighted (amber/yellow background) in preview mode.
- Hover on a highlight shows a tooltip with the comment preview.
- Click on a highlight opens the full comment with a delete option.
- If the user selects text that already has a comment, open the existing comment instead of creating a new one.

**Implementation strategy:** The current codebase has a `CommentHighlight` component and a deferred note about inline highlighting ("true inline highlight requires DOM manipulation after render"). The approach is to post-process the rendered markdown DOM: after `react-markdown` renders, walk text nodes and wrap matching `selected_text` spans with the highlight component. This is the most complex frontend task and should be implemented carefully to handle edge cases (text spanning multiple DOM nodes, partial matches, overlapping selections).

### Rejection Flow

- User clicks Reject.
- Frontend checks if inline comments exist for this approval.
  - **Zero inline comments** → general comment field is **required** before submitting.
  - **Has inline comments** → general comment field is **optional**.
- Calls `POST /api/approvals/{id}/verdict` with verdict and optional general comment.

### Approval Flow

- User clicks Approve → calls verdict endpoint → UI clears all highlights (backend deletes comments).

## Skills

### Rename CLI → Tools

- `cli/SKILL.md` renamed to `tools/SKILL.md`.
- Tools categorized into:
  - **CLI**: data operations (CRUD for epics, stories, tasks, versions, phases, memory, etc.)
  - **MCP**: blocking/interactive operations (`request_approval`, `approval_status`)

### Approval Migration: CLI Polling → MCP Blocking

All skills that currently poll `oraculo tools approval status` migrate to using the MCP `request_approval` tool, which blocks until the verdict is received.

Affected skills and phases:
- **Epic** — `phases/08-approval.md`
- **Story** — `phases/05-approval.md`
- **Plan** — `phases/05-artifact.md` (execution-plan and design approvals)
- **Validate** — `phases/01-qa-dispatch.md` (qa-escalation approval)

### Rejection Handling with Comments

When `request_approval` returns with `status: "rejected"`:

1. If `comments[]` is not empty → skill analyzes inline comments, identifies what needs to change, and routes to the appropriate phase.
2. If `comments[]` is empty but `comment` (general) is not empty → skill uses the general comment as the rejection reason.
3. If both are empty → skill asks the user for the rejection reason via `AskUserQuestion`.

### Install Update

`oraculo install` updates `.claude/settings.json` to register:

```json
{
  "mcpServers": {
    "oraculo": {
      "command": "oraculo",
      "args": ["start"]
    }
  }
}
```

Changed from `["start", "mcp"]` to `["start"]` so that HTTP and MCP run in the same process, enabling the bridge's in-memory channel mechanism.

Note: `oraculo start` already supports combined HTTP+MCP mode — see `cli/start.go` `runStartAll()`. No changes needed to the start command itself, only to the install command's settings output.

## Flow Diagram

```
Agent (Skill)                    Backend                         Dashboard
     |                              |                               |
     |-- request_approval (MCP) --> |                               |
     |   (blocks on bridge)         |-- WS: approval_requested ---> |
     |                              |                               |
     |                              |    User reviews document      |
     |                              | <-- GET /comments ----------- |
     |                              |                               |
     |                              |    User adds inline comment   |
     |                              | <-- POST /comments ---------- |
     |                              |                               |
     |                              |    User clicks Reject         |
     |                              | <-- POST /verdict ----------- |
     |                              |                               |
     |   bridge.Decide():           |                               |
     |   1. Record verdict          |                               |
     |   2. Fetch comments          |                               |
     |   3. Send via channel        |-- WS: approval_decided -----> |
     |                              |                               |
     | <-- VerdictResult ---------- |                               |
     |   { status, comment,         |                               |
     |     comments[], content,     |                               |
     |     type, epic_id, ... }     |                               |
     |                              |                               |
     |  Skill analyzes comments     |                               |
     |  and routes accordingly      |                               |
```
