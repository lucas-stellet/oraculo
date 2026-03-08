// apps/backend/src/domain/task.go
package domain

import "time"

// TaskStatus represents the lifecycle state of a task.
type TaskStatus string

const (
	TaskPending    TaskStatus = "pending"
	TaskInProgress TaskStatus = "in_progress"
	TaskCompleted  TaskStatus = "completed"
	TaskFailed     TaskStatus = "failed"
)

var validTaskStatuses = map[TaskStatus]bool{
	TaskPending:    true,
	TaskInProgress: true,
	TaskCompleted:  true,
	TaskFailed:     true,
}

// Valid reports whether s is a recognized task status.
func (s TaskStatus) Valid() bool {
	return validTaskStatuses[s]
}

var validTransitions = map[TaskStatus][]TaskStatus{
	TaskPending:    {TaskInProgress},
	TaskInProgress: {TaskCompleted, TaskFailed},
}

// CanTransitionTo reports whether a transition from s to target is valid.
// Valid transitions: pending -> in_progress -> completed | failed.
func (s TaskStatus) CanTransitionTo(target TaskStatus) bool {
	for _, allowed := range validTransitions[s] {
		if allowed == target {
			return true
		}
	}
	return false
}

// Task represents an executable unit of work within a story.
type Task struct {
	ID            int
	StoryID       int
	Name          string
	Description   string
	Status        TaskStatus
	FailureReason string
	CreatedAt     time.Time
	UpdatedAt     time.Time
	StartedAt     *time.Time
	CompletedAt   *time.Time
}

// TaskResult holds rich completion data for a finished task.
type TaskResult struct {
	ID            int
	TaskID        int
	Summary       string
	Logs          string
	SkillsUsed    []string
	FilesModified []string
	CreatedAt     time.Time
}
