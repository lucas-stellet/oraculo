// src/domain/validation.go
package domain

import "time"

// Validation represents a QA validation verdict for a story.
type Validation struct {
	ID        int
	StoryID   int
	TaskID    *int
	Verdict   string
	Findings  string
	CreatedAt time.Time
}
