// src/domain/session.go
package domain

import "time"

// SessionType identifies which command owns a session.
type SessionType string

const (
	SessionEpic     SessionType = "epic"
	SessionStory    SessionType = "story"
	SessionPlan     SessionType = "plan"
	SessionExecute  SessionType = "execute"
	SessionValidate SessionType = "validate"
)

var validSessionTypes = map[SessionType]bool{
	SessionEpic:     true,
	SessionStory:    true,
	SessionPlan:     true,
	SessionExecute:  true,
	SessionValidate: true,
}

// Valid reports whether s is a recognized session type.
func (s SessionType) Valid() bool {
	return validSessionTypes[s]
}

// SessionStatus represents the lifecycle state of a session.
type SessionStatus string

const (
	SessionActive    SessionStatus = "active"
	SessionCompleted SessionStatus = "completed"
	SessionAbandoned SessionStatus = "abandoned"
)

var validSessionStatuses = map[SessionStatus]bool{
	SessionActive:    true,
	SessionCompleted: true,
	SessionAbandoned: true,
}

// Valid reports whether s is a recognized session status.
func (s SessionStatus) Valid() bool {
	return validSessionStatuses[s]
}

// Session tracks a skill invocation's progress through its phases.
type Session struct {
	ID        string
	Type      SessionType
	EpicID    *int
	Status    SessionStatus
	CreatedAt time.Time
	ClosedAt  *time.Time
}

// Phase records one completed phase within a session.
type Phase struct {
	SessionID   string
	Name        string
	Data        string // JSON blob, skill-defined
	CompletedAt time.Time
}
