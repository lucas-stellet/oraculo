// src/approval/bridge.go
package approval

import (
	"context"

	"github.com/lucas/oraculo/src/domain"
)

// Store is the interface for persisting and retrieving approvals.
type Store interface {
	Request(approvalType domain.ApprovalType, epicID, storyID *int, content string) (*domain.Approval, error)
	GetByID(id string) (*domain.Approval, error)
	Verdict(id string, verdict domain.Verdict, comment string) (*domain.Approval, error)
}

// Broadcaster is the interface for notifying connected UI clients.
type Broadcaster interface {
	Broadcast(msg []byte)
}

// Bridge connects the MCP tool layer with the HTTP/WebSocket layer for
// real-time approval workflows.
type Bridge struct {
	store       Store
	broadcaster Broadcaster
}

// NewBridge creates a new Bridge backed by the given store and broadcaster.
func NewBridge(store Store, broadcaster Broadcaster) *Bridge {
	return &Bridge{store: store, broadcaster: broadcaster}
}

// ApprovalRequest carries the parameters for requesting a human approval.
type ApprovalRequest struct {
	Type    domain.ApprovalType
	EpicID  *int
	StoryID *int
	Content string
}

// Request creates a new approval and blocks until a verdict is recorded.
func (b *Bridge) Request(ctx context.Context, req ApprovalRequest) (*domain.Approval, error) {
	return b.store.Request(req.Type, req.EpicID, req.StoryID, req.Content)
}

// Decide records a verdict for the approval identified by id.
func (b *Bridge) Decide(id string, verdict domain.Verdict, comment string) (*domain.Approval, error) {
	return b.store.Verdict(id, verdict, comment)
}
