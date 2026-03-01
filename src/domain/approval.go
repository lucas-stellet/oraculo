// src/domain/approval.go
package domain

import "time"

// ApprovalType identifies the kind of approval gate.
type ApprovalType string

const (
	ApprovalEpicRequirements ApprovalType = "epic-requirements"
	ApprovalStoryDefinition  ApprovalType = "story-definition"
	ApprovalQAEscalation     ApprovalType = "qa-escalation"
	ApprovalExecutionPlan    ApprovalType = "execution-plan"
)

var validApprovalTypes = map[ApprovalType]bool{
	ApprovalEpicRequirements: true, ApprovalStoryDefinition: true,
	ApprovalQAEscalation: true, ApprovalExecutionPlan: true,
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
	ID              string
	Type            ApprovalType
	EpicID          *int
	StoryID         *int
	Content         string
	PreviousVersion string
	Status          ApprovalStatus
	VerdictComment  string
	RequestedAt     time.Time
	DecidedAt       *time.Time
}
