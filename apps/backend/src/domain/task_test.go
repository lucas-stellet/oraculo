// apps/backend/src/domain/task_test.go
package domain

import "testing"

func TestTaskStatus_CanTransitionTo(t *testing.T) {
	tests := []struct {
		from TaskStatus
		to   TaskStatus
		want bool
	}{
		{TaskPending, TaskInProgress, true},
		{TaskPending, TaskCompleted, false},
		{TaskPending, TaskFailed, false},
		{TaskPending, TaskPending, false},
		{TaskInProgress, TaskCompleted, true},
		{TaskInProgress, TaskFailed, true},
		{TaskInProgress, TaskPending, false},
		{TaskInProgress, TaskInProgress, false},
		{TaskCompleted, TaskPending, false},
		{TaskCompleted, TaskInProgress, false},
		{TaskCompleted, TaskFailed, false},
		{TaskFailed, TaskPending, false},
		{TaskFailed, TaskInProgress, false},
		{TaskFailed, TaskCompleted, false},
	}
	for _, tt := range tests {
		got := tt.from.CanTransitionTo(tt.to)
		if got != tt.want {
			t.Errorf("%s → %s: got %v, want %v", tt.from, tt.to, got, tt.want)
		}
	}
}

func TestTaskStatus_Valid(t *testing.T) {
	tests := []struct {
		s    TaskStatus
		want bool
	}{
		{TaskPending, true},
		{TaskInProgress, true},
		{TaskCompleted, true},
		{TaskFailed, true},
		{TaskStatus("unknown"), false},
	}
	for _, tt := range tests {
		if got := tt.s.Valid(); got != tt.want {
			t.Errorf("TaskStatus(%q).Valid() = %v, want %v", tt.s, got, tt.want)
		}
	}
}
