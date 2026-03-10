// apps/backend/src/domain/validation.go
package domain

import "time"

// Validation represents a QA validation verdict for a story.
type Validation struct {
	ID        int       `json:"id"`
	StoryID   int       `json:"story_id"`
	TaskID    *int      `json:"task_id,omitempty"`
	Verdict   string    `json:"verdict"`
	Findings  string    `json:"findings"`
	CreatedAt time.Time `json:"created_at"`
}
