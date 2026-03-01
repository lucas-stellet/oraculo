package integration_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/lucas/oraculo/src/approval"
	"github.com/lucas/oraculo/src/db"
	"github.com/lucas/oraculo/src/dbtest"
	"github.com/lucas/oraculo/src/domain"
	"github.com/lucas/oraculo/src/server"
	"github.com/lucas/oraculo/src/ws"
)

func TestIntegration_ApprovalFlow(t *testing.T) {
	// 1. Set up all components.
	database := dbtest.Open(t)

	// Create an epic (required for the epic-requirements approval type).
	epicStore := db.NewEpicStore(database)
	epic, _, err := epicStore.Create("test-epic", "An epic for integration testing")
	if err != nil {
		t.Fatalf("create epic: %v", err)
	}

	hub := ws.NewHub()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	go hub.Run(ctx)

	approvalStore := db.NewApprovalStore(database)
	bridge := approval.NewBridge(approvalStore, hub)
	srv := server.New(database, bridge, hub)

	// 2. Start the HTTP server as a test server.
	httpSrv := httptest.NewServer(srv)
	defer httpSrv.Close()

	// 3. In a goroutine, simulate an MCP tool calling bridge.Request().
	type requestResult struct {
		result *approval.VerdictResult
		err    error
	}
	resultCh := make(chan requestResult, 1)

	go func() {
		r, err := bridge.Request(ctx, approval.ApprovalRequest{
			Type:    domain.ApprovalEpicRequirements,
			EpicID:  &epic.ID,
			Content: "# Requirements\n\nThis is the requirements document.",
		})
		resultCh <- requestResult{r, err}
	}()

	// 4. Wait for the approval to be persisted to the database.
	time.Sleep(200 * time.Millisecond)

	// 5. GET /api/approvals?status=pending — verify the approval shows up.
	resp, err := http.Get(httpSrv.URL + "/api/approvals?status=pending")
	if err != nil {
		t.Fatalf("GET /api/approvals: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET /api/approvals: status %d", resp.StatusCode)
	}

	var approvals []map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&approvals); err != nil {
		t.Fatalf("decode approvals: %v", err)
	}
	if len(approvals) == 0 {
		t.Fatal("expected at least 1 pending approval")
	}

	// domain.Approval has no JSON tags, so the JSON key matches the field name "ID".
	approvalID, ok := approvals[0]["ID"].(string)
	if !ok || approvalID == "" {
		t.Fatalf("missing or invalid approval ID, got: %v", approvals[0]["ID"])
	}

	// 6. POST /api/approvals/{id}/verdict — approve it.
	verdictBody := `{"verdict":"approved","comment":"LGTM"}`
	verdictResp, err := http.Post(
		httpSrv.URL+"/api/approvals/"+approvalID+"/verdict",
		"application/json",
		strings.NewReader(verdictBody),
	)
	if err != nil {
		t.Fatalf("POST verdict: %v", err)
	}
	defer verdictResp.Body.Close()
	if verdictResp.StatusCode != http.StatusOK {
		t.Fatalf("POST verdict: status %d", verdictResp.StatusCode)
	}

	// 7. Verify bridge.Request() returned with the correct verdict.
	select {
	case res := <-resultCh:
		if res.err != nil {
			t.Fatalf("bridge.Request returned error: %v", res.err)
		}
		if res.result.Verdict != domain.VerdictApproved {
			t.Errorf("Verdict = %q, want %q", res.result.Verdict, domain.VerdictApproved)
		}
		if res.result.Comment != "LGTM" {
			t.Errorf("Comment = %q, want %q", res.result.Comment, "LGTM")
		}
	case <-ctx.Done():
		t.Fatal("timed out waiting for bridge.Request to return")
	}
}
