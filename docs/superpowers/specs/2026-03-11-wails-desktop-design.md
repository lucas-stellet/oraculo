# Wails Desktop App Design

## Summary

Split Oraculo into two binaries: `oraculo` (CLI/server) and `oraculo-desktop` (Wails v3 native app). The desktop app is a multi-window launcher and viewer that connects to multiple project servers via HTTP/WS. The server drops embedded frontend assets; the desktop embeds the Next.js build and bundles the `oraculo` binary.

## Motivation

- **Native window** -- desktop app in the dock/taskbar, not a browser tab
- **Distribution** -- installable app (`.dmg`, `.exe`) with bundled server binary
- **Native features** -- built-in system tray, native notifications, multi-window support

## Architecture

Two binaries, three apps in the monorepo:

```
apps/
├── backend/     -- Go: CLI + HTTP + WS + MCP (oraculo binary, no frontend embed)
├── frontend/    -- Next.js (rename from dashboard, static export)
└── desktop/     -- Wails v3: embeds frontend/out/, bundles oraculo binary
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
- `project_name` field in `GET /health` response (derived as described in "Project Name Derivation")

### Unchanged

- CLI commands (tools, hooks, install, setup, status, restart, kill, etc.)
- MCP server (stdio)
- WS hub and broadcast
- All API endpoints (including approval comments)
- Approval bridge
- DB, config, logging
- npm distribution (`@oraculo/cli`)

## Desktop App (Wails v3)

### Why v3 over v2

- **System tray built-in** -- `app.SystemTray.New()` replaces external `fyne-io/systray` dependency and avoids macOS main thread conflicts
- **Multi-window native** -- each project can open its own dashboard window via `app.NewWebviewWindowWithOptions()`
- **Service pattern** -- dependency injection replaces global context threading, more testable
- **Better build system** -- `Taskfile.yml` replaces `wails.json`, transparent and customizable
- **Auto-generated TypeScript bindings** -- `wails3 generate bindings` produces typed TS functions in `./bindings/`
- **Tray-attached windows** -- `tray.AttachWindow(win)` auto-toggles window near tray icon with focus-loss hiding

v3 is in alpha (v3.0.0-alpha.68+) with daily releases. Acceptable risk for an internal tool used by ~5 people.

### Project Structure

```
apps/desktop/
├── main.go              -- application.New() + window creation + app.Run()
├── services/
│   ├── launcher.go      -- LauncherService: ListProjects, AddProject, RemoveProject, StartServer, StopServer
│   ├── server.go        -- ServerService: GetCurrentServer, SelectServer
│   └── notifications.go -- NotificationService: GetSettings, SetSettings
├── tray.go              -- built-in system tray (app.SystemTray.New())
├── ws_monitor.go        -- Go-side WS client for event monitoring
├── spa.go               -- SPA routing for asset server (reuses withPlaceholders/spaShell)
├── Taskfile.yml          -- Wails v3 build tasks (replaces wails.json)
├── build/               -- icons, platform manifests
└── frontend/            -- copied from apps/frontend/out/ at build time
```

### Application Lifecycle

```go
app := application.New(application.Options{
    Name: "Oraculo",
    Services: []application.Service{
        application.NewService(&LauncherService{}),
        application.NewService(&ServerService{}),
        application.NewService(&NotificationService{}),
    },
    Assets: application.AssetOptions{
        Handler: NewSPAHandler(assets),
    },
    Mac: application.MacOptions{
        TerminateOnLastWindowClosed: false, // tray keeps running
    },
})

// Create launcher window
app.NewWebviewWindowWithOptions(application.WebviewWindowOptions{
    Title: "Oraculo", Name: "launcher",
    Width: 900, Height: 600, URL: "/",
})

app.Run()
```

### SPA Routing in Wails v3

In v3, `Assets.Handler` serves assets and `Assets.Middleware` wraps the handler chain. The Wails runtime serves internal routes at `/wails/*` (runtime JS, IPC transport). A custom middleware must forward `/wails/*` to the next handler, then apply SPA routing for all other requests:

1. Forward `/wails/*` to `next` (Wails runtime — required for bindings and events)
2. Try the exact file from the embedded FS
3. If not found, try placeholder substitution for dynamic routes
4. If still not found, serve the SPA shell

`BundledAssetFileServer(assets)` is used as the base `Handler` (it also serves `/wails/runtime.js`). The `withPlaceholders` and `spaShell` functions are extracted into a shared package (`apps/backend/src/spa/`) for reuse by the desktop.

### Wails Services (Go -> JS)

Services replace v2's `Bind` pattern. Each service is a Go struct whose public methods are auto-exposed to the frontend via `wails3 generate bindings`.

**LauncherService:**

| Method | Description |
|--------|-------------|
| `ListProjects()` | Read projects.json + servers.json, validate PIDs, return merged list with status |
| `AddProject()` | Open native directory picker, validate `.oraculo/` exists, add to projects.json |
| `RemoveProject(path)` | Remove from projects.json |
| `StartServer(path)` | Spawn bundled `oraculo start http --dir <path>`, poll health until ready (5s timeout) |
| `StopServer(path)` | Execute bundled `oraculo kill --dir <path>` |

**ServerService:**

| Method | Description |
|--------|-------------|
| `SelectServer(port)` | Store selected server URL, return base URL |
| `GetCurrentServer()` | Return base URL of currently selected project |

### Multi-Window Architecture

The launcher is the main window. When the user clicks "Open Dashboard" on an active project:

1. `SelectServer(port)` stores the active server URL
2. The frontend navigates to the dashboard view (same window, React routing)
3. Future enhancement: open a **new window** per project via `app.NewWebviewWindowWithOptions()`

Each window can load a different URL route. The Go backend knows which window called a method via `ctx.Value(application.WindowNameKey)`.

### Launcher Screen

Lists all known projects with contextual actions by state:

| Project State | Visible Actions |
|---------------|-----------------|
| Offline | Start, Remove |
| Online | Open Dashboard, Stop |

"Add Project" button with native file picker (Wails dialog API via `application.OpenDirectoryDialog`).

### Navigation and Server Context Injection

1. User clicks "Open Dashboard" on a project
2. Frontend calls `SelectServer(port)` binding, receives `http://localhost:<port>`
3. This populates a React `ServerContext`
4. All `fetch()` calls in `api.ts` and the WS connection use this base URL
5. In dev (browser), `window.wails` is not available -- falls back to `window.location.origin`

### System Tray (built-in)

Wails v3 provides native system tray support:

```go
tray := app.SystemTray.New()
tray.SetIcon(iconBytes)
tray.SetDarkModeIcon(darkIconBytes)
tray.SetTooltip("Oraculo Desktop")
```

- Click tray icon -> show/toggle launcher window
- Menu: list of active projects, separator, "Open Window", "Quit"
- Dynamic icon switching for pending approvals (no native badge API -- swap icon instead)
- Hide-on-close via `RegisterHook(events.Common.WindowClosing)` + `e.Cancel()` + `window.Hide()`

### Native Notifications (beeep)

- Cross-platform via `gen2brain/beeep` (macOS, Windows, Linux) -- Wails v3 has no built-in notification API
- Desktop maintains Go-side WS connections to each active server (in `ws_monitor.go`)
  - Why Go-side: notifications must work even with the window hidden (tray mode)
  - One goroutine per active server, managed by connection lifecycle
- Events that trigger notifications:
  - `approval_requested` -- approval pending (critical, agent blocked)
  - `story_completed` / `epic_completed` -- completion notifications

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
- On mount, check `window.wails` to detect Wails environment
- If in Wails: call `GetCurrentServer()` binding to get the base URL
- If in browser: fall back to `window.location.origin`
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
cd apps/desktop && wails3 build            # -> build/bin/oraculo-desktop
```

### Development Workflow

- **Frontend dev (browser):** `cd apps/frontend && bun dev` -- runs Next.js dev server with proxy to `localhost:6077`. Start `oraculo start http` separately. Same as today.
- **Desktop dev:** `cd apps/desktop && wails3 dev` -- Wails runs the Vite dev server on port 5173 with HMR. Requires an `oraculo` server running separately for API/WS.
- **Backend dev:** unchanged, `go test ./...` etc.

### Distribution Channels

| Channel | Delivers | Audience |
|---------|----------|----------|
| npm (`@oraculo/cli`) | `oraculo` binary (CLI/server) | Claude Code plugin users |
| Desktop installer (`.dmg`, `.exe`) | `oraculo-desktop` + `oraculo` bundled | Users wanting native interface |

### Makefile Updates

- `build-frontend` -- builds Next.js
- `build-backend` -- compiles Go (no dashboard embed)
- `build-desktop` -- builds frontend + wails3 build
- Remove targets that copy assets to backend

### CI

- Existing npm publish workflow unchanged (CLI binary only)
- New workflow for desktop app (GitHub releases with `.dmg`/`.exe`)

## Follow-up Items

- Design the app icon
- Design the launcher screen in `docs/ui/dashboard.pen`
