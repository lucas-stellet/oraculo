// src/domain/epic.go
package domain

import "time"

// Epic represents a product engineering initiative.
type Epic struct {
	ID             int            `json:"id"`
	Name           string         `json:"name"`
	Description    string         `json:"description"`
	ApprovalStatus ApprovalStatus `json:"approval_status"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
}

// EpicSummary extends Epic with aggregated data for dashboard display.
type EpicSummary struct {
	Epic
	Phase              string `json:"phase"`
	PhaseStatus        string `json:"phase_status"`
	StoryCount         int    `json:"story_count"`
	TaskCount          int    `json:"task_count"`
	CompletedTaskCount int    `json:"completed_task_count"`
}

// SessionTypeToPhase maps session types to dashboard phase labels.
var SessionTypeToPhase = map[SessionType]string{
	SessionEpic:     "discover",
	SessionStory:    "discover",
	SessionPlan:     "plan",
	SessionExecute:  "execute",
	SessionValidate: "validate",
}
