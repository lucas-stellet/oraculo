// Package approval provides the Bridge that connects MCP approval tools to the HTTP server.
// An MCP tool calls Bridge.Request which blocks until a human submits a verdict via the
// HTTP server, which calls Bridge.Decide to unblock the waiting goroutine.
package approval

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/lucas/oraculo/src/db"
	"github.com/lucas/oraculo/src/domain"
)

// Broadcaster is satisfied by *ws.Hub. It broadcasts a JSON message to all
// connected WebSocket clients.
type Broadcaster interface {
	Broadcast([]byte)
}

// ApprovalRequest carries the parameters for a new approval gate.
type ApprovalRequest struct {
	// Type identifies the kind of approval (epic-requirements, story-definition, etc.).
	Type domain.ApprovalType

	// EpicID is the numeric ID of the parent epic (required for most approval types).
	EpicID *int

	// StoryID is the numeric ID of the parent story (nil for epic-level approvals).
	StoryID *int

	// Content is the markdown artifact to be reviewed.
	Content string
}

// VerdictResult is returned by Bridge.Request once a human has decided.
type VerdictResult struct {
	// ID is the UUID of the approval record.
	ID string

	// Verdict is the human's decision.
	Verdict domain.Verdict

	// Comment is optional human feedback.
	Comment string
}

// Bridge connects the blocking MCP approval tools to the non-blocking HTTP verdict endpoint.
// Each pending approval has a dedicated Go channel; Bridge.Decide sends on the channel to
// unblock the goroutine that called Bridge.Request.
type Bridge struct {
	store       *db.ApprovalStore
	broadcaster Broadcaster

	mu       sync.Mutex
	channels map[string]chan VerdictResult
}

// NewBridge returns a Bridge backed by the given store and broadcaster.
func NewBridge(store *db.ApprovalStore, broadcaster Broadcaster) *Bridge {
	return &Bridge{
		store:       store,
		broadcaster: broadcaster,
		channels:    make(map[string]chan VerdictResult),
	}
}

// Request creates a pending approval record, broadcasts a WebSocket notification, and
// blocks until a verdict is received or ctx is cancelled.
// It returns the verdict from the human reviewer.
func (b *Bridge) Request(ctx context.Context, req ApprovalRequest) (*VerdictResult, error) {
	approval, err := b.store.Request(req.Type, req.EpicID, req.StoryID, req.Content)
	if err != nil {
		return nil, fmt.Errorf("approval.Bridge.Request: %w", err)
	}

	// Create a buffered channel so that Decide never blocks even if this goroutine
	// has already been cancelled.
	ch := make(chan VerdictResult, 1)

	b.mu.Lock()
	b.channels[approval.ID] = ch
	b.mu.Unlock()

	defer func() {
		b.mu.Lock()
		delete(b.channels, approval.ID)
		b.mu.Unlock()
	}()

	// Broadcast WebSocket notification so the dashboard can render the artifact.
	b.broadcastApprovalRequested(approval)

	select {
	case result := <-ch:
		return &result, nil
	case <-ctx.Done():
		return nil, fmt.Errorf("approval.Bridge.Request: context cancelled: %w", ctx.Err())
	}
}

// Decide records the verdict in the database and unblocks the goroutine waiting in Request.
// It returns domain.ErrNotFound if the approval does not exist, and domain.ErrInvalidTransition
// if the approval is not pending.
func (b *Bridge) Decide(id string, verdict domain.Verdict, comment string) (*domain.Approval, error) {
	approval, err := b.store.Verdict(id, verdict, comment)
	if err != nil {
		return nil, fmt.Errorf("approval.Bridge.Decide: %w", err)
	}

	// Unblock the waiting goroutine if there is one.
	b.mu.Lock()
	ch, ok := b.channels[id]
	b.mu.Unlock()

	if ok {
		ch <- VerdictResult{
			ID:      id,
			Verdict: verdict,
			Comment: comment,
		}
	}

	// Broadcast WebSocket notification so the dashboard can update its UI.
	b.broadcastApprovalDecided(approval)

	return approval, nil
}

// Status returns the current state of an approval without blocking.
func (b *Bridge) Status(id string) (*domain.Approval, error) {
	approval, err := b.store.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("approval.Bridge.Status: %w", err)
	}
	return approval, nil
}

// broadcastApprovalRequested sends an "approval_requested" WebSocket event.
func (b *Bridge) broadcastApprovalRequested(a *domain.Approval) {
	b.broadcast("approval_requested", a)
}

// broadcastApprovalDecided sends an "approval_decided" WebSocket event.
func (b *Bridge) broadcastApprovalDecided(a *domain.Approval) {
	b.broadcast("approval_decided", a)
}

func (b *Bridge) broadcast(eventType string, payload any) {
	msg, err := json.Marshal(map[string]any{
		"type":    eventType,
		"payload": payload,
	})
	if err != nil {
		return
	}
	b.broadcaster.Broadcast(msg)
}
