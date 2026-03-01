// src/domain/status.go
package domain

// ApprovalStatus represents the approval state of an epic or story.
type ApprovalStatus string

const (
	ApprovalNone          ApprovalStatus = "none"
	ApprovalPending       ApprovalStatus = "pending"
	ApprovalApproved      ApprovalStatus = "approved"
	ApprovalRejected      ApprovalStatus = "rejected"
	ApprovalNeedsRevision ApprovalStatus = "needs_revision"
)

var validApprovalStatuses = map[ApprovalStatus]bool{
	ApprovalNone:          true,
	ApprovalPending:       true,
	ApprovalApproved:      true,
	ApprovalRejected:      true,
	ApprovalNeedsRevision: true,
}

// Valid reports whether s is a recognized approval status.
func (s ApprovalStatus) Valid() bool {
	return validApprovalStatuses[s]
}
