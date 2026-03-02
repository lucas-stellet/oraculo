package server_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/lucas/oraculo/src/applog"
	"github.com/lucas/oraculo/src/approval"
	"github.com/lucas/oraculo/src/db"
	"github.com/lucas/oraculo/src/dbtest"
	"github.com/lucas/oraculo/src/domain"
	"github.com/lucas/oraculo/src/server"
	"github.com/lucas/oraculo/src/ws"
)

func testServerWithDB(t *testing.T) (*server.Server, *db.DB) {
	t.Helper()
	database := dbtest.Open(t)
	hub := ws.NewHub()
	bridge := approval.NewBridge(db.NewApprovalStore(database), hub)
	return server.New(database, bridge, hub, nil), database
}

func TestListEpics_Empty(t *testing.T) {
	srv, _ := testServerWithDB(t)
	req := httptest.NewRequest(http.MethodGet, "/api/epics", nil)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d, want 200", rec.Code)
	}
	var result []domain.Epic
	if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(result) != 0 {
		t.Fatalf("expected 0 epics, got %d", len(result))
	}
}

func TestListEpics_WithData(t *testing.T) {
	srv, database := testServerWithDB(t)

	// Seed an epic directly via store.
	epicStore := db.NewEpicStore(database)
	_, _, err := epicStore.Create("my-epic", "Epic description")
	if err != nil {
		t.Fatalf("create epic: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/epics", nil)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d, want 200", rec.Code)
	}
	var result []domain.Epic
	if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 epic, got %d", len(result))
	}
	if result[0].Name != "my-epic" {
		t.Errorf("epic name = %q, want %q", result[0].Name, "my-epic")
	}
}

func TestListApprovals_Empty(t *testing.T) {
	srv, _ := testServerWithDB(t)
	req := httptest.NewRequest(http.MethodGet, "/api/approvals", nil)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d, want 200", rec.Code)
	}
	var result []domain.Approval
	if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(result) != 0 {
		t.Fatalf("expected 0 approvals, got %d", len(result))
	}
}

func TestListApprovals_PendingFilter(t *testing.T) {
	srv, database := testServerWithDB(t)

	// Seed approvals via store.
	approvalStore := db.NewApprovalStore(database)
	_, err := approvalStore.Request(domain.ApprovalEpicRequirements, nil, nil, "content-1")
	if err != nil {
		t.Fatalf("request approval 1: %v", err)
	}
	appr2, err := approvalStore.Request(domain.ApprovalStoryDefinition, nil, nil, "content-2")
	if err != nil {
		t.Fatalf("request approval 2: %v", err)
	}
	// Decide on appr2 so it's no longer pending.
	_, err = approvalStore.Verdict(appr2.ID, domain.VerdictApproved, "looks good")
	if err != nil {
		t.Fatalf("verdict approval 2: %v", err)
	}

	// Without filter: should return 2.
	req := httptest.NewRequest(http.MethodGet, "/api/approvals", nil)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("list all status %d, want 200", rec.Code)
	}
	var all []domain.Approval
	if err := json.NewDecoder(rec.Body).Decode(&all); err != nil {
		t.Fatalf("decode all: %v", err)
	}
	if len(all) != 2 {
		t.Fatalf("expected 2 approvals, got %d", len(all))
	}

	// With ?status=pending: should return 1.
	req2 := httptest.NewRequest(http.MethodGet, "/api/approvals?status=pending", nil)
	rec2 := httptest.NewRecorder()
	srv.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusOK {
		t.Fatalf("list pending status %d, want 200", rec2.Code)
	}
	var pending []domain.Approval
	if err := json.NewDecoder(rec2.Body).Decode(&pending); err != nil {
		t.Fatalf("decode pending: %v", err)
	}
	if len(pending) != 1 {
		t.Fatalf("expected 1 pending approval, got %d", len(pending))
	}
	if pending[0].Status != domain.ApprovalPending {
		t.Errorf("approval status = %q, want pending", pending[0].Status)
	}
}

func TestVerdict_ApprovesPendingApproval(t *testing.T) {
	srv, database := testServerWithDB(t)

	// Seed an approval.
	approvalStore := db.NewApprovalStore(database)
	appr, err := approvalStore.Request(domain.ApprovalEpicRequirements, nil, nil, "some content")
	if err != nil {
		t.Fatalf("request approval: %v", err)
	}

	// Submit verdict via API.
	body := `{"verdict":"approved","comment":"LGTM"}`
	path := "/api/approvals/" + appr.ID + "/verdict"
	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("verdict status %d, want 200; body: %s", rec.Code, rec.Body.String())
	}

	var updated domain.Approval
	if err := json.NewDecoder(rec.Body).Decode(&updated); err != nil {
		t.Fatalf("decode verdict response: %v", err)
	}
	if updated.Status != domain.ApprovalApproved {
		t.Errorf("approval status = %q, want approved", updated.Status)
	}
	if updated.VerdictComment != "LGTM" {
		t.Errorf("verdict comment = %q, want LGTM", updated.VerdictComment)
	}
}

func TestLogsEndpoint_ReturnsSSE(t *testing.T) {
	database := dbtest.Open(t)
	hub := ws.NewHub()
	bridge := approval.NewBridge(db.NewApprovalStore(database), hub)
	logs := applog.NewBroadcaster(io.Discard)
	srv := server.New(database, bridge, hub, logs)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately so ServeSSE returns after replay
	req := httptest.NewRequest(http.MethodGet, "/logs", nil).WithContext(ctx)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("GET /logs status %d, want 200", rec.Code)
	}
	ct := rec.Header().Get("Content-Type")
	if !strings.Contains(ct, "text/event-stream") {
		t.Errorf("Content-Type = %q, want text/event-stream", ct)
	}
}

func TestVerdict_InvalidVerdict(t *testing.T) {
	srv, database := testServerWithDB(t)

	approvalStore := db.NewApprovalStore(database)
	appr, err := approvalStore.Request(domain.ApprovalEpicRequirements, nil, nil, "content")
	if err != nil {
		t.Fatalf("request approval: %v", err)
	}

	body := `{"verdict":"invalid_verdict","comment":""}`
	path := "/api/approvals/" + appr.ID + "/verdict"
	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status %d, want 400", rec.Code)
	}
}
