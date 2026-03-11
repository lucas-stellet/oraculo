# Wails Desktop App Design

## Summary

Split Oraculo into two binaries: `oraculo` (CLI/server) and `oraculo-desktop` (Wails v2 native app). The desktop app is a launcher and viewer that connects to multiple project servers via HTTP/WS. The server drops embedded frontend assets; the desktop embeds the Next.js build and bundles the `oraculo` binary.

## Motivation

- **Native window** -- desktop app in the dock/taskbar, not a browser tab
- **Distribution** -- installable app (`.dmg`, `.exe`) with bundled server binary
- **Native features** -- system tray, native notifications, global shortcuts

## Architecture

Two binaries, three apps in the monorepo:

```
apps/
├── backend/     -- Go: CLI + HTTP + WS + MCP (oraculo binary, no frontend embed)
├── frontend/    -- Next.js (rename from dashboard, static export)
└── desktop/     -- Wails v2: embeds frontend/out/, bundles oraculo binary
```

Communication:

```
Claude Code <-- stdio (MCP) --> oraculo server (per project)
                                      ^
                                      | HTTP + WS
                                      v
                              oraculo-desktop (single instance, multi-project)
```

The MCP server stays in the `oraculo` binary, running via stdio with Claude Code. The desktop never touches MCP.

## Registry and Discovery

### Server Registry: `~/.oraculo/servers.json`

Written exclusively by the `oraculo` binary. Each server registers on startup and unregisters on graceful shutdown via `defer`.

```json
[
  {
    "project": "gastos-pessoais",
    "path": "/Users/lucas/dev/gastos",
    "port": 3100,
    "pid": 12345,
    "started_at": "2026-03-11T10:30:00Z"
  }
]
```

- File created on first `oraculo start` if it doesn't exist
- Cross-platform file locking via `github.com/gofrs/flock` (`flock` is Unix-only; this library handles Windows `LockFileEx` transparently)
- Desktop reads and validates entries (checks PID alive via `kill(pid, 0)`)
- Orphan cleanup happens on every `ListServers()` call -- validate PIDs, remove dead entries, rewrite file

### Known Projects: `~/.oraculo/projects.json`

Written by the desktop app. Lists projects the user has added via the launcher, independent of server status.

```json
[
  { "path": "/Users/lucas/dev/gastos", "name": "gastos-pessoais" },
  { "path": "/Users/lucas/dev/outro", "name": "outro-projeto" }
]
```

The desktop crosses both files: `projects.json` gives the full list, `servers.json` gives the status of each.

### Project Name Derivation

The `project` field in the registry comes from `.oraculo/config.json` if a `name` field exists, otherwise falls back to the basename of the project directory.

## Changes to Server (`oraculo`)

### Removes

- `assets.go` (`//go:embed all:dashboard_assets` + `DashboardAssets`)
- `dashboard_assets/` directory
- SPA handler: `newSPAHandler`, `withPlaceholders`, `spaShell`, `splitPath`
- Route `mux.Handle("GET /", newSPAHandler(...))`
- `staticPath` parameter from `server.New()` (currently a no-op -- logged but never alters behavior)
- `openBrowser()` and `ORACULO_NO_BROWSER` logic
- `POST /api/system/restart` (`handleRestart` with `syscall.Exec`) -- the desktop manages lifecycle via `StopServer`/`StartServer` bindings
- `assets_test.go`

Note: `oraculo restart` CLI command survives (useful for headless users without the desktop app). It continues using the kill-then-re-exec pattern.

### Adds

- Package `registry` in `apps/backend/src/registry/`:
  - `Register(project, path, port, pid)` -- called on server startup
  - `Unregister(path)` -- called on graceful shutdown via `defer`
  - Cross-platform file locking via `github.com/gofrs/flock`
  - Path: `~/.oraculo/servers.json`
- CORS headers on HTTP server:
  - `Access-Control-Allow-Origin: *` (all communication is local)
  - Apply to both HTTP responses and WebSocket upgrader (`CheckOrigin` returns `true`)
  - Wails v2.8+ uses `wails.localhost` on all platforms for HTTP requests
- `project_name` field in `GET /health` response (derived as described in "Project Name Derivation")

### Unchanged

- CLI commands (tools, hooks, install, setup, status, restart, kill, etc.)
- MCP server (stdio)
- WS hub and broadcast
- All API endpoints (including approval comments)
- Approval bridge
- DB, config, logging
- npm distribution (`@oraculo/cli`)

## Desktop App (Wails v2)

### Project Structure

```
apps/desktop/
├── main.go              -- wails.Run() with options
├── app.go               -- App struct with bindings
├── tray.go              -- systray (fyne-io/systray, maintained fork)
├── notifications.go     -- native notifications (beeep, cross-platform)
├── registry.go          -- reads ~/.oraculo/servers.json
├── ws_monitor.go        -- Go-side WS client for event monitoring
├── wails.json           -- Wails config
├── build/               -- icons, platform manifests
└── frontend/            -- symlink to apps/frontend/out/ (see wails.json config)
```

The `wails.json` `frontend:dir` points to `../frontend` and `frontend:build` runs `bun run build`. In production, `//go:embed all:frontend/dist` (or equivalent) embeds the static output.

### SPA Routing in Wails

The current server uses `withPlaceholders` and `spaShell` to handle Next.js dynamic routes and RSC payloads. Since the desktop embeds the same static export, it needs equivalent routing. This is implemented via Wails' `AssetServer.Handler` option -- a custom `http.Handler` that replicates the placeholder substitution logic:

1. Wails checks the embedded FS for an exact file match (built-in behavior)
2. If not found, the custom handler applies `withPlaceholders` to try `__placeholder__` paths
3. If still not found, falls back to the best-matching HTML shell via `spaShell`

The `withPlaceholders` and `spaShell` functions are extracted into a shared package (`apps/backend/src/spa/` or similar) so both the server (if ever re-enabled) and the desktop can use them. On initial migration, only the desktop uses them.

### Wails Bindings (Go -> JS)

| Binding | Description |
|---------|-------------|
| `ListServers()` | Read registry, validate PIDs, clean orphans, return project list with status |
| `StartServer(path)` | Spawn bundled `oraculo start http --dir <path>`, poll registry until entry appears (5s timeout, error on failure) |
| `StopServer(path)` | Execute bundled `oraculo kill --dir <path>` |
| `AddProject(path)` | Add to `~/.oraculo/projects.json` |
| `RemoveProject(path)` | Remove from `~/.oraculo/projects.json` |
| `GetCurrentServer()` | Return the base URL of the currently selected project (e.g., `http://localhost:3100`) |
| `GetNotificationSettings()` | Read notification preferences |
| `SetNotificationSettings(...)` | Save notification preferences |

### Launcher Screen

Lists all known projects with contextual actions by state:

| Project State | Visible Actions |
|---------------|-----------------|
| Offline | Start, Remove |
| Online | Open Dashboard, Stop |

"Add Project" button with native file picker (Wails dialog API).

### Navigation and Server Context Injection

On selecting an active project, the frontend loads pointing to that server. The URL injection works via Wails binding:

1. User clicks "Open Dashboard" on a project
2. Desktop stores the selected server URL internally
3. Frontend JS calls `GetCurrentServer()` binding on mount
4. Returns `http://localhost:<port>` which populates a React `ServerContext`
5. All `fetch()` calls in `api.ts` and the WS connection in `useWebSocket` use this base URL
6. In dev (browser), `GetCurrentServer()` is not available -- falls back to `window.location.origin`

### System Tray (fyne-io/systray)

- Started in Wails `OnStartup`
- Close window -> `Hide()` instead of `Quit()` (via `OnBeforeClose` hook)
- Click tray -> `Show()`
- Menu: list of active projects, separator, "Open Window", "Quit"
- Visual badge when approvals are pending

### Native Notifications (beeep)

- Cross-platform via `gen2brain/beeep` (macOS, Windows, Linux)
- Desktop maintains Go-side WS connections to each active server (in `ws_monitor.go`)
  - Why Go-side: notifications must work even with the window hidden (tray mode)
  - One goroutine per active server, managed by connection lifecycle
- Events that trigger notifications:
  - `approval_requested` -- approval pending (critical, agent blocked)
  - `story_completed` / `epic_completed` -- completion notifications
- Clicking a notification opens the window in the correct context

### Bundled Binary

- Build process includes the `oraculo` binary in the desktop app bundle (platform-specific location)
- Desktop always uses the bundled binary for `StartServer`/`StopServer`
- If the user also has `oraculo` installed via npm, both coexist independently (npm version for Claude Code MCP, bundled for desktop-managed servers)

## Changes to Frontend (`apps/frontend/`)

### Rename

- `apps/dashboard/` -> `apps/frontend/`
- Update references in Makefile, CLAUDE.md, imports, CI

### Dynamic Server Context

Today the frontend assumes same-origin (`/api/...`, `/ws`). With the desktop, each project has a server on a different port.

- Add a React `ServerContext` with a `baseUrl` field
- On mount, call `GetCurrentServer()` Wails binding if available
- Fallback to `window.location.origin` when running in a regular browser
- `api.ts`: all fetch calls use `${baseUrl}/api/...` instead of `/api/...`
- `useWebSocket`: connect to `${baseUrl.replace('http', 'ws')}/ws` instead of `ws://${window.location.host}/ws`

### Unchanged

- Components, pages, styles
- Build: `output: "export"` (static)
- `next.config.ts` dev proxy (still useful for standalone browser development)

## Build and Distribution

### Build Pipeline

```bash
# 1. Frontend
cd apps/frontend && bun run build          # -> apps/frontend/out/

# 2. Backend (no frontend embed)
cd apps/backend && go build ./cmd/oraculo  # -> oraculo binary

# 3. Desktop (embeds frontend + bundles oraculo)
cd apps/desktop && wails build             # -> build/bin/oraculo-desktop
```

### Development Workflow

- **Frontend dev (browser):** `cd apps/frontend && bun dev` -- runs Next.js dev server with proxy to `localhost:6077`. Start `oraculo start http` separately. Same as today.
- **Desktop dev:** `cd apps/desktop && wails dev` -- Wails runs the frontend dev server and opens the WebView. Requires an `oraculo` server running separately for API/WS.
- **Backend dev:** unchanged, `go test ./...` etc.

### Distribution Channels

| Channel | Delivers | Audience |
|---------|----------|----------|
| npm (`@oraculo/cli`) | `oraculo` binary (CLI/server) | Claude Code plugin users |
| Desktop installer (`.dmg`, `.exe`) | `oraculo-desktop` + `oraculo` bundled | Users wanting native interface |

### Makefile Updates

- `build-frontend` -- builds Next.js
- `build-backend` -- compiles Go (no dashboard embed)
- `build-desktop` -- builds frontend + wails build
- Remove targets that copy assets to backend

### CI

- Existing npm publish workflow unchanged (CLI binary only)
- New workflow for desktop app (GitHub releases with `.dmg`/`.exe`)

## Follow-up Items

- Design the app icon
- Design the launcher screen in `docs/ui/dashboard.pen`
