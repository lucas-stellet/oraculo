# Epic Selection Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Entregar a primeira tela do dashboard como `Landing (Epic Selection)`, com listagem agregada de épicos e criação de novo épico via trust layer.

**Architecture:** A primeira fatia será servida pelo próprio HTTP server em Go usando `html/template` + assets embutidos, enquanto o backend expõe um contrato HTTP específico para a tela. A mutação de criação de épico será consolidada em uma operação interna compartilhada para que CLI e dashboard produzam o mesmo efeito: registro no SQLite e pasta materializada em `.oraculo/epics/<nome>`.

**Tech Stack:** Go, `net/http`, `html/template`, `embed`, SQLite, stores existentes em `src/db`, testes com `httptest`.

---

### Task 1: Consolidate Epic Creation in the Trust Layer

**Files:**
- Create: `src/project/epics.go`
- Create: `src/project/epics_test.go`
- Modify: `src/cli/tools/epic.go`
- Modify: `src/cli/tools/epic_test.go`

**Step 1: Write the failing tests**

Add a package-level test for the shared creation flow and extend the CLI test so `tools epic init` must also create the epic directory.

```go
func TestEnsureEpic_CreatesDatabaseRowAndDirectory(t *testing.T) {
    database := dbtest.Open(t)
    root := t.TempDir()

    epic, created, err := project.EnsureEpic(root, database, "checkout-redesign", "Refresh checkout UX")
    if err != nil {
        t.Fatalf("EnsureEpic: %v", err)
    }

    if !created {
        t.Fatal("expected epic to be created")
    }

    path := filepath.Join(root, ".oraculo", "epics", "checkout-redesign")
    if _, err := os.Stat(path); err != nil {
        t.Fatalf("expected epic directory: %v", err)
    }

    if epic.Name != "checkout-redesign" {
        t.Fatalf("unexpected epic name: %s", epic.Name)
    }
}
```

Extend `TestEpicInit` so it asserts the directory exists after the command runs.

**Step 2: Run the tests to verify they fail**

Run: `go test ./src/project ./src/cli/tools -run 'TestEnsureEpic|TestEpicInit'`

Expected: FAIL because `src/project` does not exist yet and `epic init` does not create the epic directory.

**Step 3: Write the minimal implementation**

Create a reusable helper that wraps store creation plus filesystem materialization.

```go
package project

func EnsureEpic(root string, database *db.DB, name, description string) (*domain.Epic, bool, error) {
    epic, created, err := db.NewEpicStore(database).Create(name, description)
    if err != nil {
        return nil, false, err
    }

    dir := filepath.Join(root, ".oraculo", "epics", name)
    if err := os.MkdirAll(dir, 0o755); err != nil {
        return nil, false, fmt.Errorf("create epic directory: %w", err)
    }

    return epic, created, nil
}
```

Update `newEpicInitCmd` to call that helper with `"."` as the workspace root.

**Step 4: Run the tests to verify they pass**

Run: `go test ./src/project ./src/cli/tools -run 'TestEnsureEpic|TestEpicInit'`

Expected: PASS.

**Step 5: Commit**

```bash
git add src/project/epics.go src/project/epics_test.go src/cli/tools/epic.go src/cli/tools/epic_test.go
git commit -m "feat: unify epic creation flow for cli and dashboard"
```

### Task 2: Expose an Epic Selection API Contract

**Files:**
- Create: `src/server/epic_selection.go`
- Modify: `src/server/api.go`
- Modify: `src/server/server.go`
- Modify: `src/server/api_test.go`

**Step 1: Write the failing tests**

Add one test for the enriched epic list payload and one for epic creation through the API.

```go
func TestListEpics_ReturnsEpicSelectionCards(t *testing.T) {
    srv, database := testServerWithDB(t)

    epic, _, _ := db.NewEpicStore(database).Create("checkout-redesign", "Refresh checkout UX")
    story, _, _ := db.NewStoryStore(database).Create(epic.ID, "payment-step", "")
    _, _, _ = db.NewTaskStore(database).Create(story.ID, "instrument-api", "", nil)

    req := httptest.NewRequest(http.MethodGet, "/api/epics", nil)
    rec := httptest.NewRecorder()
    srv.ServeHTTP(rec, req)

    var cards []map[string]any
    json.NewDecoder(rec.Body).Decode(&cards)

    if cards[0]["story_count"].(float64) != 1 {
        t.Fatalf("expected story_count=1, got %#v", cards[0]["story_count"])
    }
}

func TestCreateEpic_CreatesDirectoryAndRow(t *testing.T) {
    srv, _ := testServerWithDB(t)

    body := `{"name":"checkout-redesign","description":"Refresh checkout UX"}`
    req := httptest.NewRequest(http.MethodPost, "/api/epics", strings.NewReader(body))
    req.Header.Set("Content-Type", "application/json")
    rec := httptest.NewRecorder()
    srv.ServeHTTP(rec, req)

    if rec.Code != http.StatusOK {
        t.Fatalf("status %d body=%s", rec.Code, rec.Body.String())
    }
}
```

**Step 2: Run the tests to verify they fail**

Run: `go test ./src/server -run 'TestListEpics_ReturnsEpicSelectionCards|TestCreateEpic_CreatesDirectoryAndRow'`

Expected: FAIL because `GET /api/epics` still returns raw `domain.Epic` rows and `POST /api/epics` does not exist.

**Step 3: Write the minimal implementation**

Create a view-model assembler dedicated to the landing screen.

```go
type EpicCard struct {
    Name               string  `json:"name"`
    Description        string  `json:"description"`
    ApprovalStatus     string  `json:"approval_status"`
    CurrentPhase       string  `json:"current_phase"`
    StoryCount         int     `json:"story_count"`
    TaskCount          int     `json:"task_count"`
    CompletedTaskCount int     `json:"completed_task_count"`
    CompletionRatio    float64 `json:"completion_ratio"`
    OverallStatus      string  `json:"overall_status"`
}
```

Implement `POST /api/epics` in `APIHandler` using the shared `project.EnsureEpic` helper, and update `GET /api/epics` to return `[]EpicCard`.

For phase derivation, start with a small helper:

```go
func deriveEpicPhase(sessionStore *db.SessionStore, epicID int) string {
    for _, candidate := range []domain.SessionType{
        domain.SessionValidate,
        domain.SessionExecute,
        domain.SessionPlan,
        domain.SessionStory,
        domain.SessionEpic,
    } {
        if _, err := sessionStore.ActiveByEpic(epicID, candidate); err == nil {
            return mapPhase(candidate)
        }
    }
    return "Discover"
}
```

**Step 4: Run the tests to verify they pass**

Run: `go test ./src/server -run 'TestListEpics_ReturnsEpicSelectionCards|TestCreateEpic_CreatesDirectoryAndRow'`

Expected: PASS.

**Step 5: Commit**

```bash
git add src/server/epic_selection.go src/server/api.go src/server/server.go src/server/api_test.go
git commit -m "feat: add epic selection api contract"
```

### Task 3: Serve the Landing UI from the Go HTTP Server

**Files:**
- Create: `src/server/ui.go`
- Create: `src/server/ui_test.go`
- Create: `src/server/assets/epic-selection.html`
- Create: `src/server/assets/epic-selection.css`
- Create: `src/server/assets/epic-selection.js`
- Modify: `src/server/server.go`

**Step 1: Write the failing tests**

Add a test that hits `/` and asserts the response is HTML and contains the Epic Selection shell.

```go
func TestLandingUI_RendersEpicSelectionShell(t *testing.T) {
    srv := testServer(t)

    req := httptest.NewRequest(http.MethodGet, "/", nil)
    rec := httptest.NewRecorder()
    srv.ServeHTTP(rec, req)

    if rec.Code != http.StatusOK {
        t.Fatalf("status %d", rec.Code)
    }

    body := rec.Body.String()
    if !strings.Contains(body, "Select an Epic") {
        t.Fatalf("missing title in body: %s", body)
    }
    if !strings.Contains(body, "Create Epic") {
        t.Fatalf("missing primary action in body: %s", body)
    }
}
```

**Step 2: Run the tests to verify they fail**

Run: `go test ./src/server -run TestLandingUI_RendersEpicSelectionShell`

Expected: FAIL because `/` is not registered.

**Step 3: Write the minimal implementation**

Serve a small HTML shell plus embedded CSS/JS assets.

```go
//go:embed assets/epic-selection.html assets/epic-selection.css assets/epic-selection.js
var uiFS embed.FS

func (s *UIHandler) handleLanding(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "text/html; charset=utf-8")
    html, _ := uiFS.ReadFile("assets/epic-selection.html")
    w.Write(html)
}
```

In CSS, encode the design-system foundations as custom properties:

```css
:root {
  --bg-canvas: #0f1720;
  --bg-card: #16202b;
  --border-subtle: #243244;
  --text-strong: #f3f6fb;
  --text-muted: #9fb0c3;
  --accent-blue: #4c8bf5;
  --success-green: #3fbf7f;
  --warning-amber: #e0a94a;
  --danger-red: #d05c5c;
}
```

In JS, fetch `/api/epics`, render the grid, open the create modal, and POST new epics back to the API.

**Step 4: Run the tests to verify they pass**

Run: `go test ./src/server -run TestLandingUI_RendersEpicSelectionShell`

Expected: PASS.

**Step 5: Commit**

```bash
git add src/server/ui.go src/server/ui_test.go src/server/assets/epic-selection.html src/server/assets/epic-selection.css src/server/assets/epic-selection.js src/server/server.go
git commit -m "feat: serve epic selection landing ui"
```

### Task 4: Verify the Vertical Slice End-to-End

**Files:**
- Verify: `src/project/epics_test.go`
- Verify: `src/server/api_test.go`
- Verify: `src/server/ui_test.go`
- Verify: `src/server/server_e2e_test.go`

**Step 1: Add or extend end-to-end coverage**

Add an HTTP-level integration test that:

- starts the server
- `POST /api/epics`
- `GET /api/epics`
- checks the created epic appears with zeroed counts and `Discover` phase

```go
func TestEpicSelectionFlow_EndToEnd(t *testing.T) {
    srv, _ := testServerWithDB(t)

    createReq := httptest.NewRequest(http.MethodPost, "/api/epics", strings.NewReader(
        `{"name":"checkout-redesign","description":"Refresh checkout UX"}`,
    ))
    createReq.Header.Set("Content-Type", "application/json")
    createRec := httptest.NewRecorder()
    srv.ServeHTTP(createRec, createReq)

    listReq := httptest.NewRequest(http.MethodGet, "/api/epics", nil)
    listRec := httptest.NewRecorder()
    srv.ServeHTTP(listRec, listReq)

    if !strings.Contains(listRec.Body.String(), "checkout-redesign") {
        t.Fatalf("expected created epic in list: %s", listRec.Body.String())
    }
}
```

**Step 2: Run the tests to verify the full slice**

Run: `go test ./src/project ./src/server ./src/cli/tools`

Expected: PASS.

**Step 3: Run broader regression**

Run: `go test ./...`

Expected: PASS.

**Step 4: Manual verification**

Run: `go run . start http`

Expected:
- `GET /` loads the Epic Selection landing page
- existing epics render as cards
- creating a new epic adds `.oraculo/epics/<name>`
- the new epic appears immediately in the grid

**Step 5: Commit**

```bash
git add .
git commit -m "feat: deliver epic selection dashboard slice"
```
