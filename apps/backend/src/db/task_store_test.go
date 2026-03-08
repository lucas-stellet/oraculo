package db

import (
	"errors"
	"testing"

	"github.com/lucas/oraculo/apps/backend/src/domain"
)

// createStory is a test helper that creates an epic and a story, returning storyID.
func createStory(t *testing.T, database *DB, epicName, storyName string) int {
	t.Helper()
	epicID := createEpic(t, database, epicName)
	story, _, err := NewStoryStore(database).Create(epicID, storyName, "test story")
	if err != nil {
		t.Fatalf("createStory(%q, %q): %v", epicName, storyName, err)
	}
	return story.ID
}

// --- CRUD tests ---

func TestTaskStore_Create(t *testing.T) {
	database := testDB(t)
	storyID := createStory(t, database, "epic", "story")
	store := NewTaskStore(database)

	task, created, err := store.Create(storyID, "build-api", "build the API", nil)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if !created {
		t.Error("expected created=true on first call")
	}
	if task.Name != "build-api" {
		t.Errorf("Name = %q, want %q", task.Name, "build-api")
	}
	if task.Description != "build the API" {
		t.Errorf("Description = %q, want %q", task.Description, "build the API")
	}
	if task.StoryID != storyID {
		t.Errorf("StoryID = %d, want %d", task.StoryID, storyID)
	}
	if task.Status != domain.TaskPending {
		t.Errorf("Status = %q, want %q", task.Status, domain.TaskPending)
	}
}

func TestTaskStore_CreateIdempotent(t *testing.T) {
	database := testDB(t)
	storyID := createStory(t, database, "epic", "story")
	store := NewTaskStore(database)

	t1, _, _ := store.Create(storyID, "build-api", "desc", nil)
	t2, created, err := store.Create(storyID, "build-api", "desc", nil)
	if err != nil {
		t.Fatalf("second Create: %v", err)
	}
	if created {
		t.Error("expected created=false on second call")
	}
	if t2.ID != t1.ID {
		t.Errorf("IDs differ: %d vs %d", t2.ID, t1.ID)
	}
}

func TestTaskStore_GetByName(t *testing.T) {
	database := testDB(t)
	storyID := createStory(t, database, "epic", "story")
	store := NewTaskStore(database)

	store.Create(storyID, "build-api", "desc", nil)
	task, err := store.GetByName(storyID, "build-api")
	if err != nil {
		t.Fatalf("GetByName: %v", err)
	}
	if task.Name != "build-api" {
		t.Errorf("Name = %q, want %q", task.Name, "build-api")
	}
	if task.Status != domain.TaskPending {
		t.Errorf("Status = %q, want %q", task.Status, domain.TaskPending)
	}
}

func TestTaskStore_GetByName_NotFound(t *testing.T) {
	database := testDB(t)
	storyID := createStory(t, database, "epic", "story")
	store := NewTaskStore(database)

	_, err := store.GetByName(storyID, "nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent task")
	}
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got: %v", err)
	}
}

func TestTaskStore_List(t *testing.T) {
	database := testDB(t)
	storyID := createStory(t, database, "epic", "story")
	store := NewTaskStore(database)

	store.Create(storyID, "alpha", "", nil)
	store.Create(storyID, "beta", "", nil)

	tasks, err := store.List(storyID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(tasks) != 2 {
		t.Fatalf("len = %d, want 2", len(tasks))
	}
	if tasks[0].Name != "alpha" {
		t.Errorf("tasks[0].Name = %q, want %q", tasks[0].Name, "alpha")
	}
	if tasks[1].Name != "beta" {
		t.Errorf("tasks[1].Name = %q, want %q", tasks[1].Name, "beta")
	}
}

// --- Lifecycle transition tests ---

func TestTaskStore_Start(t *testing.T) {
	database := testDB(t)
	storyID := createStory(t, database, "epic", "story")
	store := NewTaskStore(database)

	store.Create(storyID, "task-a", "desc", nil)
	task, err := store.Start(storyID, "task-a")
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	if task.Status != domain.TaskInProgress {
		t.Errorf("Status = %q, want %q", task.Status, domain.TaskInProgress)
	}
	if task.StartedAt == nil {
		t.Error("expected StartedAt to be set")
	}
}

func TestTaskStore_Start_NonPending(t *testing.T) {
	database := testDB(t)
	storyID := createStory(t, database, "epic", "story")
	store := NewTaskStore(database)

	store.Create(storyID, "task-a", "desc", nil)
	store.Start(storyID, "task-a")

	// try to start again — already in_progress
	_, err := store.Start(storyID, "task-a")
	if err == nil {
		t.Fatal("expected error starting non-pending task")
	}
	if !errors.Is(err, domain.ErrInvalidTransition) {
		t.Errorf("expected ErrInvalidTransition, got: %v", err)
	}
}

func TestTaskStore_Complete(t *testing.T) {
	database := testDB(t)
	storyID := createStory(t, database, "epic", "story")
	store := NewTaskStore(database)

	store.Create(storyID, "task-a", "desc", nil)
	store.Start(storyID, "task-a")

	result := domain.TaskResult{
		Summary: "all tests pass",
		Logs:    "some log output",
	}
	task, err := store.Complete(storyID, "task-a", result)
	if err != nil {
		t.Fatalf("Complete: %v", err)
	}
	if task.Status != domain.TaskCompleted {
		t.Errorf("Status = %q, want %q", task.Status, domain.TaskCompleted)
	}
	if task.CompletedAt == nil {
		t.Error("expected CompletedAt to be set")
	}
}

func TestTaskStore_Complete_NonInProgress(t *testing.T) {
	database := testDB(t)
	storyID := createStory(t, database, "epic", "story")
	store := NewTaskStore(database)

	store.Create(storyID, "task-a", "desc", nil)

	// try to complete a pending task
	result := domain.TaskResult{Summary: "done"}
	_, err := store.Complete(storyID, "task-a", result)
	if err == nil {
		t.Fatal("expected error completing non-in_progress task")
	}
	if !errors.Is(err, domain.ErrInvalidTransition) {
		t.Errorf("expected ErrInvalidTransition, got: %v", err)
	}
}

func TestTaskStore_Fail(t *testing.T) {
	database := testDB(t)
	storyID := createStory(t, database, "epic", "story")
	store := NewTaskStore(database)

	store.Create(storyID, "task-a", "desc", nil)
	store.Start(storyID, "task-a")

	task, err := store.Fail(storyID, "task-a", "compilation error")
	if err != nil {
		t.Fatalf("Fail: %v", err)
	}
	if task.Status != domain.TaskFailed {
		t.Errorf("Status = %q, want %q", task.Status, domain.TaskFailed)
	}
	if task.FailureReason != "compilation error" {
		t.Errorf("FailureReason = %q, want %q", task.FailureReason, "compilation error")
	}
}

func TestTaskStore_Fail_NonInProgress(t *testing.T) {
	database := testDB(t)
	storyID := createStory(t, database, "epic", "story")
	store := NewTaskStore(database)

	store.Create(storyID, "task-a", "desc", nil)

	// try to fail a pending task
	_, err := store.Fail(storyID, "task-a", "some reason")
	if err == nil {
		t.Fatal("expected error failing non-in_progress task")
	}
	if !errors.Is(err, domain.ErrInvalidTransition) {
		t.Errorf("expected ErrInvalidTransition, got: %v", err)
	}
}

// --- DAG (dependency) tests ---

func TestTaskStore_Create_WithDependencies(t *testing.T) {
	database := testDB(t)
	storyID := createStory(t, database, "epic", "story")
	store := NewTaskStore(database)

	store.Create(storyID, "setup-db", "create schema", nil)
	store.Create(storyID, "seed-data", "insert seed data", nil)

	// task-c depends on both setup-db and seed-data
	task, created, err := store.Create(storyID, "run-tests", "execute tests", []string{"setup-db", "seed-data"})
	if err != nil {
		t.Fatalf("Create with deps: %v", err)
	}
	if !created {
		t.Error("expected created=true")
	}
	if task.Name != "run-tests" {
		t.Errorf("Name = %q, want %q", task.Name, "run-tests")
	}

	// verify the dependency edges exist by querying directly
	var count int
	err = database.conn.QueryRow(
		"SELECT COUNT(*) FROM task_dependencies WHERE task_id = ?", task.ID,
	).Scan(&count)
	if err != nil {
		t.Fatalf("count dependencies: %v", err)
	}
	if count != 2 {
		t.Errorf("dependency count = %d, want 2", count)
	}
}

func TestTaskStore_Create_SelfDependency(t *testing.T) {
	database := testDB(t)
	storyID := createStory(t, database, "epic", "story")
	store := NewTaskStore(database)

	// self-dependency: task depends on itself
	_, _, err := store.Create(storyID, "task-a", "desc", []string{"task-a"})
	if err == nil {
		t.Fatal("expected error for self-dependency")
	}
}

func TestTaskStore_Create_CyclicDependency_ThreeNodes(t *testing.T) {
	database := testDB(t)
	storyID := createStory(t, database, "epic", "story")
	store := NewTaskStore(database)

	// Create A (no deps)
	store.Create(storyID, "task-a", "desc", nil)
	// Create B depends on A
	store.Create(storyID, "task-b", "desc", []string{"task-a"})
	// Create C depends on B
	store.Create(storyID, "task-c", "desc", []string{"task-b"})

	// Now try to re-create A with dependency on C.
	// This would create cycle: A depends on C, C depends on B, B depends on A.
	_, _, err := store.Create(storyID, "task-a", "desc", []string{"task-c"})
	if err == nil {
		t.Fatal("expected error for cyclic dependency")
	}
	if !errors.Is(err, domain.ErrCyclicDependency) {
		t.Errorf("expected ErrCyclicDependency, got: %v", err)
	}
}

func TestTaskStore_Create_CyclicDependency_TwoNodes(t *testing.T) {
	database := testDB(t)
	storyID := createStory(t, database, "epic", "story")
	store := NewTaskStore(database)

	// Create A (no deps)
	store.Create(storyID, "task-a", "desc", nil)
	// Create B depends on A
	store.Create(storyID, "task-b", "desc", []string{"task-a"})

	// Now try to re-create A with dependency on B.
	// This would create cycle: A depends on B, B depends on A.
	_, _, err := store.Create(storyID, "task-a", "desc", []string{"task-b"})
	if err == nil {
		t.Fatal("expected error for cyclic dependency")
	}
	if !errors.Is(err, domain.ErrCyclicDependency) {
		t.Errorf("expected ErrCyclicDependency, got: %v", err)
	}
}

// --- Delete tests ---

func TestTaskStore_Delete(t *testing.T) {
	database := testDB(t)
	storyID := createStory(t, database, "epic", "story")
	store := NewTaskStore(database)

	store.Create(storyID, "doomed", "", nil)
	if err := store.Delete(storyID, "doomed"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	_, err := store.GetByName(storyID, "doomed")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Error("expected ErrNotFound after delete")
	}
}

func TestTaskStore_Delete_NotFound(t *testing.T) {
	database := testDB(t)
	storyID := createStory(t, database, "epic", "story")
	store := NewTaskStore(database)

	err := store.Delete(storyID, "nonexistent")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got: %v", err)
	}
}
