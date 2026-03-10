// apps/backend/src/domain/story.go
package domain

import "time"

// Story represents a work item scoped to an epic.
type Story struct {
	ID             int            `json:"id"`
	EpicID         int            `json:"epic_id"`
	Name           string         `json:"name"`
	Description    string         `json:"description"`
	ApprovalStatus ApprovalStatus `json:"approval_status"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
}

// StorySummary extends Story with aggregated task data for dashboard display.
type StorySummary struct {
	Story
	TaskCount          int `json:"task_count"`
	CompletedTaskCount int `json:"completed_task_count"`
	FailedTaskCount    int `json:"failed_task_count"`
}
