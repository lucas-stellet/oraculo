// src/db/task_store.go
package db

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/lucas/oraculo/src/domain"
)

// TaskStore provides CRUD operations and lifecycle transitions for tasks backed by SQLite.
// It includes DAG validation to prevent cyclic dependencies.
type TaskStore struct {
	db *DB
}

// NewTaskStore returns a TaskStore that uses the given database connection.
func NewTaskStore(db *DB) *TaskStore {
	return &TaskStore{db: db}
}

// scanTask reads a Task from a row scanner, parsing SQLite TEXT timestamps.
// started_at and completed_at may be NULL.
func scanTask(row interface{ Scan(...any) error }) (*domain.Task, error) {
	var (
		t                    domain.Task
		createdAt, updatedAt string
		startedAt            *string
		completedAt          *string
	)
	if err := row.Scan(
		&t.ID, &t.StoryID, &t.Name, &t.Description, &t.Status,
		&t.FailureReason, &createdAt, &updatedAt, &startedAt, &completedAt,
	); err != nil {
		return nil, err
	}
	var err error
	if t.CreatedAt, err = time.Parse(sqliteTimeFmt, createdAt); err != nil {
		return nil, fmt.Errorf("parse created_at: %w", err)
	}
	if t.UpdatedAt, err = time.Parse(sqliteTimeFmt, updatedAt); err != nil {
		return nil, fmt.Errorf("parse updated_at: %w", err)
	}
	if startedAt != nil {
		parsed, err := time.Parse(sqliteTimeFmt, *startedAt)
		if err != nil {
			return nil, fmt.Errorf("parse started_at: %w", err)
		}
		t.StartedAt = &parsed
	}
	if completedAt != nil {
		parsed, err := time.Parse(sqliteTimeFmt, *completedAt)
		if err != nil {
			return nil, fmt.Errorf("parse completed_at: %w", err)
		}
		t.CompletedAt = &parsed
	}
	return &t, nil
}

const taskColumns = "id, story_id, name, description, status, failure_reason, created_at, updated_at, started_at, completed_at"

// Create inserts a new task or returns the existing one if (storyID, name) already exists.
// The created flag is true when a new row was inserted.
// If dependsOn is non-empty, dependency edges are created after resolving names to task IDs
// within the same story. Cycle detection is performed before inserting edges.
func (s *TaskStore) Create(storyID int, name, description string, dependsOn []string) (*domain.Task, bool, error) {
	// Use a transaction so that the insert + dependency edges are atomic.
	tx, err := s.db.conn.Begin()
	if err != nil {
		return nil, false, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	res, err := tx.Exec(
		"INSERT OR IGNORE INTO tasks (story_id, name, description) VALUES (?, ?, ?)",
		storyID, name, description,
	)
	if err != nil {
		return nil, false, fmt.Errorf("create task: %w", err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return nil, false, fmt.Errorf("rows affected: %w", err)
	}

	// Fetch the task (whether newly created or existing).
	row := tx.QueryRow(
		"SELECT "+taskColumns+" FROM tasks WHERE story_id = ? AND name = ?",
		storyID, name,
	)
	task, err := scanTask(row)
	if err != nil {
		return nil, false, fmt.Errorf("fetch after create: %w", err)
	}

	// Process dependencies if any.
	if len(dependsOn) > 0 {
		if err := s.addDependencies(tx, storyID, task.ID, name, dependsOn); err != nil {
			return nil, false, err
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, false, fmt.Errorf("commit: %w", err)
	}

	return task, affected > 0, nil
}

// addDependencies resolves dependency names to task IDs and inserts edges,
// performing cycle detection before each insertion.
func (s *TaskStore) addDependencies(tx *sql.Tx, storyID, taskID int, taskName string, dependsOn []string) error {
	for _, depName := range dependsOn {
		// Self-dependency check (also enforced by CHECK constraint).
		if depName == taskName {
			return fmt.Errorf("task %q cannot depend on itself: %w", taskName, domain.ErrCyclicDependency)
		}

		// Resolve dependency name to ID within the same story.
		var depID int
		err := tx.QueryRow(
			"SELECT id FROM tasks WHERE story_id = ? AND name = ?",
			storyID, depName,
		).Scan(&depID)
		if err != nil {
			if err == sql.ErrNoRows {
				return fmt.Errorf("dependency %q: %w", depName, domain.ErrNotFound)
			}
			return fmt.Errorf("resolve dependency %q: %w", depName, err)
		}

		// Cycle detection: check if adding (taskID depends on depID) would create a cycle.
		// A cycle exists if taskID is reachable from depID by following "depends on" edges.
		if err := s.detectCycle(tx, taskID, depID); err != nil {
			return err
		}

		// Insert the dependency edge. Use INSERT OR IGNORE to be idempotent.
		_, err = tx.Exec(
			"INSERT OR IGNORE INTO task_dependencies (task_id, depends_on) VALUES (?, ?)",
			taskID, depID,
		)
		if err != nil {
			return fmt.Errorf("insert dependency %q -> %q: %w", taskName, depName, err)
		}
	}
	return nil
}

// detectCycle checks whether adding an edge (taskID depends on depID) would create a cycle.
// It does BFS from depID following "depends on" edges: for each node X, find all nodes that
// X depends on (SELECT depends_on FROM task_dependencies WHERE task_id = X).
// If we reach taskID, there is a cycle.
func (s *TaskStore) detectCycle(tx *sql.Tx, taskID, depID int) error {
	visited := map[int]bool{}
	queue := []int{depID}

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		if current == taskID {
			return fmt.Errorf("adding dependency would create cycle: %w", domain.ErrCyclicDependency)
		}

		if visited[current] {
			continue
		}
		visited[current] = true

		rows, err := tx.Query(
			"SELECT depends_on FROM task_dependencies WHERE task_id = ?", current,
		)
		if err != nil {
			return fmt.Errorf("query dependencies: %w", err)
		}

		for rows.Next() {
			var next int
			if err := rows.Scan(&next); err != nil {
				rows.Close()
				return fmt.Errorf("scan dependency: %w", err)
			}
			if !visited[next] {
				queue = append(queue, next)
			}
		}
		rows.Close()
		if err := rows.Err(); err != nil {
			return fmt.Errorf("iterate dependencies: %w", err)
		}
	}
	return nil
}

// Start transitions a task from pending to in_progress, setting started_at.
// It returns domain.ErrInvalidTransition if the task is not in pending status.
func (s *TaskStore) Start(storyID int, name string) (*domain.Task, error) {
	task, err := s.GetByName(storyID, name)
	if err != nil {
		return nil, err
	}

	if !task.Status.CanTransitionTo(domain.TaskInProgress) {
		return nil, fmt.Errorf("task %q: cannot transition from %s to in_progress: %w",
			name, task.Status, domain.ErrInvalidTransition)
	}

	_, err = s.db.conn.Exec(
		"UPDATE tasks SET status = ?, started_at = datetime('now'), updated_at = datetime('now') WHERE story_id = ? AND name = ?",
		domain.TaskInProgress, storyID, name,
	)
	if err != nil {
		return nil, fmt.Errorf("start task: %w", err)
	}

	return s.GetByName(storyID, name)
}

// Complete transitions a task from in_progress to completed, setting completed_at
// and inserting a TaskResult row.
// It returns domain.ErrInvalidTransition if the task is not in in_progress status.
func (s *TaskStore) Complete(storyID int, name string, result domain.TaskResult) (*domain.Task, error) {
	task, err := s.GetByName(storyID, name)
	if err != nil {
		return nil, err
	}

	if !task.Status.CanTransitionTo(domain.TaskCompleted) {
		return nil, fmt.Errorf("task %q: cannot transition from %s to completed: %w",
			name, task.Status, domain.ErrInvalidTransition)
	}

	tx, err := s.db.conn.Begin()
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	_, err = tx.Exec(
		"UPDATE tasks SET status = ?, completed_at = datetime('now'), updated_at = datetime('now') WHERE story_id = ? AND name = ?",
		domain.TaskCompleted, storyID, name,
	)
	if err != nil {
		return nil, fmt.Errorf("complete task: %w", err)
	}

	skillsUsed := strings.Join(result.SkillsUsed, ",")
	filesModified := strings.Join(result.FilesModified, ",")

	_, err = tx.Exec(
		"INSERT INTO task_results (task_id, summary, logs, skills_used, files_modified) VALUES (?, ?, ?, ?, ?)",
		task.ID, result.Summary, result.Logs, skillsUsed, filesModified,
	)
	if err != nil {
		return nil, fmt.Errorf("insert task result: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	return s.GetByName(storyID, name)
}

// Fail transitions a task from in_progress to failed, setting failure_reason.
// It returns domain.ErrInvalidTransition if the task is not in in_progress status.
func (s *TaskStore) Fail(storyID int, name, reason string) (*domain.Task, error) {
	task, err := s.GetByName(storyID, name)
	if err != nil {
		return nil, err
	}

	if !task.Status.CanTransitionTo(domain.TaskFailed) {
		return nil, fmt.Errorf("task %q: cannot transition from %s to failed: %w",
			name, task.Status, domain.ErrInvalidTransition)
	}

	_, err = s.db.conn.Exec(
		"UPDATE tasks SET status = ?, failure_reason = ?, updated_at = datetime('now') WHERE story_id = ? AND name = ?",
		domain.TaskFailed, reason, storyID, name,
	)
	if err != nil {
		return nil, fmt.Errorf("fail task: %w", err)
	}

	return s.GetByName(storyID, name)
}

// GetByName retrieves a task by its unique (storyID, name) pair.
// It returns domain.ErrNotFound when no matching task exists.
func (s *TaskStore) GetByName(storyID int, name string) (*domain.Task, error) {
	row := s.db.conn.QueryRow(
		"SELECT "+taskColumns+" FROM tasks WHERE story_id = ? AND name = ?",
		storyID, name,
	)
	task, err := scanTask(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("task %q (story %d): %w", name, storyID, domain.ErrNotFound)
		}
		return nil, fmt.Errorf("get task: %w", err)
	}
	return task, nil
}

// List returns all tasks for the given storyID ordered by creation time.
func (s *TaskStore) List(storyID int) ([]domain.Task, error) {
	rows, err := s.db.conn.Query(
		"SELECT "+taskColumns+" FROM tasks WHERE story_id = ? ORDER BY created_at",
		storyID,
	)
	if err != nil {
		return nil, fmt.Errorf("list tasks: %w", err)
	}
	defer rows.Close()

	var tasks []domain.Task
	for rows.Next() {
		task, err := scanTask(rows)
		if err != nil {
			return nil, fmt.Errorf("scan task: %w", err)
		}
		tasks = append(tasks, *task)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate tasks: %w", err)
	}
	return tasks, nil
}

// Delete removes a task by (storyID, name).
// It returns domain.ErrNotFound when no matching task exists.
func (s *TaskStore) Delete(storyID int, name string) error {
	res, err := s.db.conn.Exec(
		"DELETE FROM tasks WHERE story_id = ? AND name = ?",
		storyID, name,
	)
	if err != nil {
		return fmt.Errorf("delete task: %w", err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}
	if affected == 0 {
		return fmt.Errorf("task %q (story %d): %w", name, storyID, domain.ErrNotFound)
	}
	return nil
}
