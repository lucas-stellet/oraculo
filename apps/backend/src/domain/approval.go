// apps/backend/src/domain/approval.go
package domain

import "time"

// ApprovalType identifies the kind of approval gate.
type ApprovalType string

const (
	ApprovalQAEscalation  ApprovalType = "qa-escalation"
	ApprovalExecutionPlan ApprovalType = "execution-plan"
	ApprovalDesign        ApprovalType = "design"
)

var validApprovalTypes = map[ApprovalType]bool{
	ApprovalQAEscalation: true, ApprovalExecutionPlan: true,
	ApprovalDesign: true,
}

// Valid reports whether a is a recognized approval type.
func (a ApprovalType) Valid() bool { return validApprovalTypes[a] }

// Verdict represents a human decision on an approval request.
type Verdict string

const (
	VerdictApproved      Verdict = "approved"
	VerdictRejected      Verdict = "rejected"
	VerdictNeedsRevision Verdict = "needs_revision"
)

var validVerdicts = map[Verdict]bool{
	VerdictApproved: true, VerdictRejected: true, VerdictNeedsRevision: true,
}

// Valid reports whether v is a recognized verdict.
func (v Verdict) Valid() bool { return validVerdicts[v] }

// Approval represents a human-in-the-loop approval gate request.
type Approval struct {
	ID              string         `json:"id"`
	Type            ApprovalType   `json:"type"`
	EpicID          *int           `json:"epic_id,omitempty"`
	StoryID         *int           `json:"story_id,omitempty"`
	Content         string         `json:"content"`
	PreviousVersion string         `json:"previous_version"`
	Status          ApprovalStatus `json:"status"`
	VerdictComment  string         `json:"verdict_comment"`
	RequestedAt     time.Time      `json:"requested_at"`
	DecidedAt       *time.Time     `json:"decided_at,omitempty"`
}
