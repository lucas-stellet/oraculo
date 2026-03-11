# Wails Desktop App Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Split Oraculo into two binaries — `oraculo` (headless CLI/server) and `oraculo-desktop` (Wails v2 native app) — connected via HTTP/WS, with server registry, system tray, and native notifications.

**Architecture:** The `oraculo` backend drops its embedded frontend and gains a global server registry (`~/.oraculo/servers.json`). The Next.js dashboard is renamed to `apps/frontend/` and gains a `ServerContext` for dynamic base URLs. A new `apps/desktop/` Wails v2 app embeds the frontend build, connects to multiple project servers, and provides native desktop features (tray, notifications).

**Tech Stack:** Go 1.24, Wails v2, Next.js (static export), `github.com/gofrs/flock`, `fyne-io/systray`, `gen2brain/beeep`

**Spec:** `docs/superpowers/specs/2026-03-11-wails-desktop-design.md`

---

## File Structure

### New files

- `apps/backend/src/registry/registry.go` — Server registry (register/unregister/list with file locking)
- `apps/backend/src/registry/registry_test.go` — Registry tests
- `apps/backend/src/server/cors.go` — CORS middleware
- `apps/backend/src/server/cors_test.go` — CORS tests
- `apps/frontend/src/lib/server-context.tsx` — React context for dynamic server URL
- `apps/desktop/main.go` — Wails entrypoint
- `apps/desktop/app.go` — App struct with bindings (ListServers, StartServer, StopServer, etc.)
- `apps/desktop/tray.go` — System tray integration
- `apps/desktop/notifications.go` — Native notification dispatch
- `apps/desktop/ws_monitor.go` — Go-side WS client for event monitoring
- `apps/desktop/projects.go` — Known projects file management
- `apps/desktop/wails.json` — Wails config
- `apps/desktop/spa.go` — SPA routing for Wails asset server (reuses withPlaceholders/spaShell)
- `apps/desktop/go.mod` — Separate Go module (with `replace github.com/lucas/oraculo => ../..` for root module access)
- `apps/desktop/build_copy.sh` — Script to copy frontend/out/ and oraculo binary into desktop bundle
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

## Chunk 1: Backend — Registry Package

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
// Creates the file and parent directories if they don't exist.
func Register(registryPath string, entry Entry) error {
	if entry.StartedAt.IsZero() {
		entry.StartedAt = time.Now()
	}
	return withLock(registryPath, func(entries []Entry) ([]Entry, error) {
		// Update existing entry for the same path, or append.
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
		return entries, nil // not found is not an error
	})
}

// List reads all entries from the registry file.
// Returns an empty slice if the file does not exist.
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

// DefaultPath returns ~/.oraculo/servers.json.
func DefaultPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".oraculo", "servers.json"), nil
}

// withLock acquires a file lock, reads entries, applies fn, and writes back.
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

- [ ] **Step 6: Write test for Unregister**

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
- Modify: `apps/backend/src/cli/start.go:63-117` (runStartAll), `apps/backend/src/cli/start.go:164-215` (runStartHTTP)
- Modify: `apps/backend/src/config/config.go` (add Name field)

- [ ] **Step 1: Add Name field to Config**

In `apps/backend/src/config/config.go`, add `Name` field to the `Config` struct:

```go
type Config struct {
	Port              int         `json:"port"`
	Name              string      `json:"name,omitempty"`
	PreferredLanguage string      `json:"preferred_language,omitempty"`
	Skills            AgentSkills `json:"skills,omitempty"`
}
```

- [ ] **Step 2: Add helper to derive project name**

Append to `apps/backend/src/config/config.go`:

```go
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

- [ ] **Step 3: Write test for ProjectName**

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

- [ ] **Step 4: Add registry register/unregister to runStartAll**

In `apps/backend/src/cli/start.go`, in `runStartAll`, after the server starts listening, register with the registry. Add defer unregister. Insert after the `g.Go` for the HTTP server (around line 98):

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

- [ ] **Step 4: Add same registry calls to runStartHTTP**

Same pattern in `runStartHTTP`, after the `g.Go` for the HTTP server (around line 198).

- [ ] **Step 5: Run existing tests to verify nothing breaks**

```bash
cd apps/backend && go test ./... -count=1
```
Expected: All PASS

- [ ] **Step 6: Commit**

```bash
git add apps/backend/src/cli/start.go apps/backend/src/config/config.go
git commit -m "feat(cli): register/unregister server in global registry on start/stop"
```

### Task 3: Add CORS middleware

**Files:**
- Create: `apps/backend/src/server/cors.go`
- Create: `apps/backend/src/server/cors_test.go`
- Modify: `apps/backend/src/server/server.go:127-130` (ServeHTTP)

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

// corsMiddleware adds permissive CORS headers for local desktop access.
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

In `apps/backend/src/server/server.go`, add a `handler` field to the Server struct and wrap the mux once in the constructor:

Add field to Server struct:
```go
type Server struct {
	mux          *http.ServeMux
	handler      http.Handler
	database     *db.DB
	lastActivity time.Time
	mu           sync.Mutex
}
```

At the end of `New()`, before the return, wrap the mux:
```go
handler := corsMiddleware(mux)
return &Server{mux: mux, handler: handler, database: database, lastActivity: time.Now()}
```

Update `ServeHTTP` to use the stored handler:
```go
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.touchActivity()
	s.handler.ServeHTTP(w, r)
}
```

This avoids creating a new handler on every request.

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
- Modify: `apps/backend/src/server/server.go:181-183` (handleHealth)

- [ ] **Step 1: Write failing test**

Add to an existing or new test file:

```go
func TestHealth_IncludesProjectName(t *testing.T) {
	srv := New(nil, nil, nil, nil, "", "test")

	req := httptest.NewRequest("GET", "/health", nil)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	var body map[string]string
	json.NewDecoder(rec.Body).Decode(&body)

	if _, ok := body["project_name"]; !ok {
		t.Error("expected project_name in health response")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
cd apps/backend && go test ./src/server/ -v -run TestHealth_IncludesProjectName
```
Expected: FAIL

- [ ] **Step 3: Modify handleHealth to include project_name**

In `server.go`, change `handleHealth` to accept a project name and return it. Update the constructor to pass the config-derived name.

Modify `New()` signature to accept `projectName string`:
```go
func New(database *db.DB, bridge *approval.Bridge, hub *ws.Hub, logs *applog.Broadcaster, projectName string, version string) *Server {
```

Replace the `handleHealth` line with a closure:
```go
mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, map[string]string{"status": "ok", "project_name": projectName})
})
```

Remove the standalone `handleHealth` function.

Update all callers of `New()` in `start.go` to pass `cfg.ProjectName()` instead of the empty `staticPath` string:
- `runStartAll` line 90: `server.New(database, bridge, hub, broadcaster, cfg.ProjectName(), version)`
- `runStartHTTP` line 191: `server.New(database, bridge, hub, broadcaster, cfg.ProjectName(), version)`

Note: existing test callers (`api_test.go`, `hooks_test.go`, `cors_test.go`) already pass `""` as the 5th arg — no code change needed since the parameter position and type are unchanged (both `string`).

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
- Modify: `apps/backend/src/server/server.go` — Remove SPA handler, withPlaceholders, spaShell, splitPath, static route

- [ ] **Step 1: Delete assets.go and assets_test.go**

```bash
rm apps/backend/src/server/assets.go apps/backend/src/server/assets_test.go
```

- [ ] **Step 2: Remove SPA handler from server.go**

In `server.go`, remove:
- The `"io/fs"` and `"strings"` imports (if no longer needed)
- Line 108: `mux.Handle("GET /", newSPAHandler(DashboardAssets))`
- Lines 103-107: The static path logging block
- The `staticPath` field from `Server` struct (line 27)
- The entire `newSPAHandler` function (lines 190-228)
- The entire `withPlaceholders` function (lines 230-253)
- The entire `spaShell` function (lines 255-283)
- The entire `splitPath` function (lines 285-294)

Note: before deleting `withPlaceholders` and `spaShell`, copy them to a new file `apps/backend/src/spa/spa.go` for reuse by the desktop app (see Task 10).

- [ ] **Step 3: Remove openBrowser and ORACULO_NO_BROWSER from start.go**

In `apps/backend/src/cli/start.go`:
- Remove the `openBrowser` function (lines 120-133)
- Remove the browser-opening goroutine in `runStartAll` (lines 100-110)
- Remove the browser-opening goroutine in `runStartHTTP` (lines 200-210)
- Remove unused imports (`"os/exec"`, `"runtime"`)

- [ ] **Step 4: Remove handleRestart from system.go**

In `apps/backend/src/server/system.go`:
- Remove the `handleRestart` method (lines 46-56)
- Remove `"syscall"` import
- Remove the route in `server.go`: `mux.HandleFunc("POST /api/system/restart", sys.handleRestart)`

- [ ] **Step 5: Remove ORACULO_NO_BROWSER from restart.go**

In `apps/backend/src/cli/restart.go`, line 55: change:
```go
Env: append(os.Environ(), "ORACULO_NO_BROWSER=1"),
```
to:
```go
Env: os.Environ(),
```

- [ ] **Step 6: Run all backend tests**

```bash
cd apps/backend && go test ./... -count=1
```
Expected: All PASS (assets_test.go is deleted, remaining tests should not depend on SPA handler)

- [ ] **Step 7: Commit**

```bash
git add -A apps/backend/src/server/ apps/backend/src/cli/start.go apps/backend/src/cli/restart.go
git commit -m "refactor(server): remove embedded frontend, SPA handler, and browser auto-open"
```

> **Note:** The frontend's restart banner in `apps/dashboard/src/app/epics/[id]/layout.tsx` (lines 75-103) will break after this change. This is addressed in Task 15 which removes `restartServer` from the frontend.

### Task 6: Extract SPA routing to shared package

**Files:**
- Create: `apps/backend/src/spa/spa.go`
- Create: `apps/backend/src/spa/spa_test.go`

- [ ] **Step 1: Write test for withPlaceholders**

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
		{"epics/gastos/stories/registro/somefile.txt", "epics/__placeholder__/stories/__placeholder__/somefile.txt"},
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

func TestSPAShell(t *testing.T) {
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
Expected: FAIL

- [ ] **Step 3: Create spa package**

Create `apps/backend/src/spa/spa.go` with the functions extracted from the deleted server.go code, renamed to exported:

```go
package spa

import "strings"

// WithPlaceholders replaces dynamic route segments with "__placeholder__".
// Known dynamic positions: epics/{id} at index 1, approvals/{approvalId} or stories/{storyId} at index 3.
func WithPlaceholders(fsPath string) string {
	segs := strings.Split(strings.Trim(fsPath, "/"), "/")
	if len(segs) < 2 || segs[0] != "epics" {
		return fsPath
	}
	result := make([]string, len(segs))
	copy(result, segs)
	changed := false
	if result[1] != "__placeholder__" {
		result[1] = "__placeholder__"
		changed = true
	}
	if len(result) >= 4 && (result[2] == "approvals" || result[2] == "stories") && result[3] != "__placeholder__" {
		result[3] = "__placeholder__"
		changed = true
	}
	if !changed {
		return fsPath
	}
	return strings.Join(result, "/")
}

// Shell maps a URL path to the most appropriate pre-rendered shell file.
func Shell(urlPath string, isRSC bool) string {
	ext := ".html"
	if isRSC {
		ext = ".txt"
	}
	segs := splitPath(urlPath)
	n := len(segs)

	if n >= 5 && segs[0] == "epics" && segs[2] == "approvals" && segs[4] == "review" {
		return "/epics/__placeholder__/approvals/__placeholder__/review" + ext
	}
	if n >= 3 && segs[0] == "epics" && segs[2] == "approvals" {
		return "/epics/__placeholder__/approvals" + ext
	}
	if n >= 4 && segs[0] == "epics" && segs[2] == "stories" {
		return "/epics/__placeholder__/stories/__placeholder__" + ext
	}
	if n >= 2 && segs[0] == "epics" {
		return "/epics/__placeholder__" + ext
	}
	return "/"
}

func splitPath(p string) []string {
	var segs []string
	for _, s := range strings.Split(strings.Trim(p, "/"), "/") {
		if s != "" {
			segs = append(segs, s)
		}
	}
	return segs
}
```

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
- Modify: `apps/frontend/next.config.ts` (if any self-references)

- [ ] **Step 1: Rename the directory**

```bash
git mv apps/dashboard apps/frontend
```

- [ ] **Step 2: Update Makefile**

In `Makefile`, replace all `DASHBOARD_DIR` references:

Replace line 6:
```makefile
DASHBOARD_DIR := ./apps/dashboard
```
With:
```makefile
FRONTEND_DIR := ./apps/frontend
```

Replace all `$(DASHBOARD_DIR)` with `$(FRONTEND_DIR)` throughout.

Remove the `dashboard_assets` copy steps from `build` and `rebuild` targets (the server no longer embeds the frontend). Simplify:

```makefile
build-frontend:
	cd $(FRONTEND_DIR) && bun run build

build-backend:
	go build -ldflags "$(LDFLAGS)" -o $(BINARY) $(BUILD)

build: build-frontend build-backend
```

Update `clean` to remove `$(FRONTEND_DIR)/out` instead of `dashboard_assets`.

Update `web-dev` to reference `$(FRONTEND_DIR)`.

- [ ] **Step 3: Update CLAUDE.md references**

Replace all occurrences of `apps/dashboard` with `apps/frontend` in `CLAUDE.md`.

- [ ] **Step 4: Verify frontend still builds**

```bash
cd apps/frontend && bun run build
```
Expected: Build succeeds

- [ ] **Step 5: Verify backend still builds**

```bash
cd apps/backend && go build ./cmd/oraculo
```
Expected: Build succeeds (no longer depends on dashboard_assets)

- [ ] **Step 6: Commit**

```bash
git add -A
git commit -m "refactor: rename apps/dashboard to apps/frontend"
```

### Task 8: Add ServerContext to frontend

**Files:**
- Create: `apps/frontend/src/lib/server-context.tsx`
- Modify: `apps/frontend/src/lib/api.ts`
- Modify: `apps/frontend/src/lib/ws.tsx`
- Modify: `apps/frontend/src/app/layout.tsx` (wrap with provider)
- Modify: `apps/frontend/src/app/page.tsx` (use api.ts instead of direct fetch)

- [ ] **Step 1: Create ServerContext provider**

Create `apps/frontend/src/lib/server-context.tsx`:

```tsx
"use client";

import { createContext, useContext, useMemo } from "react";

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
```

- [ ] **Step 2: Refactor api.ts to accept a base URL**

Replace `apps/frontend/src/lib/api.ts` — change `fetchJSON` and all endpoints to accept a base URL parameter. Use a factory pattern:

```typescript
import type {
  EpicSummary, Story, StoryTask, StoryVersion,
  Review, Validation, Approval, InlineComment,
} from "./types";

async function fetchJSON<T>(baseUrl: string, path: string): Promise<T> {
  const res = await fetch(`${baseUrl}${path}`);
  if (!res.ok) {
    throw new Error(`API ${res.status}: ${path}`);
  }
  return res.json();
}

export function createApi(baseUrl: string) {
  return {
    listEpics: () =>
      fetchJSON<EpicSummary[]>(baseUrl, "/api/epics"),

    createEpic: (name: string, description: string) =>
      fetch(`${baseUrl}/api/epics`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ name, description }),
      }).then((r) => r.json()),

    listStories: (epicName: string) =>
      fetchJSON<Story[]>(baseUrl, `/api/epics/${encodeURIComponent(epicName)}/stories`),

    listTasks: (epicName: string, storyName: string) =>
      fetchJSON<StoryTask[]>(baseUrl,
        `/api/epics/${encodeURIComponent(epicName)}/stories/${encodeURIComponent(storyName)}/tasks`
      ),

    listStoryVersions: (epicName: string, storyName: string) =>
      fetchJSON<StoryVersion[]>(baseUrl,
        `/api/epics/${encodeURIComponent(epicName)}/stories/${encodeURIComponent(storyName)}/versions`
      ),

    listStoryReviews: (epicName: string, storyName: string) =>
      fetchJSON<Review[]>(baseUrl,
        `/api/epics/${encodeURIComponent(epicName)}/stories/${encodeURIComponent(storyName)}/reviews`
      ),

    listValidations: (epicName: string, storyName: string) =>
      fetchJSON<Validation[]>(baseUrl,
        `/api/epics/${encodeURIComponent(epicName)}/stories/${encodeURIComponent(storyName)}/validations`
      ),

    listApprovals: (epicId?: number, status?: string) => {
      const params = new URLSearchParams();
      if (epicId) params.set("epic_id", String(epicId));
      if (status) params.set("status", status);
      const qs = params.toString();
      return fetchJSON<Approval[]>(baseUrl, `/api/approvals${qs ? `?${qs}` : ""}`);
    },

    getApproval: (id: string) =>
      fetchJSON<Approval>(baseUrl, `/api/approvals/${id}`),

    submitVerdict: (id: string, verdict: string, comment: string) =>
      fetch(`${baseUrl}/api/approvals/${id}/verdict`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ verdict, comment }),
      }).then((r) => r.json()),

    createComment: (approvalId: string, selectedText: string, comment: string) =>
      fetch(`${baseUrl}/api/approvals/${approvalId}/comments`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ selected_text: selectedText, comment }),
      }).then((r) => r.json()),

    listComments: (approvalId: string) =>
      fetchJSON<InlineComment[]>(baseUrl, `/api/approvals/${approvalId}/comments`),

    deleteComment: (approvalId: string, commentId: number) =>
      fetch(`${baseUrl}/api/approvals/${approvalId}/comments/${commentId}`, {
        method: "DELETE",
      }),

    getSystemStatus: () =>
      fetchJSON<{ update_available: boolean; started_at: string; version: string; project_commit: string; new_version: string }>(baseUrl, "/api/system/status"),
  };
}

// Default api instance for backward compatibility (same-origin).
export const api = createApi("");
```

- [ ] **Step 3: Create useApi hook**

Add to `apps/frontend/src/lib/server-context.tsx`:

```tsx
import { createApi } from "./api";

export function useApi() {
  const baseUrl = useServerUrl();
  return useMemo(() => createApi(baseUrl), [baseUrl]);
}
```

- [ ] **Step 4: Refactor ws.tsx to use ServerContext**

In `apps/frontend/src/lib/ws.tsx`, modify the `connect` callback (lines 36-39) to accept baseUrl. Change the `WebSocketProvider` to accept an optional `serverUrl` prop:

Replace lines 36-41:
```typescript
const connect = useCallback(() => {
  const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
  const url = `${protocol}//${window.location.host}/ws`;
```

With:
```typescript
const connect = useCallback(() => {
  let wsBase: string;
  if (serverUrl) {
    wsBase = serverUrl.replace(/^http/, "ws");
  } else {
    const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
    wsBase = `${protocol}//${window.location.host}`;
  }
  const url = `${wsBase}/ws`;
```

Add `serverUrl` prop to `WebSocketProvider`:
```typescript
export function WebSocketProvider({ serverUrl, children }: { serverUrl?: string; children: React.ReactNode }) {
```

Add `serverUrl` to the `useCallback` dependency array.

- [ ] **Step 5: Refactor page.tsx to use api.ts**

In `apps/frontend/src/app/page.tsx`, replace direct fetch calls with the `useApi` hook:

```tsx
import { useApi } from "@/lib/server-context";

// Inside component:
const api = useApi();

useEffect(() => {
  api.listEpics()
    .then((data) => setEpics(data ?? []))
    .catch(() => setEpics([]));
}, [api]);

async function handleCreate(name: string, description: string) {
  await api.createEpic(name, description);
  const updated = await api.listEpics();
  setEpics(updated ?? []);
}
```

- [ ] **Step 6: Wire ServerProvider into root layout**

In `apps/frontend/src/app/layout.tsx`, wrap the children with `ServerProvider`. Do NOT add `WebSocketProvider` here — it already lives in `apps/frontend/src/app/epics/[id]/layout.tsx`.

```tsx
import { ServerProvider } from "@/lib/server-context";

// In the layout body, wrap the existing content:
<ThemeProvider>
  <ServerProvider>
    <main className="min-h-screen bg-background">
      {children}
    </main>
  </ServerProvider>
</ThemeProvider>
```

Then in `apps/frontend/src/app/epics/[id]/layout.tsx`, pass serverUrl to the existing `WebSocketProvider`:

```tsx
import { useServerUrl } from "@/lib/server-context";

// Inside the component:
const serverUrl = useServerUrl();

// In JSX:
<WebSocketProvider serverUrl={serverUrl}>
```

- [ ] **Step 7: Update all components using `api` import to use `useApi` hook**

These 6 files import `{ api }` from `@/lib/api` and need updating:

1. `apps/frontend/src/app/epics/[id]/_client.tsx`
2. `apps/frontend/src/app/epics/[id]/layout.tsx`
3. `apps/frontend/src/app/epics/[id]/approvals/_client.tsx`
4. `apps/frontend/src/app/epics/[id]/stories/[storyId]/_client.tsx`
5. `apps/frontend/src/app/epics/[id]/approvals/[approvalId]/review/_client.tsx`
6. `apps/frontend/src/app/epics/[id]/stories/[storyId]/_components/design-tab.tsx`

For each file:
- Replace `import { api } from "@/lib/api"` with `import { useApi } from "@/lib/server-context"`
- Add `const api = useApi();` inside the component function body
- All existing `api.xxx()` calls remain unchanged (the hook returns the same interface)

- [ ] **Step 8: Verify frontend builds**

```bash
cd apps/frontend && bun run build
```
Expected: Build succeeds

- [ ] **Step 9: Verify dev mode works**

Start the backend: `cd apps/backend && go run ./cmd/oraculo start http`
Start the frontend: `cd apps/frontend && bun dev`
Open `http://localhost:3000` — should work as before.

- [ ] **Step 10: Commit**

```bash
git add apps/frontend/src/
git commit -m "feat(frontend): add ServerContext for dynamic server URLs"
```

---

## Chunk 3: Desktop App — Wails Setup and Bindings

### Task 9: Initialize Wails project

**Files:**
- Create: `apps/desktop/main.go`
- Create: `apps/desktop/app.go`
- Create: `apps/desktop/wails.json`
- Create: `apps/desktop/go.mod`

- [ ] **Step 1: Install Wails CLI**

```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

- [ ] **Step 2: Initialize Wails project**

```bash
cd apps && wails init -n desktop -t vanilla
```

This creates the skeleton. Then clean up the generated files and customize.

- [ ] **Step 3: Set up go.mod with local replace**

The desktop is a separate Go module that imports the root module's packages (`registry`, `spa`):

```bash
cd apps/desktop && go mod init github.com/lucas/oraculo/apps/desktop
```

Add a replace directive for the root module (the root `go.mod` is at `github.com/lucas/oraculo`, two directories up):
```
replace github.com/lucas/oraculo => ../..
```

Then: `cd apps/desktop && go get github.com/wailsapp/wails/v2`

- [ ] **Step 4: Set up frontend asset copy**

Go's `//go:embed` cannot reference paths outside the module directory (`..` is not allowed). The build step must copy the frontend output into the desktop directory.

Add to `wails.json` a build command that copies the frontend:
```json
"frontend:build": "cd ../frontend && bun run build && rm -rf ../desktop/frontend/dist && mkdir -p ../desktop/frontend/dist && cp -r out/* ../desktop/frontend/dist/"
```

Add `frontend/dist/` to `.gitignore` in `apps/desktop/`.

Create a placeholder so the embed directive works even without a build:
```bash
mkdir -p apps/desktop/frontend/dist
echo '<!DOCTYPE html><html><body>Build required</body></html>' > apps/desktop/frontend/dist/index.html
```

- [ ] **Step 5: Create main.go**

```go
package main

import (
	"embed"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	app := NewApp()

	err := wails.Run(&options.App{
		Title:  "Oraculo",
		Width:  1280,
		Height: 800,
		AssetServer: &assetserver.Options{
			Assets:  assets,
			Handler: NewSPAHandler(assets),
		},
		OnStartup:     app.startup,
		OnBeforeClose: app.beforeClose,
		OnShutdown:    app.shutdown,
		Bind: []interface{}{
			app,
		},
	})
	if err != nil {
		panic(err)
	}
}
```

Note: `AssetServer.Handler` is a **fallback** — Wails first checks the embedded FS for an exact match, and only calls the handler if the file is not found. This is exactly what we need: exact files are served directly, and the handler provides SPA routing for missing paths.

- [ ] **Step 6: Create app.go with basic bindings**

```go
package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"github.com/lucas/oraculo/apps/backend/src/registry"
	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx            context.Context
	binaryPath     string
	wsMonitor      *wsMonitor
	selectedServer string
	mu             sync.Mutex
}

func NewApp() *App {
	return &App{
		wsMonitor: newWSMonitor(),
	}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.binaryPath = a.findBinary()
	a.setupTray()
}

func (a *App) beforeClose(ctx context.Context) bool {
	// Hide instead of quit — tray keeps running
	wailsRuntime.Hide(ctx)
	return true // prevent default close
}

func (a *App) shutdown(ctx context.Context) {
	a.wsMonitor.DisconnectAll()
}

// findBinary locates the bundled oraculo binary.
func (a *App) findBinary() string {
	exe, err := os.Executable()
	if err != nil {
		return "oraculo" // fallback to PATH
	}
	// macOS .app bundle: binary is in Contents/MacOS/
	bundled := filepath.Join(filepath.Dir(exe), "oraculo")
	if _, err := os.Stat(bundled); err == nil {
		return bundled
	}
	return "oraculo"
}

type ServerInfo struct {
	Project   string    `json:"project"`
	Path      string    `json:"path"`
	Port      int       `json:"port"`
	PID       int       `json:"pid"`
	StartedAt time.Time `json:"started_at"`
	Online    bool      `json:"online"`
}

// ListServers reads the registry, validates PIDs, cleans orphans.
func (a *App) ListServers() ([]ServerInfo, error) {
	regPath, err := registry.DefaultPath()
	if err != nil {
		return nil, err
	}
	entries, err := registry.List(regPath)
	if err != nil {
		return nil, err
	}

	var servers []ServerInfo
	var alive []registry.Entry
	for _, e := range entries {
		online := processAlive(e.PID)
		if online {
			alive = append(alive, e)
		}
		servers = append(servers, ServerInfo{
			Project:   e.Project,
			Path:      e.Path,
			Port:      e.Port,
			PID:       e.PID,
			StartedAt: e.StartedAt,
			Online:    online,
		})
	}

	// Clean orphaned entries atomically
	if len(alive) != len(entries) {
		_ = registry.WriteAll(regPath, alive)
	}

	return servers, nil
}

// StartServer spawns oraculo start http in the given project directory.
func (a *App) StartServer(projectPath string) error {
	cmd := exec.Command(a.binaryPath, "start", "http")
	cmd.Dir = projectPath
	// Detach the child process so it survives desktop app exit.
	detachProcess(cmd)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start server: %w", err)
	}
	cmd.Process.Release()

	// Poll registry for the new entry (up to 5s).
	regPath, _ := registry.DefaultPath()
	for range 50 {
		time.Sleep(100 * time.Millisecond)
		entries, _ := registry.List(regPath)
		for _, e := range entries {
			if e.Path == projectPath {
				return nil
			}
		}
	}
	return fmt.Errorf("server did not register within 5s")
}

// StopServer kills the server for the given project.
func (a *App) StopServer(projectPath string) error {
	cmd := exec.Command(a.binaryPath, "kill")
	cmd.Dir = projectPath
	return cmd.Run()
}

// SelectServer sets the active project and returns its base URL.
func (a *App) SelectServer(port int) string {
	url := fmt.Sprintf("http://localhost:%d", port)
	a.mu.Lock()
	a.selectedServer = url
	a.mu.Unlock()
	return url
}

// GetCurrentServer returns the base URL for the currently selected project.
func (a *App) GetCurrentServer() string {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.selectedServer
}

func processAlive(pid int) bool {
	proc, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	return proc.Signal(syscall.Signal(0)) == nil
}
```

Create `apps/desktop/detach_unix.go` (build-tagged for Unix):
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
```

Create `apps/desktop/detach_windows.go`:
```go
//go:build windows

package main

import (
	"os/exec"
	"syscall"
)

func detachProcess(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP}
}
```

Also add `WriteAll` to the registry package (`apps/backend/src/registry/registry.go`):
```go
// WriteAll atomically replaces all entries in the registry file.
func WriteAll(registryPath string, entries []Entry) error {
	return withLock(registryPath, func(_ []Entry) ([]Entry, error) {
		return entries, nil
	})
}
```

- [ ] **Step 7: Create SPA handler for Wails asset server**

Create `apps/desktop/spa.go`.

**Important:** Wails' `AssetServer.Handler` is a **fallback** — it is only called when the embedded FS does NOT contain the requested file. So we never need to check for exact file matches; Wails already did that. We only handle:
1. Placeholder substitution (dynamic routes stored under `__placeholder__`)
2. SPA shell fallback (HTML/TXT for unknown paths)

```go
package main

import (
	"embed"
	"io/fs"
	"net/http"

	"github.com/lucas/oraculo/apps/backend/src/spa"
)

// NewSPAHandler creates an http.Handler for Wails' AssetServer fallback.
// Only called when the exact file is NOT found in the embedded FS.
func NewSPAHandler(rawAssets embed.FS) http.Handler {
	assets, _ := fs.Sub(rawAssets, "frontend/dist")
	fileServer := http.FileServer(http.FS(assets))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fsPath := r.URL.Path
		if len(fsPath) > 0 && fsPath[0] == '/' {
			fsPath = fsPath[1:]
		}

		// Try placeholder substitution for dynamic routes.
		// e.g., /epics/gastos/approvals -> /epics/__placeholder__/approvals
		if phPath := spa.WithPlaceholders(fsPath); phPath != fsPath {
			if info, err := fs.Stat(assets, phPath); err == nil && !info.IsDir() {
				r2 := r.Clone(r.Context())
				r2.URL.Path = "/" + phPath
				fileServer.ServeHTTP(w, r2)
				return
			}
		}

		// Fallback: serve the best-matching SPA shell.
		isRSC := r.URL.Query().Get("_rsc") != ""
		shell := spa.Shell(r.URL.Path, isRSC)
		r2 := r.Clone(r.Context())
		r2.URL.Path = shell
		fileServer.ServeHTTP(w, r2)
	})
}
```

- [ ] **Step 8: Configure wails.json**

Create `apps/desktop/wails.json`:
```json
{
  "name": "oraculo-desktop",
  "outputfilename": "oraculo-desktop",
  "frontend:dir": "../frontend",
  "frontend:install": "bun install",
  "frontend:build": "cd ../frontend && bun run build && rm -rf ../desktop/frontend/dist && mkdir -p ../desktop/frontend/dist && cp -r out/* ../desktop/frontend/dist/",
  "frontend:dev:watcher": "bun dev",
  "frontend:dev:serverUrl": "auto",
  "author": {
    "name": "Lucas Stellet"
  }
}
```

The `frontend:build` command builds Next.js then copies the output into `apps/desktop/frontend/dist/` where the `//go:embed` directive can find it.

- [ ] **Step 9: Verify Wails builds**

```bash
cd apps/desktop && wails build
```
Expected: Produces `build/bin/oraculo-desktop`

- [ ] **Step 10: Commit**

```bash
git add apps/desktop/
git commit -m "feat(desktop): initialize Wails app with launcher bindings and SPA routing"
```

### Task 10: Add project management bindings

**Files:**
- Create: `apps/desktop/projects.go`

- [ ] **Step 1: Implement projects file management**

Create `apps/desktop/projects.go`:

```go
package main

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
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

// AddProject opens a directory picker and adds the selected project.
func (a *App) AddProject() (*KnownProject, error) {
	dir, err := wailsRuntime.OpenDirectoryDialog(a.ctx, wailsRuntime.OpenDialogOptions{
		Title: "Select Oraculo Project",
	})
	if err != nil || dir == "" {
		return nil, err
	}

	// Verify it has .oraculo/ directory
	if _, err := os.Stat(filepath.Join(dir, ".oraculo")); err != nil {
		return nil, errors.New("selected directory is not an Oraculo project (missing .oraculo/)")
	}

	name := filepath.Base(dir)
	project := KnownProject{Path: dir, Name: name}

	projects, _ := readProjects()
	// Avoid duplicates
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

// RemoveProject removes a project from the known projects list.
func (a *App) RemoveProject(projectPath string) error {
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

// ListProjects returns all known projects with their server status.
func (a *App) ListProjects() ([]ProjectWithStatus, error) {
	projects, _ := readProjects()
	servers, _ := a.ListServers()

	serverMap := make(map[string]ServerInfo)
	for _, s := range servers {
		serverMap[s.Path] = s
	}

	var result []ProjectWithStatus
	for _, p := range projects {
		status := ProjectWithStatus{
			Name: p.Name,
			Path: p.Path,
		}
		if s, ok := serverMap[p.Path]; ok && s.Online {
			status.Online = true
			status.Port = s.Port
		}
		result = append(result, status)
	}
	return result, nil
}

type ProjectWithStatus struct {
	Name   string `json:"name"`
	Path   string `json:"path"`
	Online bool   `json:"online"`
	Port   int    `json:"port,omitempty"`
}
```

- [ ] **Step 2: Commit**

```bash
git add apps/desktop/projects.go
git commit -m "feat(desktop): add project management bindings"
```

### Task 11: Add system tray

**Files:**
- Create: `apps/desktop/tray.go`

- [ ] **Step 1: Add systray dependency**

```bash
cd apps/desktop && go get fyne.io/systray
```

- [ ] **Step 2: Implement tray**

**Important macOS caveat:** `systray.Run()` calls into Cocoa APIs that must run on the main thread on macOS. Running it in a goroutine will crash. The `fyne-io/systray` fork provides `systray.Register(onReady, onExit)` which registers callbacks without taking over the main thread — use this instead of `systray.Run()`.

Create `apps/desktop/tray.go`:

```go
package main

import (
	"fyne.io/systray"
	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

func (a *App) setupTray() {
	// Use Register (not Run) to avoid conflicting with Wails' event loop on macOS.
	systray.Register(func() {
		systray.SetTitle("Oraculo")
		systray.SetTooltip("Oraculo Desktop")
		// TODO: Set icon from embedded resource
		// systray.SetIcon(iconBytes)

		mShow := systray.AddMenuItem("Open Window", "Show Oraculo window")
		systray.AddSeparator()
		mQuit := systray.AddMenuItem("Quit", "Quit Oraculo Desktop")

		go func() {
			for {
				select {
				case <-mShow.ClickedCh:
					wailsRuntime.Show(a.ctx)
				case <-mQuit.ClickedCh:
					systray.Quit()
					wailsRuntime.Quit(a.ctx)
					return
				}
			}
		}()
	}, func() {
		// Cleanup
	})
}
```

Note: if `fyne-io/systray.Register` is not available in the version used, fall back to `go systray.Run(...)` and test on macOS. If it crashes, the systray will need to be integrated differently (e.g., via Wails' own tray support when available, or via a helper process).

- [ ] **Step 3: Commit**

```bash
git add apps/desktop/tray.go apps/desktop/app.go
git commit -m "feat(desktop): add system tray with show/quit menu"
```

### Task 12: Add WS monitor and native notifications

**Files:**
- Create: `apps/desktop/ws_monitor.go`
- Create: `apps/desktop/notifications.go`

- [ ] **Step 1: Add beeep and websocket dependencies**

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
	cancels map[string]context.CancelFunc // projectPath -> cancel
}

func newWSMonitor() *wsMonitor {
	return &wsMonitor{cancels: make(map[string]context.CancelFunc)}
}

// Connect starts monitoring a server's WebSocket for events.
func (m *wsMonitor) Connect(project string, port int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Already connected
	if _, ok := m.cancels[project]; ok {
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	m.cancels[project] = cancel

	go m.listen(ctx, project, port)
}

// Disconnect stops monitoring a server.
func (m *wsMonitor) Disconnect(project string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if cancel, ok := m.cancels[project]; ok {
		cancel()
		delete(m.cancels, project)
	}
}

// DisconnectAll stops all monitors.
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

- [ ] **Step 4: Wire monitor into App**

In `app.go`, add the monitor field and wire connect/disconnect:

```go
type App struct {
	ctx        context.Context
	binaryPath string
	wsMonitor  *wsMonitor
}

func NewApp() *App {
	return &App{
		wsMonitor: newWSMonitor(),
	}
}

func (a *App) shutdown(ctx context.Context) {
	a.wsMonitor.DisconnectAll()
}
```

Update `StartServer` to auto-connect the monitor after a server starts, and `StopServer` to disconnect.

- [ ] **Step 5: Commit**

```bash
git add apps/desktop/ws_monitor.go apps/desktop/notifications.go apps/desktop/app.go
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

# Bundle the oraculo binary alongside the desktop app.
# wails.json frontend:build handles copying frontend/out/ -> desktop/frontend/dist/.
build-desktop: build-backend
	cp $(BINARY) apps/desktop/build/bin/oraculo
	cd apps/desktop && wails build

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

Note: `build-desktop` depends on `build-backend` (not `build-frontend`) because `wails build` triggers `frontend:build` from `wails.json`, which handles the frontend build + copy in one step. This avoids double-building.

- [ ] **Step 2: Verify targets work**

```bash
make build-backend
make build-frontend
make test
```
Expected: All succeed

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
- Update the project structure section to include `apps/desktop/`
- Update the "Dashboard Static Assets" section — replace with a note that the frontend is embedded in the desktop app, not the server
- Remove the `withPlaceholders`/`spaShell` server routing docs (moved to desktop)
- Add brief description of the desktop app

- [ ] **Step 2: Commit**

```bash
git add CLAUDE.md
git commit -m "docs: update CLAUDE.md for Wails desktop architecture"
```

### Task 15: Remove restart feature from frontend

**Files:**
- Modify: `apps/frontend/src/lib/api.ts` — remove `restartServer` method
- Modify: `apps/frontend/src/app/epics/[id]/layout.tsx` — remove restart banner and related state

- [ ] **Step 1: Remove restartServer from api.ts**

In `createApi` factory, remove the `restartServer` entry. Also remove it from the default `api` export if still present.

- [ ] **Step 2: Remove restart UI from epic layout**

In `apps/frontend/src/app/epics/[id]/layout.tsx`:
- Remove `handleRestart` function (lines 75-80)
- Remove `restarting` state (line 35)
- Remove `updateAvailable` and `newVersion` state (lines 33-34)
- Remove the `getSystemStatus` polling `useEffect` (lines 46-59)
- Remove the entire update banner JSX block (lines 84-103)
- Remove the `RefreshCw` import from lucide-react
- Keep `version` state and the sidebar — the version display still works via `getSystemStatus` if needed, or remove it entirely since the desktop manages system info

- [ ] **Step 3: Verify frontend builds**

```bash
cd apps/frontend && bun run build
```
Expected: Build succeeds

- [ ] **Step 4: Commit**

```bash
git add apps/frontend/
git commit -m "refactor(frontend): remove restart feature (desktop manages server lifecycle)"
```
