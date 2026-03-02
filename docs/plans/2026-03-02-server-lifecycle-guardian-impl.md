# Server Lifecycle Guardian Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Auto-start the HTTP server from SessionStart hook, add idle timeout auto-shutdown, and split `oraculo start` into `start mcp` / `start http` subcommands.

**Architecture:** SessionStart hook becomes the guardian — detects offline server, spawns `oraculo start http` as detached daemon, polls until healthy. HTTP server self-terminates after 15 minutes of inactivity via watchdog goroutine.

**Tech Stack:** Go, cobra (CLI), net/http, os/exec, syscall (process detachment), errgroup

**Design doc:** `docs/plans/2026-03-02-server-lifecycle-guardian-design.md`

---

### Task 1: Add idle timeout to HTTP server

The HTTP server needs to track last activity and auto-shutdown when idle.

**Files:**
- Modify: `src/server/server.go:15-19` (Server struct), `src/server/server.go:69-72` (ServeHTTP), `src/server/server.go:76-95` (ListenAndServe)
- Test: `src/server/server_test.go` (new)

**Step 1: Write the failing test for activity tracking**

Create `src/server/server_test.go`:

```go
package server_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/lucas/oraculo/src/approval"
	"github.com/lucas/oraculo/src/db"
	"github.com/lucas/oraculo/src/dbtest"
	"github.com/lucas/oraculo/src/server"
	"github.com/lucas/oraculo/src/ws"
)

func testServer(t *testing.T) *server.Server {
	database := dbtest.Open(t)
	hub := ws.NewHub()
	bridge := approval.NewBridge(db.NewApprovalStore(database), hub)
	return server.New(database, bridge, hub, nil)
}

func TestLastActivity_UpdatedOnRequest(t *testing.T) {
	srv := testServer(t)

	before := srv.LastActivity()

	time.Sleep(10 * time.Millisecond)
	req := httptest.NewRequest("GET", "/health", nil)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	after := srv.LastActivity()
	if !after.After(before) {
		t.Errorf("LastActivity not updated: before=%v after=%v", before, after)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./src/server/ -run TestLastActivity -v`
Expected: FAIL — `srv.LastActivity undefined`

**Step 3: Implement activity tracking**

In `src/server/server.go`, add `sync/atomic` time tracking:

```go
import (
	"sync"
	"time"
	// ... existing imports
)

type Server struct {
	mux          *http.ServeMux
	database     *db.DB
	lastActivity time.Time
	mu           sync.Mutex
}
```

Add `LastActivity()` method:

```go
func (s *Server) LastActivity() time.Time {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.lastActivity
}

func (s *Server) touchActivity() {
	s.mu.Lock()
	s.lastActivity = time.Now()
	s.mu.Unlock()
}
```

Update `ServeHTTP` to track activity:

```go
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.touchActivity()
	s.mux.ServeHTTP(w, r)
}
```

Initialize `lastActivity` in `New()`:

```go
return &Server{mux: mux, database: database, lastActivity: time.Now()}
```

**Step 4: Run test to verify it passes**

Run: `go test ./src/server/ -run TestLastActivity -v`
Expected: PASS

**Step 5: Write the failing test for idle shutdown**

Add to `src/server/server_test.go`:

```go
func TestListenAndServe_ShutdownOnIdle(t *testing.T) {
	srv := testServer(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	idleTimeout := 200 * time.Millisecond
	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.ListenAndServe(ctx, 0, idleTimeout)
	}()

	// Give server time to start then let it idle
	time.Sleep(50 * time.Millisecond)

	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// Server shut down due to idle — success
	case <-time.After(2 * time.Second):
		cancel()
		t.Fatal("server did not shut down after idle timeout")
	}
}
```

**Step 6: Run test to verify it fails**

Run: `go test ./src/server/ -run TestListenAndServe_ShutdownOnIdle -v`
Expected: FAIL — `too many arguments in call to ListenAndServe`

**Step 7: Implement idle timeout in ListenAndServe**

Update signature and add watchdog:

```go
func (s *Server) ListenAndServe(ctx context.Context, port int, idleTimeout time.Duration) error {
	addr := fmt.Sprintf(":%d", port)
	httpSrv := &http.Server{
		Addr:    addr,
		Handler: s,
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Idle timeout watchdog
	if idleTimeout > 0 {
		go func() {
			ticker := time.NewTicker(idleTimeout / 4)
			defer ticker.Stop()
			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					if time.Since(s.LastActivity()) > idleTimeout {
						cancel()
						return
					}
				}
			}
		}()
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCancel()
		httpSrv.Shutdown(shutdownCtx)
	}()

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	if err := httpSrv.Serve(ln); err != http.ErrServerClosed {
		return err
	}
	return nil
}
```

**Step 8: Run tests to verify they pass**

Run: `go test ./src/server/ -v`
Expected: PASS

**Step 9: Fix callers of ListenAndServe**

In `src/cli/start.go:63`, update the call to pass `0` for no idle timeout:

```go
return srv.ListenAndServe(ctx, port, 0)
```

**Step 10: Run all tests**

Run: `go test ./...`
Expected: PASS

**Step 11: Commit**

```bash
git add src/server/server.go src/server/server_test.go src/cli/start.go
git commit -m "feat: add idle timeout watchdog to HTTP server"
```

---

### Task 2: Split `oraculo start` into subcommands

**Files:**
- Modify: `src/cli/start.go` (split into parent + mcp + http subcommands)
- Modify: `src/cli/root.go:22` (register subcommands)
- Test: `src/cli/start_test.go` (new)

**Step 1: Write the failing test**

Create `src/cli/start_test.go`:

```go
package cli_test

import (
	"bytes"
	"testing"

	"github.com/lucas/oraculo/src/cli"
)

func TestStartCmd_HasSubcommands(t *testing.T) {
	root := cli.NewRoot("test")

	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetErr(&buf)
	root.SetArgs([]string{"start", "--help"})
	root.Execute()

	out := buf.String()
	for _, sub := range []string{"mcp", "http"} {
		if !bytes.Contains([]byte(out), []byte(sub)) {
			t.Errorf("start --help missing subcommand %q", sub)
		}
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./src/cli/ -run TestStartCmd_HasSubcommands -v`
Expected: FAIL — help output won't contain "mcp" or "http"

**Step 3: Implement the split**

Rewrite `src/cli/start.go`:

```go
package cli

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"

	"github.com/lucas/oraculo/src/applog"
	"github.com/lucas/oraculo/src/approval"
	"github.com/lucas/oraculo/src/config"
	"github.com/lucas/oraculo/src/db"
	mcpserver "github.com/lucas/oraculo/src/mcp"
	"github.com/lucas/oraculo/src/server"
	"github.com/lucas/oraculo/src/ws"
)

const defaultIdleTimeout = 15 * time.Minute

func newStartCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start Oraculo services",
		Long:  "Start Oraculo services. Without subcommand, starts both MCP and HTTP servers.",
		RunE:  runStartAll,
	}
	cmd.AddCommand(newStartMCPCmd())
	cmd.AddCommand(newStartHTTPCmd())
	return cmd
}

func newStartMCPCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "mcp",
		Short: "Start MCP server on stdio (managed by Claude Code)",
		RunE:  runStartMCP,
	}
}

func newStartHTTPCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "http",
		Short: "Start HTTP + WebSocket server as daemon",
		RunE:  runStartHTTP,
	}
}

// runStartAll starts both MCP and HTTP servers (backwards compatible).
func runStartAll(cmd *cobra.Command, _ []string) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	database, err := db.Open()
	if err != nil {
		return err
	}
	defer database.Close()

	cfg, err := config.Read()
	if err != nil {
		return err
	}

	port := cfg.Port
	if port == 0 {
		port = 3100
	}

	broadcaster := applog.NewBroadcaster(os.Stderr)
	logger := slog.New(broadcaster)

	hub := ws.NewHub()
	bridge := approval.NewBridge(db.NewApprovalStore(database), hub)
	srv := server.New(database, bridge, hub, broadcaster)
	mcpSrv := mcpserver.New(bridge, db.NewApprovalStore(database), logger)

	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error { return hub.Run(ctx) })
	g.Go(func() error {
		logger.Info("server.started", "port", port)
		return srv.ListenAndServe(ctx, port, 0)
	})
	g.Go(func() error { return mcpSrv.RunStdio(ctx) })

	err = g.Wait()
	logger.Info("server.stopping")
	return err
}

// runStartMCP starts only the MCP server on stdio.
func runStartMCP(cmd *cobra.Command, _ []string) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	database, err := db.Open()
	if err != nil {
		return err
	}
	defer database.Close()

	broadcaster := applog.NewBroadcaster(os.Stderr)
	logger := slog.New(broadcaster)

	hub := ws.NewHub()
	bridge := approval.NewBridge(db.NewApprovalStore(database), hub)
	mcpSrv := mcpserver.New(bridge, db.NewApprovalStore(database), logger)

	return mcpSrv.RunStdio(ctx)
}

// runStartHTTP starts the HTTP + WebSocket server as a daemon with idle timeout.
func runStartHTTP(cmd *cobra.Command, _ []string) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	database, err := db.Open()
	if err != nil {
		return err
	}
	defer database.Close()

	cfg, err := config.Read()
	if err != nil {
		return err
	}

	port := cfg.Port
	if port == 0 {
		port = 3100
	}

	broadcaster := applog.NewBroadcaster(os.Stderr)
	logger := slog.New(broadcaster)

	hub := ws.NewHub()
	bridge := approval.NewBridge(db.NewApprovalStore(database), hub)
	srv := server.New(database, bridge, hub, broadcaster)

	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error { return hub.Run(ctx) })
	g.Go(func() error {
		logger.Info("server.started", "port", port, "idle_timeout", defaultIdleTimeout)
		return srv.ListenAndServe(ctx, port, defaultIdleTimeout)
	})

	err = g.Wait()
	logger.Info("server.stopping")
	return err
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./src/cli/ -run TestStartCmd_HasSubcommands -v`
Expected: PASS

**Step 5: Run all tests**

Run: `go test ./...`
Expected: PASS

**Step 6: Commit**

```bash
git add src/cli/start.go src/cli/start_test.go
git commit -m "feat: split oraculo start into start mcp and start http subcommands"
```

---

### Task 3: Add auto-start logic to SessionStart hook

**Files:**
- Modify: `src/cli/hook_session.go:33-77` (add auto-start with polling)
- Test: `src/cli/hook_session_test.go` (extend)

**Step 1: Write the failing test for auto-start detection**

Add to `src/cli/hook_session_test.go`:

```go
func TestHookSessionStart_AlertsWhenServerOffline(t *testing.T) {
	tmp := setupInstallDir(t)

	// Write config with port that nothing is listening on
	os.MkdirAll(filepath.Join(tmp, ".oraculo"), 0o755)
	os.WriteFile(
		filepath.Join(tmp, ".oraculo", "config.json"),
		[]byte(`{"port": 39999}`),
		0o644,
	)

	var buf bytes.Buffer
	cmd := cli.NewRoot("test")
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"hook", "session-start"})
	err := cmd.Execute()

	if err != nil {
		t.Fatalf("hook should never return error: %v", err)
	}
	stderr := buf.String()
	if !strings.Contains(stderr, "warning") {
		t.Errorf("expected warning in output, got: %s", stderr)
	}
}
```

**Step 2: Run test to verify it passes (baseline)**

Run: `go test ./src/cli/ -run TestHookSessionStart_AlertsWhenServerOffline -v`
Expected: PASS (this validates current behavior)

**Step 3: Write the failing test for auto-start attempt**

Add to `src/cli/hook_session_test.go`:

```go
func TestHookSessionStart_AttemptsAutoStart(t *testing.T) {
	tmp := setupInstallDir(t)

	os.MkdirAll(filepath.Join(tmp, ".oraculo"), 0o755)
	os.WriteFile(
		filepath.Join(tmp, ".oraculo", "config.json"),
		[]byte(`{"port": 39999}`),
		0o644,
	)

	var buf bytes.Buffer
	cmd := cli.NewRoot("test")
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"hook", "session-start"})
	cmd.Execute()

	output := buf.String()
	// Should attempt auto-start and report failure (oraculo binary may not exist in test)
	if !strings.Contains(output, "auto-start") && !strings.Contains(output, "starting") {
		t.Errorf("expected auto-start attempt in output, got: %s", output)
	}
}
```

**Step 4: Run test to verify it fails**

Run: `go test ./src/cli/ -run TestHookSessionStart_AttemptsAutoStart -v`
Expected: FAIL — output won't contain "auto-start"

**Step 5: Implement auto-start logic**

Rewrite `hookSessionStart()` in `src/cli/hook_session.go`:

```go
func hookSessionStart(cmd *cobra.Command) error {
	database, err := db.Open()
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer database.Close()

	// Collect metadata
	id := uuid.New().String()
	wd, _ := os.Getwd()
	branch := gitBranch()
	metadata := map[string]string{
		"session_id":  id,
		"working_dir": wd,
		"git_branch":  branch,
		"started_at":  time.Now().UTC().Format(time.RFC3339),
	}
	metadataJSON, _ := json.Marshal(metadata)

	// Register in SQLite
	_, err = database.Conn().Exec(
		"INSERT INTO claude_sessions (id, metadata) VALUES (?, ?)",
		id, string(metadataJSON),
	)
	if err != nil {
		return fmt.Errorf("register session: %w", err)
	}

	// Health check and auto-start
	cfg, _ := config.Read()
	port := cfg.Port
	if port == 0 {
		return nil
	}

	healthURL := fmt.Sprintf("http://localhost:%d/health", port)
	online := isServerHealthy(healthURL)

	if !online {
		fmt.Fprintf(cmd.ErrOrStderr(), "warning: Oraculo HTTP server offline — auto-starting on port %d\n", port)
		if err := startHTTPDaemon(); err != nil {
			msg := fmt.Sprintf("warning: failed to auto-start Oraculo server: %v", err)
			fmt.Fprintln(cmd.ErrOrStderr(), msg)
			fmt.Fprintln(cmd.OutOrStdout(), msg)
			return nil
		}
		online = pollHealth(healthURL, 500*time.Millisecond, 10*time.Second)
		if !online {
			msg := "warning: Oraculo server started but not responding. Telemetry unavailable."
			fmt.Fprintln(cmd.ErrOrStderr(), msg)
			fmt.Fprintln(cmd.OutOrStdout(), msg)
			return nil
		}
	}

	// POST session-start
	postURL := fmt.Sprintf("http://localhost:%d/hooks/session-start", port)
	client := &http.Client{Timeout: 2 * time.Second}
	client.Post(postURL, "application/json", strings.NewReader(string(metadataJSON)))

	return nil
}

func isServerHealthy(url string) bool {
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(url)
	return err == nil && resp.StatusCode == http.StatusOK
}

func startHTTPDaemon() error {
	exe, err := os.Executable()
	if err != nil {
		exe = "oraculo"
	}
	cmd := exec.Command(exe, "start", "http")
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	cmd.Stdout = nil
	cmd.Stderr = nil
	cmd.Stdin = nil
	if err := cmd.Start(); err != nil {
		return err
	}
	return cmd.Process.Release()
}

func pollHealth(url string, interval, timeout time.Duration) bool {
	deadline := time.After(timeout)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-deadline:
			return false
		case <-ticker.C:
			if isServerHealthy(url) {
				return true
			}
		}
	}
}
```

**Step 6: Run test to verify it passes**

Run: `go test ./src/cli/ -run TestHookSessionStart -v`
Expected: PASS

**Step 7: Run all tests**

Run: `go test ./...`
Expected: PASS

**Step 8: Commit**

```bash
git add src/cli/hook_session.go src/cli/hook_session_test.go
git commit -m "feat: auto-start HTTP server from SessionStart hook"
```

---

### Task 4: Update install.go for new mcpServers args

**Files:**
- Modify: `src/cli/install.go:112-117` (change args to `["start", "mcp"]`)
- Modify: `src/cli/install_test.go` (verify new args)

**Step 1: Write the failing test**

Add to `src/cli/install_test.go`:

```go
func TestInstall_MCPServerArgs(t *testing.T) {
	setupInstallDir(t)
	_, err := installCmd(t)
	if err != nil {
		t.Fatalf("install: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(".claude", "settings.json"))
	if err != nil {
		t.Fatalf("read settings.json: %v", err)
	}

	var settings struct {
		MCPServers map[string]struct {
			Command string   `json:"command"`
			Args    []string `json:"args"`
		} `json:"mcpServers"`
	}
	if err := json.Unmarshal(data, &settings); err != nil {
		t.Fatalf("parse settings.json: %v", err)
	}

	oraculo, ok := settings.MCPServers["oraculo"]
	if !ok {
		t.Fatal("mcpServers.oraculo not found")
	}
	if len(oraculo.Args) != 2 || oraculo.Args[0] != "start" || oraculo.Args[1] != "mcp" {
		t.Errorf("expected args [start mcp], got %v", oraculo.Args)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./src/cli/ -run TestInstall_MCPServerArgs -v`
Expected: FAIL — args is `["start"]` not `["start", "mcp"]`

**Step 3: Update install.go**

In `src/cli/install.go`, change line 115:

```go
// Before:
Args: []string{"start"},

// After:
Args: []string{"start", "mcp"},
```

**Step 4: Run test to verify it passes**

Run: `go test ./src/cli/ -run TestInstall_MCPServerArgs -v`
Expected: PASS

**Step 5: Run all tests**

Run: `go test ./...`
Expected: PASS

**Step 6: Commit**

```bash
git add src/cli/install.go src/cli/install_test.go
git commit -m "feat: update mcpServers args to use start mcp subcommand"
```

---

### Task 5: Update settings.json hook format

The current `install.go` puts `matcher` at the wrong level for PostToolUse. Verify and fix if the nested structure matches Claude Code's expected format.

**Files:**
- Verify: `src/cli/install.go:107-110` (PostToolUse hook group structure)
- Verify: `src/cli/install_test.go` (ensure settings.json is valid)

**Step 1: Verify current format is correct**

Run: `go test ./src/cli/ -v`
Expected: PASS — existing tests validate the format

**Step 2: Commit (if no changes needed)**

If all tests pass, this task is a no-op verification. Move to next task.

---

### Task 6: End-to-end verification

**Step 1: Build the binary**

Run: `go build -o oraculo .`

**Step 2: Test `oraculo start http` starts and serves health**

```bash
./oraculo start http &
sleep 1
curl -s http://localhost:3100/health
# Expected: {"status":"ok"}
kill %1
```

**Step 3: Test `oraculo start mcp` starts on stdio**

```bash
echo '{"jsonrpc":"2.0","method":"initialize","id":1,"params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"1.0"}}}' | timeout 3 ./oraculo start mcp 2>/dev/null | head -1
# Expected: JSON-RPC response with server info
```

**Step 4: Test `oraculo start` runs both (backwards compat)**

```bash
./oraculo start &
sleep 1
curl -s http://localhost:3100/health
# Expected: {"status":"ok"}
kill %1
```

**Step 5: Test auto-start from session-start hook**

```bash
# Make sure no server is running
pkill -f "oraculo start http" 2>/dev/null
sleep 1

# Run session-start hook — should auto-start server
./oraculo hook session-start 2>&1
# Expected: "warning: Oraculo HTTP server offline — auto-starting on port 3100"

# Verify server is now running
curl -s http://localhost:3100/health
# Expected: {"status":"ok"}

pkill -f "oraculo start http"
```

**Step 6: Test idle timeout**

```bash
# Start with short idle timeout for testing (modify defaultIdleTimeout temporarily)
# Or test via unit tests (Task 1 already covers this)
```

**Step 7: Final commit**

```bash
git add -A
git commit -m "chore: end-to-end verification of server lifecycle guardian"
```
