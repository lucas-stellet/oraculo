// Package approval coordinates human-in-the-loop approval gates between
// MCP tools (callers of Request) and HTTP handlers (callers of Decide).
package approval

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/lucas/oraculo/apps/backend/src/domain"
)

// Broadcaster sends event messages to connected observers (e.g. WebSocket clients).
type Broadcaster interface {
	Broadcast(msg []byte)
}

// ApprovalRequest carries the parameters for a new approval gate.
type ApprovalRequest struct {
	Type    domain.ApprovalType
	EpicID  *int   // numeric epic ID; may be nil
	StoryID *int   // numeric story ID; may be nil
	Content string
}

// VerdictResult is returned by Request once a decision has been recorded.
type VerdictResult struct {
	ID       string
	Type     domain.ApprovalType
	EpicID   *int
	StoryID  *int
	Content  string
	Verdict  domain.Verdict
	Comment  string
	Comments []domain.ApprovalComment
}

// store is the subset of db.ApprovalStore used by Bridge.
type store interface {
	Request(approvalType domain.ApprovalType, epicID, storyID *int, content string) (*domain.Approval, error)
	Verdict(id string, verdict domain.Verdict, comment string) (*domain.Approval, error)
	GetByID(id string) (*domain.Approval, error)
	ListComments(approvalID string) ([]domain.ApprovalComment, error)
	DeleteCommentsByApproval(approvalID string) error
}

// Bridge connects approval requestors (blocking goroutines) with approval
// deciders (HTTP handlers or CLI commands) via in-process Go channels.
type Bridge struct {
	store       store
	broadcaster Broadcaster

	mu      sync.Mutex
	pending map[string]chan VerdictResult
}

// NewBridge creates a Bridge backed by the given store and broadcaster.
func NewBridge(s store, b Broadcaster) *Bridge {
	return &Bridge{
		store:       s,
		broadcaster: b,
		pending:     make(map[string]chan VerdictResult),
	}
}

// Request persists a new approval gate in the database, broadcasts an
// "approval_requested" event, then blocks until a verdict is delivered via
// Decide or the context is cancelled.
func (br *Bridge) Request(ctx context.Context, req ApprovalRequest) (*VerdictResult, error) {
	appr, err := br.store.Request(req.Type, req.EpicID, req.StoryID, req.Content)
	if err != nil {
		return nil, fmt.Errorf("approval request: %w", err)
	}

	// Register a channel so Decide can unblock us.
	ch := make(chan VerdictResult, 1)
	br.mu.Lock()
	br.pending[appr.ID] = ch
	br.mu.Unlock()

	br.broadcast("approval_requested", appr.ID)

	defer func() {
		br.mu.Lock()
		delete(br.pending, appr.ID)
		br.mu.Unlock()
	}()

	select {
	case verdict := <-ch:
		return &verdict, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// Decide records a verdict for the approval identified by id, then unblocks
// any goroutine waiting in Request for that id.
func (br *Bridge) Decide(id string, verdict domain.Verdict, comment string) error {
	appr, err := br.store.Verdict(id, verdict, comment)
	if err != nil {
		return fmt.Errorf("verdict: %w", err)
	}

	var comments []domain.ApprovalComment

	if verdict == domain.VerdictApproved {
		// Delete all comments for approved approvals — they are no longer relevant.
		if err := br.store.DeleteCommentsByApproval(id); err != nil {
			return fmt.Errorf("delete comments: %w", err)
		}
	} else {
		// Fetch comments for rejected/needs_revision so the requestor can act on them.
		comments, err = br.store.ListComments(id)
		if err != nil {
			return fmt.Errorf("list comments: %w", err)
		}
	}

	result := VerdictResult{
		ID:       appr.ID,
		Type:     appr.Type,
		EpicID:   appr.EpicID,
		StoryID:  appr.StoryID,
		Content:  appr.Content,
		Verdict:  verdict,
		Comment:  comment,
		Comments: comments,
	}

	br.mu.Lock()
	ch, ok := br.pending[id]
	br.mu.Unlock()

	if ok {
		select {
		case ch <- result:
		default:
		}
	}

	br.broadcast("approval_decided", id)
	return nil
}

// Status returns the current state of the approval identified by id.
func (br *Bridge) Status(id string) (*domain.Approval, error) {
	appr, err := br.store.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("status: %w", err)
	}
	return appr, nil
}

// broadcast sends a JSON event to all connected observers.
func (br *Bridge) broadcast(event, id string) {
	payload, err := json.Marshal(map[string]string{"event": event, "id": id})
	if err != nil {
		return
	}
	br.broadcaster.Broadcast(payload)
}
