package approval_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/lucas/oraculo/src/approval"
	"github.com/lucas/oraculo/src/db"
	"github.com/lucas/oraculo/src/dbtest"
	"github.com/lucas/oraculo/src/domain"
)

type stubBroadcaster struct {
	msgs [][]byte
	mu   sync.Mutex
}

func (b *stubBroadcaster) Broadcast(msg []byte) {
	b.mu.Lock()
	b.msgs = append(b.msgs, msg)
	b.mu.Unlock()
}

func TestBridge_RequestAndDecide(t *testing.T) {
	database := dbtest.Open(t)
	epicStore := db.NewEpicStore(database)
	epicStore.Create("my-epic", "desc")

	bridge := approval.NewBridge(db.NewApprovalStore(database), &stubBroadcaster{})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var result *approval.VerdictResult
	var reqErr error
	done := make(chan struct{})

	go func() {
		defer close(done)
		result, reqErr = bridge.Request(ctx, approval.ApprovalRequest{
			Type:    domain.ApprovalEpicRequirements,
			Epic:    "my-epic",
			Content: "# Requirements\nSome content",
		})
	}()

	// Wait for approval to appear in DB
	time.Sleep(100 * time.Millisecond)

	// Get the pending approval
	approvals, _ := db.NewApprovalStore(database).List(true)
	if len(approvals) == 0 {
		t.Fatal("no pending approvals found")
	}

	// Decide
	err := bridge.Decide(approvals[0].ID, domain.VerdictApproved, "looks good")
	if err != nil {
		t.Fatalf("Decide: %v", err)
	}

	<-done
	if reqErr != nil {
		t.Fatalf("Request: %v", reqErr)
	}
	if result.Verdict != domain.VerdictApproved {
		t.Errorf("Verdict = %q, want %q", result.Verdict, domain.VerdictApproved)
	}
}

func TestBridge_ContextCancellation(t *testing.T) {
	database := dbtest.Open(t)
	epicStore := db.NewEpicStore(database)
	epicStore.Create("my-epic", "desc")

	bridge := approval.NewBridge(db.NewApprovalStore(database), &stubBroadcaster{})

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		defer close(done)
		_, err := bridge.Request(ctx, approval.ApprovalRequest{
			Type:    domain.ApprovalEpicRequirements,
			Epic:    "my-epic",
			Content: "content",
		})
		if err != context.Canceled {
			t.Errorf("expected context.Canceled, got: %v", err)
		}
	}()

	time.Sleep(100 * time.Millisecond)
	cancel()
	<-done
}

func TestBridge_Status(t *testing.T) {
	database := dbtest.Open(t)
	epicStore := db.NewEpicStore(database)
	epicStore.Create("my-epic", "desc")

	store := db.NewApprovalStore(database)
	bridge := approval.NewBridge(store, &stubBroadcaster{})

	appr, _ := store.Request(domain.ApprovalEpicRequirements, nil, nil, "content")

	status, err := bridge.Status(appr.ID)
	if err != nil {
		t.Fatalf("Status: %v", err)
	}
	if status.Status != domain.ApprovalPending {
		t.Errorf("Status = %q, want %q", status.Status, domain.ApprovalPending)
	}
}
