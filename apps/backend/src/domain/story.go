// apps/backend/src/domain/story.go
package domain

import "time"

// Story represents a work item scoped to an epic.
type Story struct {
	ID             int
	EpicID         int
	Name           string
	Description    string
	ApprovalStatus ApprovalStatus
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
