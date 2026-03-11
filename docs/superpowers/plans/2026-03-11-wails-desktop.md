# Wails Desktop App Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Split Oraculo into two binaries — `oraculo` (headless CLI/server) and `oraculo-desktop` (Wails v3 native app) — connected via HTTP/WS, with server registry, system tray, and native notifications.

**Architecture:** The `oraculo` backend drops its embedded frontend and gains a global server registry (`~/.oraculo/servers.json`). The Next.js dashboard is renamed to `apps/frontend/` and gains a `ServerContext` for dynamic base URLs. A new `apps/desktop/` Wails v3 app embeds the frontend build, connects to multiple project servers, and provides native desktop features (tray, notifications, multi-window).

**Tech Stack:** Go 1.24, Wails v3 alpha (`github.com/wailsapp/wails/v3`), Next.js (static export), `github.com/gofrs/flock`, `gen2brain/beeep`

**Spec:** `docs/superpowers/specs/2026-03-11-wails-desktop-design.md`

---

## File Structure

### New files

- `apps/backend/src/registry/registry.go` — Server registry (register/unregister/list with file locking)
- `apps/backend/src/registry/registry_test.go` — Registry tests
- `apps/backend/src/server/cors.go` — CORS middleware
- `apps/backend/src/server/cors_test.go` — CORS tests
- `apps/backend/src/spa/spa.go` — Extracted SPA routing (withPlaceholders/spaShell)
- `apps/backend/src/spa/spa_test.go` — SPA routing tests
- `apps/frontend/src/lib/server-context.tsx` — React context for dynamic server URL
- `apps/desktop/main.go` — Wails v3 entrypoint (application.New + app.Run)
- `apps/desktop/services/launcher.go` — LauncherService (ListProjects, AddProject, RemoveProject, StartServer, StopServer)
- `apps/desktop/services/server.go` — ServerService (SelectServer, GetCurrentServer)
- `apps/desktop/services/notifications.go` — NotificationService (settings)
- `apps/desktop/tray.go` — System tray (built-in v3 API)
- `apps/desktop/notifications.go` — Native notification dispatch (beeep)
- `apps/desktop/ws_monitor.go` — Go-side WS client for event monitoring
- `apps/desktop/spa.go` — SPA routing handler for Wails asset server
- `apps/desktop/detach_unix.go` — Unix process detachment (build-tagged)
- `apps/desktop/detach_windows.go` — Windows process detachment (build-tagged)
- `apps/desktop/go.mod` — Separate Go module
- `apps/desktop/Taskfile.yml` — Wails v3 build tasks
- `apps/desktop/wails.json` — Wails config
- `apps/desktop/build/` — Icons, platform manifests

### Modified files

- `apps/backend/src/server/server.go` — Remove SPA handler, staticPath param; add CORS middleware
- `apps/backend/src/server/assets.go` — Delete entirely
- `apps/backend/src/server/assets_test.go` — Delete entirely
- `apps/backend/src/server/system.go` — Remove handleRestart; keep handleStatus; add project_name to health
- `apps/backend/src/cli/start.go` — Remove openBrowser/ORACULO_NO_BROWSER; add registry calls
- `apps/backend/src/cli/restart.go` — Remove ORACULO_NO_BROWSER env
- `apps/backend/go.mod` — Add `github.com/gofrs/flock`
- `apps/frontend/src/lib/api.ts` — Use ServerContext base URL
- `apps/frontend/src/lib/ws.tsx` — Use ServerContext base URL
- `apps/frontend/src/app/page.tsx` — Use api.ts instead of direct fetch
- `apps/frontend/next.config.ts` — Rename references
- `Makefile` — New targets, remove dashboard_assets copy
- `CLAUDE.md` — Update paths and descriptions

### Deleted files

- `apps/backend/src/server/assets.go`
- `apps/backend/src/server/assets_test.go`
- `apps/backend/src/server/dashboard_assets/` (generated directory, may not exist in repo)

---

## Chunk 1: Backend — Registry, CORS, Health, Cleanup

### Task 1: Create registry package with Register/Unregister

**Files:**
- Create: `apps/backend/src/registry/registry.go`
- Create: `apps/backend/src/registry/registry_test.go`
- Modify: `apps/backend/go.mod` (add `github.com/gofrs/flock`)

- [ ] **Step 1: Add flock dependency**

```bash
cd apps/backend && go get github.com/gofrs/flock
```

- [ ] **Step 2: Write failing test for Register**

Create `apps/backend/src/registry/registry_test.go`:

```go
package registry_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/lucas/oraculo/apps/backend/src/registry"
)

func TestRegister_CreatesFileAndAddsEntry(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "servers.json")

	err := registry.Register(path, registry.Entry{
		Project:   "test-project",
		Path:      "/tmp/test",
		Port:      3100,
		PID:       os.Getpid(),
	})
	if err != nil {
		t.Fatalf("Register: %v", err)
	}

	entries, err := registry.List(path)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Project != "test-project" {
		t.Errorf("expected project 'test-project', got %q", entries[0].Project)
	}
	if entries[0].Port != 3100 {
		t.Errorf("expected port 3100, got %d", entries[0].Port)
	}
}
```

- [ ] **Step 3: Run test to verify it fails**

```bash
cd apps/backend && go test ./src/registry/ -v -run TestRegister_CreatesFileAndAddsEntry
```
Expected: FAIL — package does not exist

- [ ] **Step 4: Implement registry package**

Create `apps/backend/src/registry/registry.go`:

```go
package registry

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"time"

	"github.com/gofrs/flock"
)

// Entry represents a running oraculo server instance.
type Entry struct {
	Project   string    `json:"project"`
	Path      string    `json:"path"`
	Port      int       `json:"port"`
	PID       int       `json:"pid"`
	StartedAt time.Time `json:"started_at"`
}

// Register adds or updates an entry in the registry file.
func Register(registryPath string, entry Entry) error {
	if entry.StartedAt.IsZero() {
		entry.StartedAt = time.Now()
	}
	return withLock(registryPath, func(entries []Entry) ([]Entry, error) {
		for i, e := range entries {
			if e.Path == entry.Path {
				entries[i] = entry
				return entries, nil
			}
		}
		return append(entries, entry), nil
	})
}

// Unregister removes the entry matching the given project path.
func Unregister(registryPath string, projectPath string) error {
	return withLock(registryPath, func(entries []Entry) ([]Entry, error) {
		for i, e := range entries {
			if e.Path == projectPath {
				return append(entries[:i], entries[i+1:]...), nil
			}
		}
		return entries, nil
	})
}

// List reads all entries from the registry file.
func List(registryPath string) ([]Entry, error) {
	data, err := os.ReadFile(registryPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	var entries []Entry
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, err
	}
	return entries, nil
}

// WriteAll atomically replaces all entries in the registry file.
func WriteAll(registryPath string, entries []Entry) error {
	return withLock(registryPath, func(_ []Entry) ([]Entry, error) {
		return entries, nil
	})
}

// DefaultPath returns ~/.oraculo/servers.json.
func DefaultPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".oraculo", "servers.json"), nil
}

func withLock(registryPath string, fn func([]Entry) ([]Entry, error)) error {
	if err := os.MkdirAll(filepath.Dir(registryPath), 0o755); err != nil {
		return err
	}

	lock := flock.New(registryPath + ".lock")
	if err := lock.Lock(); err != nil {
		return err
	}
	defer lock.Unlock()

	entries, err := List(registryPath)
	if err != nil {
		return err
	}

	entries, err = fn(entries)
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')

	return os.WriteFile(registryPath, data, 0o644)
}
```

- [ ] **Step 5: Run test to verify it passes**

```bash
cd apps/backend && go test ./src/registry/ -v -run TestRegister_CreatesFileAndAddsEntry
```
Expected: PASS

- [ ] **Step 6: Write remaining registry tests**

Append to `registry_test.go`:

```go
func TestUnregister_RemovesEntry(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "servers.json")

	_ = registry.Register(path, registry.Entry{Project: "a", Path: "/a", Port: 3100, PID: os.Getpid()})
	_ = registry.Register(path, registry.Entry{Project: "b", Path: "/b", Port: 3101, PID: os.Getpid()})

	err := registry.Unregister(path, "/a")
	if err != nil {
		t.Fatalf("Unregister: %v", err)
	}

	entries, _ := registry.List(path)
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Path != "/b" {
		t.Errorf("expected /b, got %q", entries[0].Path)
	}
}

func TestRegister_UpdatesExistingEntry(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "servers.json")

	_ = registry.Register(path, registry.Entry{Project: "a", Path: "/a", Port: 3100, PID: 1})
	_ = registry.Register(path, registry.Entry{Project: "a", Path: "/a", Port: 3200, PID: 2})

	entries, _ := registry.List(path)
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Port != 3200 {
		t.Errorf("expected port 3200, got %d", entries[0].Port)
	}
}

func TestList_ReturnsNilForMissingFile(t *testing.T) {
	entries, err := registry.List("/nonexistent/servers.json")
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if entries != nil {
		t.Errorf("expected nil, got %v", entries)
	}
}
```

- [ ] **Step 7: Run all registry tests**

```bash
cd apps/backend && go test ./src/registry/ -v
```
Expected: All PASS

- [ ] **Step 8: Commit**

```bash
git add apps/backend/src/registry/ apps/backend/go.mod apps/backend/go.sum
git commit -m "feat(registry): add server registry package with file locking"
```

### Task 2: Wire registry into server startup/shutdown

**Files:**
- Modify: `apps/backend/src/cli/start.go`
- Modify: `apps/backend/src/config/config.go`

- [ ] **Step 1: Add Name field to Config and ProjectName helper**

In `apps/backend/src/config/config.go`, add `Name` field to the `Config` struct and a `ProjectName()` method:

```go
type Config struct {
	Port              int         `json:"port"`
	Name              string      `json:"name,omitempty"`
	PreferredLanguage string      `json:"preferred_language,omitempty"`
	Skills            AgentSkills `json:"skills,omitempty"`
}

// ProjectName returns the configured name, or falls back to the working directory basename.
func (c *Config) ProjectName() string {
	if c.Name != "" {
		return c.Name
	}
	wd, err := os.Getwd()
	if err != nil {
		return "unknown"
	}
	return filepath.Base(wd)
}
```

- [ ] **Step 2: Write test for ProjectName**

Add to `apps/backend/src/config/config_test.go` (create if needed):

```go
func TestConfig_ProjectName_UsesConfiguredName(t *testing.T) {
	cfg := &Config{Name: "my-project"}
	if got := cfg.ProjectName(); got != "my-project" {
		t.Errorf("expected 'my-project', got %q", got)
	}
}

func TestConfig_ProjectName_FallsBackToBasename(t *testing.T) {
	cfg := &Config{}
	name := cfg.ProjectName()
	if name == "" || name == "unknown" {
		t.Errorf("expected a non-empty basename, got %q", name)
	}
}
```

Run: `cd apps/backend && go test ./src/config/ -v -run TestConfig_ProjectName`

- [ ] **Step 3: Add registry register/unregister to runStartAll and runStartHTTP**

In `apps/backend/src/cli/start.go`, after the server starts listening, register with the registry. Add defer unregister:

```go
// Register in global server registry.
regPath, regErr := registry.DefaultPath()
if regErr == nil {
	wd, _ := os.Getwd()
	_ = registry.Register(regPath, registry.Entry{
		Project: cfg.ProjectName(),
		Path:    wd,
		Port:    port,
		PID:     os.Getpid(),
	})
	defer registry.Unregister(regPath, wd)
}
```

Add the import: `"github.com/lucas/oraculo/apps/backend/src/registry"`

Apply the same pattern to both `runStartAll` and `runStartHTTP`.

- [ ] **Step 4: Run existing tests**

```bash
cd apps/backend && go test ./... -count=1
```
Expected: All PASS

- [ ] **Step 5: Commit**

```bash
git add apps/backend/src/cli/start.go apps/backend/src/config/
git commit -m "feat(cli): register/unregister server in global registry on start/stop"
```

### Task 3: Add CORS middleware

**Files:**
- Create: `apps/backend/src/server/cors.go`
- Create: `apps/backend/src/server/cors_test.go`
- Modify: `apps/backend/src/server/server.go`

- [ ] **Step 1: Write failing test for CORS headers**

Create `apps/backend/src/server/cors_test.go`:

```go
package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCORS_SetsHeaders(t *testing.T) {
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	handler := corsMiddleware(inner)

	req := httptest.NewRequest("GET", "/api/epics", nil)
	req.Header.Set("Origin", "http://wails.localhost")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "*" {
		t.Errorf("expected ACAO *, got %q", got)
	}
}

func TestCORS_HandlesPreflight(t *testing.T) {
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	handler := corsMiddleware(inner)

	req := httptest.NewRequest("OPTIONS", "/api/epics", nil)
	req.Header.Set("Origin", "http://wails.localhost")
	req.Header.Set("Access-Control-Request-Method", "POST")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}
	if got := rec.Header().Get("Access-Control-Allow-Methods"); got == "" {
		t.Error("expected Access-Control-Allow-Methods header")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
cd apps/backend && go test ./src/server/ -v -run TestCORS
```
Expected: FAIL — `corsMiddleware` undefined

- [ ] **Step 3: Implement CORS middleware**

Create `apps/backend/src/server/cors.go`:

```go
package server

import "net/http"

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
```

- [ ] **Step 4: Run CORS tests**

```bash
cd apps/backend && go test ./src/server/ -v -run TestCORS
```
Expected: PASS

- [ ] **Step 5: Wire CORS into server constructor**

In `apps/backend/src/server/server.go`, add a `handler` field to Server and wrap the mux once in the constructor:

```go
type Server struct {
	mux          *http.ServeMux
	handler      http.Handler
	database     *db.DB
	lastActivity time.Time
	mu           sync.Mutex
}
```

At end of `New()`: `s.handler = corsMiddleware(mux)`

Update `ServeHTTP`: `s.handler.ServeHTTP(w, r)`

- [ ] **Step 6: Run all server tests**

```bash
cd apps/backend && go test ./src/server/ -v
```
Expected: All PASS

- [ ] **Step 7: Commit**

```bash
git add apps/backend/src/server/cors.go apps/backend/src/server/cors_test.go apps/backend/src/server/server.go
git commit -m "feat(server): add CORS middleware for desktop cross-origin access"
```

### Task 4: Add project_name to health endpoint

**Files:**
- Modify: `apps/backend/src/server/server.go`

- [ ] **Step 1: Write failing test**

```go
func TestHealth_IncludesProjectName(t *testing.T) {
	srv := New(nil, nil, nil, nil, "test-project", "test")
	req := httptest.NewRequest("GET", "/health", nil)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	var body map[string]string
	json.NewDecoder(rec.Body).Decode(&body)
	if body["project_name"] != "test-project" {
		t.Errorf("expected project_name 'test-project', got %q", body["project_name"])
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
cd apps/backend && go test ./src/server/ -v -run TestHealth_IncludesProjectName
```

- [ ] **Step 3: Modify handleHealth to include project_name**

Update `New()` signature to accept `projectName string` (replacing the old `staticPath` parameter). Replace `handleHealth` with a closure that includes `project_name` in the JSON response.

Update callers in `start.go` to pass `cfg.ProjectName()`.

- [ ] **Step 4: Run all tests**

```bash
cd apps/backend && go test ./... -count=1
```
Expected: All PASS

- [ ] **Step 5: Commit**

```bash
git add apps/backend/src/server/server.go apps/backend/src/cli/start.go
git commit -m "feat(server): add project_name to health endpoint"
```

### Task 5: Remove embedded frontend from server

**Files:**
- Delete: `apps/backend/src/server/assets.go`
- Delete: `apps/backend/src/server/assets_test.go`
- Modify: `apps/backend/src/server/server.go`
- Modify: `apps/backend/src/server/system.go`
- Modify: `apps/backend/src/cli/start.go`
- Modify: `apps/backend/src/cli/restart.go`

- [ ] **Step 1: Delete assets.go and assets_test.go**

```bash
rm apps/backend/src/server/assets.go apps/backend/src/server/assets_test.go
```

- [ ] **Step 2: Remove SPA handler from server.go**

Remove: `newSPAHandler`, `withPlaceholders`, `spaShell`, `splitPath`, the `"GET /"` route, `staticPath` field, and unused imports.

- [ ] **Step 3: Remove openBrowser and ORACULO_NO_BROWSER from start.go**

Remove the `openBrowser` function and the browser-opening goroutines in `runStartAll` and `runStartHTTP`. Remove unused imports (`"os/exec"`, `"runtime"`).

- [ ] **Step 4: Remove handleRestart from system.go**

Remove `handleRestart` method and `"syscall"` import. Remove the route in `server.go`.

- [ ] **Step 5: Remove ORACULO_NO_BROWSER from restart.go**

Change `Env: append(os.Environ(), "ORACULO_NO_BROWSER=1")` to `Env: os.Environ()`.

- [ ] **Step 6: Run all backend tests**

```bash
cd apps/backend && go test ./... -count=1
```
Expected: All PASS

- [ ] **Step 7: Commit**

```bash
git add -A apps/backend/src/server/ apps/backend/src/cli/start.go apps/backend/src/cli/restart.go
git commit -m "refactor(server): remove embedded frontend, SPA handler, and browser auto-open"
```

### Task 6: Extract SPA routing to shared package

**Files:**
- Create: `apps/backend/src/spa/spa.go`
- Create: `apps/backend/src/spa/spa_test.go`

- [ ] **Step 1: Write test for WithPlaceholders and Shell**

Create `apps/backend/src/spa/spa_test.go`:

```go
package spa_test

import (
	"testing"

	"github.com/lucas/oraculo/apps/backend/src/spa"
)

func TestWithPlaceholders(t *testing.T) {
	tests := []struct {
		input, want string
	}{
		{"epics/gastos/approvals", "epics/__placeholder__/approvals"},
		{"epics/__placeholder__/approvals", "epics/__placeholder__/approvals"},
		{"epics/gastos/approvals/abc-123/review", "epics/__placeholder__/approvals/__placeholder__/review"},
		{"epics/gastos/stories/registro", "epics/__placeholder__/stories/__placeholder__"},
		{"other/path", "other/path"},
		{"epics", "epics"},
	}
	for _, tt := range tests {
		got := spa.WithPlaceholders(tt.input)
		if got != tt.want {
			t.Errorf("WithPlaceholders(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestShell(t *testing.T) {
	tests := []struct {
		path  string
		isRSC bool
		want  string
	}{
		{"/epics/gastos/approvals/abc/review", false, "/epics/__placeholder__/approvals/__placeholder__/review.html"},
		{"/epics/gastos/approvals", false, "/epics/__placeholder__/approvals.html"},
		{"/epics/gastos/stories/registro", false, "/epics/__placeholder__/stories/__placeholder__.html"},
		{"/epics/gastos", false, "/epics/__placeholder__.html"},
		{"/epics/gastos", true, "/epics/__placeholder__.txt"},
		{"/other", false, "/"},
	}
	for _, tt := range tests {
		got := spa.Shell(tt.path, tt.isRSC)
		if got != tt.want {
			t.Errorf("Shell(%q, %v) = %q, want %q", tt.path, tt.isRSC, got, tt.want)
		}
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
cd apps/backend && go test ./src/spa/ -v
```

- [ ] **Step 3: Create spa package**

Create `apps/backend/src/spa/spa.go` with the `WithPlaceholders` and `Shell` functions extracted from the deleted server.go code (renamed to exported). Include `splitPath` as unexported helper.

- [ ] **Step 4: Run tests**

```bash
cd apps/backend && go test ./src/spa/ -v
```
Expected: All PASS

- [ ] **Step 5: Commit**

```bash
git add apps/backend/src/spa/
git commit -m "refactor(spa): extract SPA routing to shared package for desktop reuse"
```

---

## Chunk 2: Frontend — Rename and ServerContext

### Task 7: Rename dashboard to frontend

**Files:**
- Rename: `apps/dashboard/` → `apps/frontend/`
- Modify: `Makefile`
- Modify: `CLAUDE.md`

- [ ] **Step 1: Rename the directory**

```bash
git mv apps/dashboard apps/frontend
```

- [ ] **Step 2: Update Makefile**

Replace `DASHBOARD_DIR := ./apps/dashboard` with `FRONTEND_DIR := ./apps/frontend`. Replace all `$(DASHBOARD_DIR)` with `$(FRONTEND_DIR)`. Remove the `dashboard_assets` copy steps from `build` and `rebuild` targets. Simplify:

```makefile
build-frontend:
	cd $(FRONTEND_DIR) && bun run build

build-backend:
	go build -ldflags "$(LDFLAGS)" -o $(BINARY) $(BUILD)

build: build-backend
```

Update `clean` and `web-dev` to reference `$(FRONTEND_DIR)`.

- [ ] **Step 3: Update CLAUDE.md references**

Replace all occurrences of `apps/dashboard` with `apps/frontend`.

- [ ] **Step 4: Verify frontend and backend still build**

```bash
cd apps/frontend && bun run build
cd apps/backend && go build ./cmd/oraculo
```

- [ ] **Step 5: Commit**

```bash
git add -A
git commit -m "refactor: rename apps/dashboard to apps/frontend"
```

### Task 8: Add ServerContext to frontend

**Files:**
- Create: `apps/frontend/src/lib/server-context.tsx`
- Modify: `apps/frontend/src/lib/api.ts`
- Modify: `apps/frontend/src/lib/ws.tsx`
- Modify: `apps/frontend/src/app/layout.tsx`
- Modify: `apps/frontend/src/app/page.tsx`
- Modify: `apps/frontend/src/app/epics/[id]/layout.tsx`
- Modify: 6 files that import `{ api }` from `@/lib/api`

- [ ] **Step 1: Create ServerContext provider**

Create `apps/frontend/src/lib/server-context.tsx`:

```tsx
"use client";

import { createContext, useContext, useMemo } from "react";
import { createApi } from "./api";

interface ServerContextValue {
  baseUrl: string;
}

const ServerContext = createContext<ServerContextValue>({
  baseUrl: "",
});

export function ServerProvider({
  baseUrl,
  children,
}: {
  baseUrl?: string;
  children: React.ReactNode;
}) {
  const value = useMemo(
    () => ({
      baseUrl: baseUrl || (typeof window !== "undefined" ? window.location.origin : ""),
    }),
    [baseUrl]
  );

  return (
    <ServerContext.Provider value={value}>{children}</ServerContext.Provider>
  );
}

export function useServerUrl(): string {
  return useContext(ServerContext).baseUrl;
}

export function useApi() {
  const baseUrl = useServerUrl();
  return useMemo(() => createApi(baseUrl), [baseUrl]);
}
```

- [ ] **Step 2: Refactor api.ts to accept a base URL**

Change `fetchJSON` to accept `baseUrl` parameter. Add `createApi(baseUrl)` factory. Keep `export const api = createApi("")` for backward compatibility.

```typescript
export function createApi(baseUrl: string) {
  return {
    listEpics: () => fetchJSON<EpicSummary[]>(baseUrl, "/api/epics"),
    // ... all methods use ${baseUrl}${path}
  };
}

export const api = createApi("");
```

- [ ] **Step 3: Refactor ws.tsx to accept serverUrl prop**

Add `serverUrl` prop to `WebSocketProvider`. Use it to construct the WS URL:

```typescript
export function WebSocketProvider({ serverUrl, children }: { serverUrl?: string; children: React.ReactNode }) {
  // ...
  const connect = useCallback(() => {
    let wsBase: string;
    if (serverUrl) {
      wsBase = serverUrl.replace(/^http/, "ws");
    } else {
      const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
      wsBase = `${protocol}//${window.location.host}`;
    }
    const url = `${wsBase}/ws`;
    // ...
  }, [serverUrl]);
}
```

- [ ] **Step 4: Wire ServerProvider into root layout**

In `apps/frontend/src/app/layout.tsx`, wrap children with `<ServerProvider>`. Do NOT add WebSocketProvider here — it stays in the epic layout.

- [ ] **Step 5: Pass serverUrl to WebSocketProvider in epic layout**

In `apps/frontend/src/app/epics/[id]/layout.tsx`, use `useServerUrl()` and pass it to `<WebSocketProvider serverUrl={serverUrl}>`.

- [ ] **Step 6: Update all components using `{ api }` import**

Replace `import { api } from "@/lib/api"` with `import { useApi } from "@/lib/server-context"` in these files:
1. `apps/frontend/src/app/page.tsx`
2. `apps/frontend/src/app/epics/[id]/_client.tsx`
3. `apps/frontend/src/app/epics/[id]/layout.tsx`
4. `apps/frontend/src/app/epics/[id]/approvals/_client.tsx`
5. `apps/frontend/src/app/epics/[id]/stories/[storyId]/_client.tsx`
6. `apps/frontend/src/app/epics/[id]/approvals/[approvalId]/review/_client.tsx`

Add `const api = useApi();` inside each component body.

- [ ] **Step 7: Remove restart feature from frontend**

In `apps/frontend/src/app/epics/[id]/layout.tsx`, remove the restart banner and `restartServer` call. Remove `restartServer` from `api.ts`.

- [ ] **Step 8: Verify frontend builds**

```bash
cd apps/frontend && bun run build
```

- [ ] **Step 9: Commit**

```bash
git add apps/frontend/src/
git commit -m "feat(frontend): add ServerContext for dynamic server URLs; remove restart feature"
```

---

## Chunk 3: Desktop App — Wails v3 Setup

### Task 9: Initialize Wails v3 project

**Files:**
- Create: `apps/desktop/main.go`
- Create: `apps/desktop/go.mod`
- Create: `apps/desktop/wails.json`
- Create: `apps/desktop/Taskfile.yml`
- Create: `apps/desktop/spa.go`

- [ ] **Step 1: Install Wails v3 CLI**

```bash
go install -v github.com/wailsapp/wails/v3/cmd/wails3@latest
```

- [ ] **Step 2: Create go.mod with local replace**

```bash
mkdir -p apps/desktop
cd apps/desktop && go mod init github.com/lucas/oraculo/apps/desktop
```

Add replace directive:
```
replace github.com/lucas/oraculo => ../..
```

Then: `cd apps/desktop && go get github.com/wailsapp/wails/v3`

- [ ] **Step 3: Set up frontend asset directory**

```bash
mkdir -p apps/desktop/frontend/dist
echo '<!DOCTYPE html><html><body>Build required</body></html>' > apps/desktop/frontend/dist/index.html
echo 'frontend/dist/' >> apps/desktop/.gitignore
```

- [ ] **Step 4: Create main.go**

```go
package main

import (
	"embed"
	"log"

	"github.com/wailsapp/wails/v3/pkg/application"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	app := application.New(application.Options{
		Name:        "Oraculo",
		Description: "Oraculo Desktop — multi-project launcher",
		Services: []application.Service{
			application.NewService(NewLauncherService()),
			application.NewService(NewServerService()),
		},
		Assets: application.AssetOptions{
			Handler: NewSPAHandler(assets),
		},
		Mac: application.MacOptions{
			// No dock icon — tray-only app
			ActivationPolicy: application.ActivationPolicyAccessory,
			ApplicationShouldTerminateAfterLastWindowClosed: false,
		},
		Windows: application.WindowsOptions{
			DisableQuitOnLastWindowClosed: true,
		},
	})

	launcherWindow := app.NewWebviewWindowWithOptions(application.WebviewWindowOptions{
		Title:  "Oraculo",
		Name:   "launcher",
		Width:  1024,
		Height: 700,
		URL:    "/",
	})

	setupTray(app, launcherWindow)

	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
```

- [ ] **Step 5: Create SPA handler**

Create `apps/desktop/spa.go`:

```go
package main

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"

	"github.com/lucas/oraculo/apps/backend/src/spa"
)

func NewSPAHandler(rawAssets embed.FS) http.Handler {
	assets, _ := fs.Sub(rawAssets, "frontend/dist")
	fileServer := http.FileServer(http.FS(assets))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fsPath := strings.TrimPrefix(r.URL.Path, "/")

		// 1. Try exact file match.
		if info, err := fs.Stat(assets, fsPath); err == nil && !info.IsDir() {
			fileServer.ServeHTTP(w, r)
			return
		}

		// 2. Try placeholder substitution for dynamic routes.
		if phPath := spa.WithPlaceholders(fsPath); phPath != fsPath {
			if info, err := fs.Stat(assets, phPath); err == nil && !info.IsDir() {
				r2 := r.Clone(r.Context())
				r2.URL.Path = "/" + phPath
				fileServer.ServeHTTP(w, r2)
				return
			}
		}

		// 3. Fallback: serve the SPA shell.
		isRSC := r.URL.Query().Get("_rsc") != ""
		shell := spa.Shell(r.URL.Path, isRSC)
		r2 := r.Clone(r.Context())
		r2.URL.Path = shell
		fileServer.ServeHTTP(w, r2)
	})
}
```

Note: In Wails v3, `Assets.Handler` is the **primary** handler — it receives all requests. Unlike v2 (fallback-only), we must check for exact file matches first.

- [ ] **Step 6: Create Taskfile.yml**

Wails v3 replaces `wails.json` with `Taskfile.yml` (go-task):

```yaml
version: '3'

vars:
  APP_NAME: oraculo-desktop
  BIN_DIR: bin
  FRONTEND_DIR: ../frontend

tasks:
  build:frontend:
    dir: '{{.FRONTEND_DIR}}'
    cmds:
      - bun install
      - bun run build
      - rm -rf ../desktop/frontend/dist
      - mkdir -p ../desktop/frontend/dist
      - cp -r out/* ../desktop/frontend/dist/

  build:
    deps: [build:frontend]
    cmds:
      - go build -tags production -trimpath -ldflags="-w -s" -o {{.BIN_DIR}}/{{.APP_NAME}}

  dev:
    cmds:
      - wails3 dev

  generate:
    cmds:
      - wails3 generate bindings
```

- [ ] **Step 8: Verify it compiles**

```bash
cd apps/desktop && go build .
```
Expected: Compiles (may not run without window manager, but should compile)

- [ ] **Step 9: Commit**

```bash
git add apps/desktop/
git commit -m "feat(desktop): initialize Wails v3 app with SPA routing"
```

### Task 10: Add LauncherService and ServerService

**Files:**
- Create: `apps/desktop/launcher_service.go`
- Create: `apps/desktop/server_service.go`
- Create: `apps/desktop/projects.go`
- Create: `apps/desktop/detach_unix.go`
- Create: `apps/desktop/detach_windows.go`

- [ ] **Step 1: Create projects.go (known projects file management)**

```go
package main

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

type KnownProject struct {
	Path string `json:"path"`
	Name string `json:"name"`
}

func projectsFilePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".oraculo", "projects.json"), nil
}

func readProjects() ([]KnownProject, error) {
	path, err := projectsFilePath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	var projects []KnownProject
	return projects, json.Unmarshal(data, &projects)
}

func writeProjects(projects []KnownProject) error {
	path, err := projectsFilePath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(projects, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0o644)
}
```

- [ ] **Step 2: Create LauncherService**

Create `apps/desktop/launcher_service.go`:

```go
package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/lucas/oraculo/apps/backend/src/registry"
	"github.com/wailsapp/wails/v3/pkg/application"
)

type ProjectWithStatus struct {
	Name   string `json:"name"`
	Path   string `json:"path"`
	Online bool   `json:"online"`
	Port   int    `json:"port,omitempty"`
}

type LauncherService struct {
	binaryPath string
	wsMonitor  *wsMonitor
}

func NewLauncherService() *LauncherService {
	return &LauncherService{
		wsMonitor: newWSMonitor(),
	}
}

// ServiceStartup is called by Wails v3 during app initialization.
func (s *LauncherService) ServiceStartup(ctx context.Context, options application.ServiceOptions) error {
	s.binaryPath = findBinary()
	return nil
}

// ServiceShutdown is called by Wails v3 during app teardown.
func (s *LauncherService) ServiceShutdown() error {
	s.wsMonitor.DisconnectAll()
	return nil
}

func (s *LauncherService) ListProjects(ctx context.Context) ([]ProjectWithStatus, error) {
	projects, _ := readProjects()
	regPath, err := registry.DefaultPath()
	if err != nil {
		return nil, err
	}
	entries, _ := registry.List(regPath)

	serverMap := make(map[string]registry.Entry)
	var alive []registry.Entry
	for _, e := range entries {
		if processAlive(e.PID) {
			serverMap[e.Path] = e
			alive = append(alive, e)
		}
	}

	// Clean orphaned entries
	if len(alive) != len(entries) {
		_ = registry.WriteAll(regPath, alive)
	}

	var result []ProjectWithStatus
	for _, p := range projects {
		ps := ProjectWithStatus{Name: p.Name, Path: p.Path}
		if e, ok := serverMap[p.Path]; ok {
			ps.Online = true
			ps.Port = e.Port
		}
		result = append(result, ps)
	}
	return result, nil
}

func (s *LauncherService) AddProject(ctx context.Context) (*KnownProject, error) {
	dir, err := application.OpenDirectoryDialog(ctx, application.OpenDialogOptions{
		Title: "Select Oraculo Project",
	})
	if err != nil || dir == "" {
		return nil, err
	}

	if _, err := os.Stat(filepath.Join(dir, ".oraculo")); err != nil {
		return nil, fmt.Errorf("selected directory is not an Oraculo project (missing .oraculo/)")
	}

	name := filepath.Base(dir)
	project := KnownProject{Path: dir, Name: name}

	projects, _ := readProjects()
	for _, p := range projects {
		if p.Path == dir {
			return &project, nil
		}
	}
	projects = append(projects, project)
	if err := writeProjects(projects); err != nil {
		return nil, err
	}
	return &project, nil
}

func (s *LauncherService) RemoveProject(ctx context.Context, projectPath string) error {
	projects, err := readProjects()
	if err != nil {
		return err
	}
	for i, p := range projects {
		if p.Path == projectPath {
			projects = append(projects[:i], projects[i+1:]...)
			return writeProjects(projects)
		}
	}
	return nil
}

func (s *LauncherService) StartServer(ctx context.Context, projectPath string) error {
	cmd := exec.Command(s.binaryPath, "start", "http")
	cmd.Dir = projectPath
	detachProcess(cmd)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start server: %w", err)
	}
	cmd.Process.Release()

	// Poll health endpoint until ready (up to 5s).
	regPath, _ := registry.DefaultPath()
	for range 50 {
		time.Sleep(100 * time.Millisecond)
		entries, _ := registry.List(regPath)
		for _, e := range entries {
			if e.Path == projectPath {
				// Confirm HTTP is actually ready
				resp, err := http.Get(fmt.Sprintf("http://localhost:%d/health", e.Port))
				if err == nil {
					resp.Body.Close()
					return nil
				}
			}
		}
	}
	return fmt.Errorf("server did not become ready within 5s")
}

func (s *LauncherService) StopServer(ctx context.Context, projectPath string) error {
	cmd := exec.Command(s.binaryPath, "kill")
	cmd.Dir = projectPath
	return cmd.Run()
}

func findBinary() string {
	exe, err := os.Executable()
	if err != nil {
		return "oraculo"
	}
	bundled := filepath.Join(filepath.Dir(exe), "oraculo")
	if _, err := os.Stat(bundled); err == nil {
		return bundled
	}
	return "oraculo"
}
```

- [ ] **Step 3: Create ServerService**

Create `apps/desktop/server_service.go`:

```go
package main

import (
	"context"
	"fmt"
	"sync"
)

type ServerService struct {
	selectedServer string
	mu             sync.Mutex
}

func NewServerService() *ServerService {
	return &ServerService{}
}

func (s *ServerService) SelectServer(ctx context.Context, port int) string {
	url := fmt.Sprintf("http://localhost:%d", port)
	s.mu.Lock()
	s.selectedServer = url
	s.mu.Unlock()
	return url
}

func (s *ServerService) GetCurrentServer(ctx context.Context) string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.selectedServer
}
```

- [ ] **Step 4: Create platform-specific process detachment**

Create `apps/desktop/detach_unix.go`:
```go
//go:build !windows

package main

import (
	"os/exec"
	"syscall"
)

func detachProcess(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
}

func processAlive(pid int) bool {
	proc, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	return proc.Signal(syscall.Signal(0)) == nil
}
```

Create `apps/desktop/detach_windows.go`:
```go
//go:build windows

package main

import (
	"os"
	"os/exec"
	"syscall"
)

func detachProcess(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP}
}

func processAlive(pid int) bool {
	proc, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	// On Windows, FindProcess always succeeds. Try opening the process.
	handle, err := syscall.OpenProcess(syscall.PROCESS_QUERY_INFORMATION, false, uint32(pid))
	if err != nil {
		return false
	}
	syscall.CloseHandle(handle)
	_ = proc
	return true
}
```

- [ ] **Step 5: Verify it compiles**

```bash
cd apps/desktop && go build .
```

- [ ] **Step 6: Commit**

```bash
git add apps/desktop/
git commit -m "feat(desktop): add LauncherService and ServerService with project management"
```

### Task 11: Add system tray (built-in v3)

**Files:**
- Create: `apps/desktop/tray.go`
- Modify: `apps/desktop/main.go`

- [ ] **Step 1: Implement tray**

Create `apps/desktop/tray.go`:

```go
package main

import (
	"runtime"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
	"github.com/wailsapp/wails/v3/pkg/icons"
)

func setupTray(app *application.App, launcherWindow *application.WebviewWindow) {
	tray := app.NewSystemTray()

	// macOS: use template icon (system handles light/dark automatically)
	if runtime.GOOS == "darwin" {
		tray.SetTemplateIcon(icons.SystrayMacTemplate)
	}
	// TODO: Set custom icon from embedded resource
	// tray.SetIcon(iconBytes)
	// tray.SetDarkModeIcon(darkIconBytes)
	tray.SetTooltip("Oraculo Desktop")

	menu := app.NewMenu()
	menu.Add("Open Window").OnClick(func(ctx *application.Context) {
		launcherWindow.Show()
		launcherWindow.Focus()
	})
	menu.AddSeparator()
	menu.Add("Quit").OnClick(func(ctx *application.Context) {
		app.Quit()
	})

	tray.SetMenu(menu)

	// AttachWindow auto-toggles the window near the tray icon on click.
	// Also hides the window when it loses focus.
	tray.AttachWindow(launcherWindow).WindowOffset(5)

	// Hide window instead of quitting when close button is clicked
	launcherWindow.RegisterHook(events.Common.WindowClosing, func(e *application.WindowEvent) {
		launcherWindow.Hide()
		e.Cancel()
	})

	// macOS: re-open window on Dock icon click
	app.OnApplicationEvent(events.Mac.ApplicationShouldHandleReopen, func(e *application.ApplicationEvent) {
		launcherWindow.Show()
	})
}
```

- [ ] **Step 2: Wire tray into main.go**

After creating the launcher window, call `setupTray(app, launcherWindow)`:

```go
launcherWindow := app.NewWebviewWindowWithOptions(...)
setupTray(app, launcherWindow)
```

- [ ] **Step 3: Verify it compiles**

```bash
cd apps/desktop && go build .
```

- [ ] **Step 4: Commit**

```bash
git add apps/desktop/tray.go apps/desktop/main.go
git commit -m "feat(desktop): add system tray with show/hide toggle"
```

### Task 12: Add WS monitor and native notifications

**Files:**
- Create: `apps/desktop/ws_monitor.go`
- Create: `apps/desktop/notifications.go`

- [ ] **Step 1: Add dependencies**

```bash
cd apps/desktop && go get github.com/gen2brain/beeep
cd apps/desktop && go get github.com/coder/websocket
```

- [ ] **Step 2: Implement notifications**

Create `apps/desktop/notifications.go`:

```go
package main

import "github.com/gen2brain/beeep"

func notifyApprovalPending(project string) {
	_ = beeep.Notify(
		"Oraculo — Approval Required",
		"An agent in "+project+" is waiting for your approval.",
		"",
	)
}

func notifyStoryCompleted(project, story string) {
	_ = beeep.Notify(
		"Oraculo — Story Completed",
		"Story '"+story+"' completed in "+project+".",
		"",
	)
}

func notifyEpicCompleted(project, epic string) {
	_ = beeep.Notify(
		"Oraculo — Epic Completed",
		"Epic '"+epic+"' completed in "+project+".",
		"",
	)
}
```

- [ ] **Step 3: Implement WS monitor**

Create `apps/desktop/ws_monitor.go`:

```go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/coder/websocket"
)

type wsMonitor struct {
	mu      sync.Mutex
	cancels map[string]context.CancelFunc
}

func newWSMonitor() *wsMonitor {
	return &wsMonitor{cancels: make(map[string]context.CancelFunc)}
}

func (m *wsMonitor) Connect(project string, port int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.cancels[project]; ok {
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	m.cancels[project] = cancel

	go m.listen(ctx, project, port)
}

func (m *wsMonitor) Disconnect(project string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if cancel, ok := m.cancels[project]; ok {
		cancel()
		delete(m.cancels, project)
	}
}

func (m *wsMonitor) DisconnectAll() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, cancel := range m.cancels {
		cancel()
	}
	m.cancels = make(map[string]context.CancelFunc)
}

func (m *wsMonitor) listen(ctx context.Context, project string, port int) {
	url := fmt.Sprintf("ws://localhost:%d/ws", port)

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		conn, _, err := websocket.Dial(ctx, url, nil)
		if err != nil {
			select {
			case <-ctx.Done():
				return
			case <-time.After(2 * time.Second):
				continue
			}
		}

		m.readLoop(ctx, conn, project)
		conn.CloseNow()
	}
}

type wsEvent struct {
	Event string          `json:"event"`
	Data  json.RawMessage `json:"data,omitempty"`
}

func (m *wsMonitor) readLoop(ctx context.Context, conn *websocket.Conn, project string) {
	for {
		_, msg, err := conn.Read(ctx)
		if err != nil {
			return
		}

		var evt wsEvent
		if err := json.Unmarshal(msg, &evt); err != nil {
			continue
		}

		switch evt.Event {
		case "approval_requested":
			notifyApprovalPending(project)
		case "story_completed":
			notifyStoryCompleted(project, "")
		case "epic_completed":
			notifyEpicCompleted(project, "")
		}
	}
}
```

- [ ] **Step 4: Wire monitor into LauncherService**

Update `StartServer` to auto-connect the WS monitor after a server starts, and `StopServer` to disconnect. Add a `wsMonitor` field to `LauncherService`.

- [ ] **Step 5: Commit**

```bash
git add apps/desktop/ws_monitor.go apps/desktop/notifications.go apps/desktop/launcher_service.go
git commit -m "feat(desktop): add WS event monitor and native notifications"
```

---

## Chunk 4: Build and Distribution

### Task 13: Update Makefile

**Files:**
- Modify: `Makefile`

- [ ] **Step 1: Rewrite Makefile**

```makefile
BINARY      := oraculo
BUILD       := ./apps/backend/cmd/oraculo
PREFIX      ?= $(HOME)/.local
VERSION     := $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS     := -s -w -X main.version=$(VERSION)
FRONTEND_DIR := ./apps/frontend

.PHONY: build build-frontend build-backend build-desktop install test vet clean web-dev cross-compile

build-frontend:
	cd $(FRONTEND_DIR) && bun run build

build-backend:
	go build -ldflags "$(LDFLAGS)" -o $(BINARY) $(BUILD)

build: build-backend

build-desktop: build-backend
	cp $(BINARY) apps/desktop/bin/oraculo
	cd apps/desktop && task build

install: build
	install -m 755 $(BINARY) $(DESTDIR)$(PREFIX)/bin/$(BINARY)

test:
	go test -v -count=1 ./apps/backend/...

vet:
	go vet ./apps/backend/...

clean:
	rm -f $(BINARY)
	rm -rf $(FRONTEND_DIR)/out
	rm -rf apps/desktop/frontend/dist
	rm -rf npm/cli-*/bin/

web-dev:
	cd $(FRONTEND_DIR) && bun run dev

cross-compile:
	@mkdir -p npm/cli-darwin-arm64/bin npm/cli-darwin-x64/bin npm/cli-linux-x64/bin npm/cli-linux-arm64/bin
	GOOS=darwin  GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o npm/cli-darwin-arm64/bin/oraculo  $(BUILD)
	GOOS=darwin  GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o npm/cli-darwin-x64/bin/oraculo    $(BUILD)
	GOOS=linux   GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o npm/cli-linux-x64/bin/oraculo     $(BUILD)
	GOOS=linux   GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o npm/cli-linux-arm64/bin/oraculo   $(BUILD)
```

- [ ] **Step 2: Verify targets work**

```bash
make build-backend
make build-frontend
make test
```

- [ ] **Step 3: Commit**

```bash
git add Makefile
git commit -m "build: update Makefile for frontend/backend/desktop split"
```

### Task 14: Update CLAUDE.md

**Files:**
- Modify: `CLAUDE.md`

- [ ] **Step 1: Update all references**

- Replace `apps/dashboard` with `apps/frontend` everywhere
- Update project structure section to include `apps/desktop/`
- Update "Dashboard Static Assets" section — note frontend is embedded in desktop app, not server
- Remove `withPlaceholders`/`spaShell` server routing docs (moved to `apps/backend/src/spa/`)
- Add brief description of the desktop app and Wails v3

- [ ] **Step 2: Commit**

```bash
git add CLAUDE.md
git commit -m "docs: update CLAUDE.md for Wails v3 desktop architecture"
```
