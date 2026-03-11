// Package mcp provides the Oraculo MCP server, which exposes approval-gate
// tools so that Claude Code agents can request and poll human approvals.
package mcp

import (
	"context"
	"fmt"
	"log/slog"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/lucas/oraculo/apps/backend/src/approval"
	"github.com/lucas/oraculo/apps/backend/src/db"
	"github.com/lucas/oraculo/apps/backend/src/domain"
)

// Server wraps the MCP SDK server and wires it to the Oraculo approval system.
type Server struct {
	inner  *sdk.Server
	bridge *approval.Bridge
	store  *db.ApprovalStore
	logger *slog.Logger
}

// requestApprovalInput is the typed input for the request_approval tool.
type requestApprovalInput struct {
	// Type is the approval gate kind (e.g. "qa-escalation").
	Type string `json:"type" jsonschema:"Approval type (qa-escalation|execution-plan|design)"`
	// Content is the document or artifact awaiting human review.
	Content string `json:"content" jsonschema:"The content to be approved"`
	// EpicID is an optional numeric epic ID.
	EpicID *int `json:"epic_id,omitempty" jsonschema:"Optional epic ID"`
	// StoryID is an optional numeric story ID.
	StoryID *int `json:"story_id,omitempty" jsonschema:"Optional story ID"`
}

// requestApprovalOutput is the typed output returned once a verdict is recorded.
type requestApprovalOutput struct {
	ID       string                   `json:"id"`
	Type     string                   `json:"type"`
	EpicID   *int                     `json:"epic_id,omitempty"`
	StoryID  *int                     `json:"story_id,omitempty"`
	Content  string                   `json:"content"`
	Status   string                   `json:"status"`
	Verdict  string                   `json:"verdict,omitempty"`
	Comment  string                   `json:"comment,omitempty"`
	Comments []domain.ApprovalComment `json:"comments"`
}

// approvalStatusInput is the typed input for the approval_status tool.
type approvalStatusInput struct {
	// ID is the UUID of the approval to check.
	ID string `json:"id" jsonschema:"The approval ID to check"`
}

// approvalStatusOutput mirrors the current state of an approval.
type approvalStatusOutput struct {
	ID       string                   `json:"id"`
	Type     string                   `json:"type"`
	EpicID   *int                     `json:"epic_id,omitempty"`
	StoryID  *int                     `json:"story_id,omitempty"`
	Content  string                   `json:"content"`
	Status   string                   `json:"status"`
	Comment  string                   `json:"comment,omitempty"`
	Comments []domain.ApprovalComment `json:"comments"`
}

// New constructs and wires an MCP server with the request_approval and
// approval_status tools registered.
func New(bridge *approval.Bridge, store *db.ApprovalStore, logger *slog.Logger) *Server {
	inner := sdk.NewServer(
		&sdk.Implementation{Name: "oraculo", Version: "1.0.0"},
		nil,
	)

	s := &Server{
		inner:  inner,
		bridge: bridge,
		store:  store,
		logger: logger,
	}

	sdk.AddTool(inner,
		&sdk.Tool{
			Name:        "request_approval",
			Description: "Submit an artifact for human approval. Blocks until a verdict (approved, rejected, or needs_revision) is recorded. Use for operational gates only (qa-escalation, execution-plan, design).",
		},
		s.handleRequestApproval,
	)

	sdk.AddTool(inner,
		&sdk.Tool{
			Name:        "approval_status",
			Description: "Get the current status of an operational approval request.",
		},
		s.handleApprovalStatus,
	)

	return s
}

// Inner returns the underlying SDK server, for use with transports.
func (s *Server) Inner() *sdk.Server {
	return s.inner
}

// RunStdio connects the server to stdin/stdout and serves until the connection
// closes. It blocks until the transport is closed.
func (s *Server) RunStdio(ctx context.Context) error {
	_, err := s.inner.Connect(ctx, &sdk.StdioTransport{}, nil)
	if err != nil {
		return fmt.Errorf("mcp stdio connect: %w", err)
	}
	s.logger.Info("mcp.connected")
	<-ctx.Done()
	return ctx.Err()
}

// handleRequestApproval is the handler for the request_approval tool.
// It creates an approval gate and blocks until a verdict is recorded.
func (s *Server) handleRequestApproval(
	ctx context.Context,
	_ *sdk.CallToolRequest,
	in requestApprovalInput,
) (*sdk.CallToolResult, requestApprovalOutput, error) {
	at := domain.ApprovalType(in.Type)
	if !at.Valid() {
		return nil, requestApprovalOutput{}, fmt.Errorf("invalid approval type %q", in.Type)
	}
	if in.Content == "" {
		return nil, requestApprovalOutput{}, fmt.Errorf("content is required")
	}

	req := approval.ApprovalRequest{
		Type:    at,
		EpicID:  in.EpicID,
		StoryID: in.StoryID,
		Content: in.Content,
	}

	result, err := s.bridge.Request(ctx, req)
	if err != nil {
		return nil, requestApprovalOutput{}, fmt.Errorf("approval request: %w", err)
	}

	out := requestApprovalOutput{
		ID:       result.ID,
		Type:     string(result.Type),
		EpicID:   result.EpicID,
		StoryID:  result.StoryID,
		Content:  result.Content,
		Verdict:  string(result.Verdict),
		Comment:  result.Comment,
		Comments: result.Comments,
	}
	return nil, out, nil
}

// handleApprovalStatus is the handler for the approval_status tool.
// It returns the current state of the requested approval without blocking.
func (s *Server) handleApprovalStatus(
	ctx context.Context,
	_ *sdk.CallToolRequest,
	in approvalStatusInput,
) (*sdk.CallToolResult, approvalStatusOutput, error) {
	_ = ctx
	if in.ID == "" {
		return nil, approvalStatusOutput{}, fmt.Errorf("id is required")
	}

	a, err := s.bridge.Status(in.ID)
	if err != nil {
		return nil, approvalStatusOutput{}, err
	}

	comments, err := s.store.ListComments(in.ID)
	if err != nil {
		return nil, approvalStatusOutput{}, fmt.Errorf("list comments: %w", err)
	}

	out := approvalStatusOutput{
		ID:       a.ID,
		Type:     string(a.Type),
		EpicID:   a.EpicID,
		StoryID:  a.StoryID,
		Content:  a.Content,
		Status:   string(a.Status),
		Comment:  a.VerdictComment,
		Comments: comments,
	}
	return nil, out, nil
}
