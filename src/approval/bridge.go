// Package approval provides the Bridge, which coordinates approval requests
// between agents (requestors) and human operators (deciders).
//
// An agent calls Bridge.Request, which creates an approval record and blocks
// until a human calls Bridge.Decide with a verdict. The Bridge uses a
// Broadcaster to notify the UI of state changes in real time.
package approval

import (
	"context"
	"fmt"
	"sync"

	"github.com/lucas/oraculo/src/db"
	"github.com/lucas/oraculo/src/domain"
)

// Broadcaster is implemented by the SSE/WebSocket layer to push events to
// connected UI clients. Pass a no-op implementation when broadcasting is
// not needed (e.g. in tests).
type Broadcaster interface {
	Broadcast(event string, data any)
}

// ApprovalRequest carries the parameters for a new approval gate.
type ApprovalRequest struct {
	Type    domain.ApprovalType
	EpicID  *int
	StoryID *int
	Content string
}

// VerdictResult is returned by Bridge.Request once a verdict is recorded.
type VerdictResult struct {
	Approval *domain.Approval
}

// Bridge coordinates approval requests between agents and human operators.
// Agents call Request (which blocks) and operators call Decide (which unblocks).
type Bridge struct {
	store       *db.ApprovalStore
	broadcaster Broadcaster

	mu      sync.Mutex
	waiters map[string]chan struct{}
}

// NewBridge creates a new Bridge backed by the given store and broadcaster.
func NewBridge(store *db.ApprovalStore, broadcaster Broadcaster) *Bridge {
	return &Bridge{
		store:       store,
		broadcaster: broadcaster,
		waiters:     make(map[string]chan struct{}),
	}
}

// Request creates a new approval and blocks until a verdict is recorded or the
// context is cancelled. It returns the final approval state on success.
func (b *Bridge) Request(ctx context.Context, req ApprovalRequest) (*VerdictResult, error) {
	approval, err := b.store.Request(req.Type, req.EpicID, req.StoryID, req.Content)
	if err != nil {
		return nil, fmt.Errorf("create approval: %w", err)
	}

	ch := make(chan struct{}, 1)
	b.mu.Lock()
	b.waiters[approval.ID] = ch
	b.mu.Unlock()

	defer func() {
		b.mu.Lock()
		delete(b.waiters, approval.ID)
		b.mu.Unlock()
	}()

	b.broadcaster.Broadcast("approval_requested", approval)

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-ch:
	}

	final, err := b.store.GetByID(approval.ID)
	if err != nil {
		return nil, fmt.Errorf("get final approval: %w", err)
	}
	return &VerdictResult{Approval: final}, nil
}

// Decide records a verdict on the given approval and unblocks any waiting
// Request call. It returns the updated approval.
func (b *Bridge) Decide(id string, verdict domain.Verdict, comment string) (*domain.Approval, error) {
	approval, err := b.store.Verdict(id, verdict, comment)
	if err != nil {
		return nil, err
	}

	b.mu.Lock()
	ch, ok := b.waiters[id]
	b.mu.Unlock()

	if ok {
		select {
		case ch <- struct{}{}:
		default:
		}
	}

	b.broadcaster.Broadcast("approval_decided", approval)
	return approval, nil
}

// Status returns the current state of the approval with the given ID.
func (b *Bridge) Status(id string) (*domain.Approval, error) {
	return b.store.GetByID(id)
}
