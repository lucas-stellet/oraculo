# Server Lifecycle Guardian Design

**Date:** 2026-03-02
**Status:** Approved

## Context

HTTP hooks (PostToolUse, SubagentStart, SessionEnd, etc.) depend on the Oraculo HTTP server being reachable at `localhost:{port}`. When the server is not running, hooks fail with visible `hook error` messages in the Claude Code UI. There is no mechanism to auto-start or alert the user.

The `oraculo start` command currently bundles both the MCP server (stdio) and the HTTP/WebSocket server in a single process. This makes it impossible for the SessionStart hook to independently manage the HTTP server lifecycle.

## Design

### 1. Split `oraculo start` into subcommands

| Command | Transport | Managed by | Purpose |
|---------|-----------|------------|---------|
| `oraculo start mcp` | stdio | Claude Code (`mcpServers` config) | Approval gates (`request_approval`, `approval_status`) |
| `oraculo start http` | TCP port | SessionStart hook (auto-start) | HTTP hooks + WebSocket + REST API |

Both share the same SQLite database and `.oraculo/config.json` port.

`oraculo start` (without subcommand) continues working as today for backwards compatibility — runs both MCP + HTTP in a single process.

### 2. SessionStart hook as guardian

The `oraculo hook session-start` command becomes the single point of guarantee for the HTTP server:

```
hookSessionStart():
  1. Open database
  2. Read config → port
  3. GET http://localhost:{port}/health (timeout 2s)
  4. If online → skip to step 7
  5. If offline:
     a. Start "oraculo start http" as detached process (setsid, no stdio)
     b. Poll GET /health every 500ms
     c. Max wait: 10s
     d. If started → continue
     e. If timeout → alert and continue
  6. Alert on failure:
     - stderr: "warning: Oraculo server failed to start on port {port}. Telemetry unavailable."
     - stdout: additional context for Claude about degraded state
  7. Register session in SQLite
  8. POST /hooks/session-start (if server is online)
  9. Return exit 0 (never blocks)
```

Claude Code waits for the SessionStart hook to return before proceeding. This effectively delays the session until the HTTP server is ready (up to 10s).

### 3. Idle timeout auto-shutdown

The HTTP server self-terminates after 15 minutes of inactivity:

- Every incoming request (any endpoint) updates a `lastActivity` timestamp
- A watchdog goroutine checks `time.Since(lastActivity)` every 60 seconds
- When idle > 15 minutes → graceful shutdown:
  1. Stop accepting new connections
  2. Wait for in-flight requests to complete (max 5s)
  3. Close database
  4. Exit process

This handles all shutdown scenarios:
- Normal: Claude Code session ends → no more hooks → server idles out
- Crash: Claude Code crashes → SessionEnd never fires → server still idles out
- Multiple sessions: server stays alive as long as any session sends hooks

### 4. Configuration changes

**`install.go`** generates updated `settings.json`:

```json
{
  "hooks": {
    "SessionStart": [{
      "hooks": [{
        "type": "command",
        "command": "oraculo hook session-start"
      }]
    }],
    "PostToolUse": [{
      "matcher": "Bash|Edit|Write|NotebookEdit",
      "hooks": [{
        "type": "http",
        "url": "http://localhost:{port}/hooks/tool-used",
        "timeout": 5
      }]
    }]
  },
  "mcpServers": {
    "oraculo": {
      "command": "oraculo",
      "args": ["start", "mcp"]
    }
  }
}
```

Key change: `mcpServers` now uses `["start", "mcp"]` instead of `["start"]`.

### 5. Process detachment

The SessionStart hook starts the HTTP daemon as a fully detached process:

```go
cmd := exec.Command("oraculo", "start", "http")
cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
cmd.Stdout = nil
cmd.Stderr = nil
cmd.Start()
cmd.Process.Release()
```

The daemon writes its PID to `.oraculo/server.pid` for diagnostics (not used for lifecycle management — idle timeout handles shutdown).

### 6. Session lifecycle diagram

```
┌─ Claude Code session ────────────────────────────────┐
│                                                       │
│  1. SessionStart hook fires                           │
│     └─ oraculo hook session-start                     │
│        ├─ Health check localhost:{port}               │
│        ├─ If offline: oraculo start http (detached)   │
│        ├─ Poll /health up to 10s                      │
│        ├─ If failed: stderr + stdout alerts           │
│        ├─ Register session in SQLite                  │
│        └─ POST /hooks/session-start                   │
│                                                       │
│  2. MCP server starts (Claude Code manages)           │
│     └─ oraculo start mcp (stdio)                      │
│        └─ Exposes request_approval + approval_status  │
│                                                       │
│  3. Agent works...                                    │
│     └─ PostToolUse hooks fire → HTTP POST             │
│        └─ Server is already running ✓                 │
│                                                       │
│  4. SessionEnd hook fires → HTTP POST                 │
│                                                       │
│  5. HTTP daemon persists (reused by next session)     │
│     └─ Idle timeout: 15min no requests → shutdown     │
└───────────────────────────────────────────────────────┘
```

## Files to modify

| File | Change |
|------|--------|
| `src/cli/start.go` | Split into `start mcp` and `start http` subcommands |
| `src/cli/hook_session.go` | Add auto-start logic with health poll |
| `src/server/server.go` | Add idle timeout watchdog |
| `src/cli/install.go` | Update `mcpServers` args to `["start", "mcp"]` |
| `src/cli/root.go` | Register new subcommands |

## Verification

1. `oraculo start http` — starts HTTP server, responds to `/health`
2. `oraculo start mcp` — starts MCP server on stdio
3. `oraculo start` — backwards compatible, runs both
4. Kill HTTP server → run `oraculo hook session-start` → server auto-starts
5. Stop sending requests → server shuts down after 15 minutes
6. `oraculo install` → verify `settings.json` has `["start", "mcp"]`
