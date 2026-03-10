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
	ID            int        `json:"id"`
	StoryID       int        `json:"story_id"`
	Name          string     `json:"name"`
	Description   string     `json:"description"`
	Status        TaskStatus `json:"status"`
	FailureReason string     `json:"failure_reason"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	StartedAt     *time.Time `json:"started_at,omitempty"`
	CompletedAt   *time.Time `json:"completed_at,omitempty"`
}

// TaskResult holds rich completion data for a finished task.
type TaskResult struct {
	ID            int       `json:"id"`
	TaskID        int       `json:"task_id"`
	Summary       string    `json:"summary"`
	Logs          string    `json:"logs"`
	SkillsUsed    []string  `json:"skills_used"`
	FilesModified []string  `json:"files_modified"`
	CreatedAt     time.Time `json:"created_at"`
}

// TaskEnriched extends Task with dependencies and result for API responses.
type TaskEnriched struct {
	Task
	DependsOn []int       `json:"depends_on"`
	Result    *TaskResult `json:"result,omitempty"`
}
