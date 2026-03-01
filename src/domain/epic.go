// src/domain/epic.go
package domain

import "time"

// Epic represents a product engineering initiative.
type Epic struct {
	ID             int
	Name           string
	Description    string
	ApprovalStatus ApprovalStatus
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
