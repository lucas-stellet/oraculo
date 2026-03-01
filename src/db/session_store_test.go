package db

import (
	"errors"
	"testing"

	"github.com/lucas/oraculo/src/domain"
)

func createEpicForTest(t *testing.T, database *DB, name string) int {
	t.Helper()
	store := NewEpicStore(database)
	epic, _, err := store.Create(name, "")
	if err != nil {
		t.Fatalf("create epic: %v", err)
	}
	return epic.ID
}

func TestSessionStore_Create(t *testing.T) {
	database := testDB(t)
	epicID := createEpicForTest(t, database, "test-epic")
	store := NewSessionStore(database)

	session, created, err := store.Create(domain.SessionEpic, &epicID)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if !created {
		t.Error("expected created=true on first call")
	}
	if session.Type != domain.SessionEpic {
		t.Errorf("Type = %q, want %q", session.Type, domain.SessionEpic)
	}
	if session.Status != domain.SessionActive {
		t.Errorf("Status = %q, want %q", session.Status, domain.SessionActive)
	}
	if session.ID == "" {
		t.Error("expected non-empty ID")
	}
}

func TestSessionStore_Create_Idempotent(t *testing.T) {
	database := testDB(t)
	epicID := createEpicForTest(t, database, "test-epic")
	store := NewSessionStore(database)

	s1, _, _ := store.Create(domain.SessionEpic, &epicID)
	s2, created, err := store.Create(domain.SessionEpic, &epicID)
	if err != nil {
		t.Fatalf("second Create: %v", err)
	}
	if created {
		t.Error("expected created=false on second call")
	}
	if s2.ID != s1.ID {
		t.Errorf("IDs differ: %q vs %q", s2.ID, s1.ID)
	}
}

func TestSessionStore_Get(t *testing.T) {
	database := testDB(t)
	epicID := createEpicForTest(t, database, "test-epic")
	store := NewSessionStore(database)

	created, _, _ := store.Create(domain.SessionEpic, &epicID)
	got, err := store.Get(created.ID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.ID != created.ID {
		t.Errorf("ID = %q, want %q", got.ID, created.ID)
	}
}

func TestSessionStore_Get_NotFound(t *testing.T) {
	database := testDB(t)
	store := NewSessionStore(database)

	_, err := store.Get("nonexistent")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got: %v", err)
	}
}

func TestSessionStore_ActiveByEpic(t *testing.T) {
	database := testDB(t)
	epicID := createEpicForTest(t, database, "test-epic")
	store := NewSessionStore(database)

	created, _, _ := store.Create(domain.SessionEpic, &epicID)
	got, err := store.ActiveByEpic(epicID, domain.SessionEpic)
	if err != nil {
		t.Fatalf("ActiveByEpic: %v", err)
	}
	if got.ID != created.ID {
		t.Errorf("ID = %q, want %q", got.ID, created.ID)
	}
}

func TestSessionStore_ActiveByEpic_NotFound(t *testing.T) {
	database := testDB(t)
	epicID := createEpicForTest(t, database, "test-epic")
	store := NewSessionStore(database)

	_, err := store.ActiveByEpic(epicID, domain.SessionEpic)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got: %v", err)
	}
}

func TestSessionStore_Close(t *testing.T) {
	database := testDB(t)
	epicID := createEpicForTest(t, database, "test-epic")
	store := NewSessionStore(database)

	session, _, _ := store.Create(domain.SessionEpic, &epicID)
	if err := store.Close(session.ID, domain.SessionCompleted); err != nil {
		t.Fatalf("Close: %v", err)
	}
	got, err := store.Get(session.ID)
	if err != nil {
		t.Fatalf("Get after close: %v", err)
	}
	if got.Status != domain.SessionCompleted {
		t.Errorf("Status = %q, want %q", got.Status, domain.SessionCompleted)
	}
	if got.ClosedAt == nil {
		t.Error("expected non-nil ClosedAt")
	}
}

func TestSessionStore_Close_Abandoned(t *testing.T) {
	database := testDB(t)
	epicID := createEpicForTest(t, database, "test-epic")
	store := NewSessionStore(database)

	session, _, _ := store.Create(domain.SessionEpic, &epicID)
	if err := store.Close(session.ID, domain.SessionAbandoned); err != nil {
		t.Fatalf("Close: %v", err)
	}
	got, _ := store.Get(session.ID)
	if got.Status != domain.SessionAbandoned {
		t.Errorf("Status = %q, want %q", got.Status, domain.SessionAbandoned)
	}
}

func TestSessionStore_Close_AlreadyClosed(t *testing.T) {
	database := testDB(t)
	epicID := createEpicForTest(t, database, "test-epic")
	store := NewSessionStore(database)

	session, _, _ := store.Create(domain.SessionEpic, &epicID)
	store.Close(session.ID, domain.SessionCompleted)

	err := store.Close(session.ID, domain.SessionCompleted)
	if !errors.Is(err, domain.ErrInvalidTransition) {
		t.Errorf("expected ErrInvalidTransition, got: %v", err)
	}
}

func TestSessionStore_Close_NotFound(t *testing.T) {
	database := testDB(t)
	store := NewSessionStore(database)

	err := store.Close("nonexistent", domain.SessionCompleted)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got: %v", err)
	}
}

func TestSessionStore_CompletePhase(t *testing.T) {
	database := testDB(t)
	epicID := createEpicForTest(t, database, "test-epic")
	store := NewSessionStore(database)

	session, _, _ := store.Create(domain.SessionEpic, &epicID)
	err := store.CompletePhase(session.ID, "setup", `{"reasoning_level":"deep"}`)
	if err != nil {
		t.Fatalf("CompletePhase: %v", err)
	}
}

func TestSessionStore_CompletePhase_OutOfOrder(t *testing.T) {
	database := testDB(t)
	epicID := createEpicForTest(t, database, "test-epic")
	store := NewSessionStore(database)

	session, _, _ := store.Create(domain.SessionEpic, &epicID)
	// Skip setup, try to complete reframing
	err := store.CompletePhase(session.ID, "reframing", "{}")
	if !errors.Is(err, domain.ErrInvalidTransition) {
		t.Errorf("expected ErrInvalidTransition, got: %v", err)
	}
}

func TestSessionStore_CompletePhase_Duplicate(t *testing.T) {
	database := testDB(t)
	epicID := createEpicForTest(t, database, "test-epic")
	store := NewSessionStore(database)

	session, _, _ := store.Create(domain.SessionEpic, &epicID)
	store.CompletePhase(session.ID, "setup", "{}")
	err := store.CompletePhase(session.ID, "setup", "{}")
	if !errors.Is(err, domain.ErrAlreadyExists) {
		t.Errorf("expected ErrAlreadyExists, got: %v", err)
	}
}

func TestSessionStore_CompletePhase_UnknownPhase(t *testing.T) {
	database := testDB(t)
	epicID := createEpicForTest(t, database, "test-epic")
	store := NewSessionStore(database)

	session, _, _ := store.Create(domain.SessionEpic, &epicID)
	err := store.CompletePhase(session.ID, "nonexistent", "{}")
	if !errors.Is(err, domain.ErrInvalidPhase) {
		t.Errorf("expected ErrInvalidPhase, got: %v", err)
	}
}

func TestSessionStore_CompletePhase_AutoClose(t *testing.T) {
	database := testDB(t)
	epicID := createEpicForTest(t, database, "test-epic")
	store := NewSessionStore(database)

	session, _, _ := store.Create(domain.SessionValidate, &epicID) // 3 phases: setup, qa-dispatch, verdict
	store.CompletePhase(session.ID, "setup", "{}")
	store.CompletePhase(session.ID, "qa-dispatch", "{}")
	err := store.CompletePhase(session.ID, "verdict", "{}")
	if err != nil {
		t.Fatalf("CompletePhase last: %v", err)
	}
	got, _ := store.Get(session.ID)
	if got.Status != domain.SessionCompleted {
		t.Errorf("Status = %q, want %q after last phase", got.Status, domain.SessionCompleted)
	}
}

func TestSessionStore_CompletePhase_SessionNotFound(t *testing.T) {
	database := testDB(t)
	store := NewSessionStore(database)

	err := store.CompletePhase("nonexistent", "setup", "{}")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got: %v", err)
	}
}

func TestSessionStore_Phases(t *testing.T) {
	database := testDB(t)
	epicID := createEpicForTest(t, database, "test-epic")
	store := NewSessionStore(database)

	session, _, _ := store.Create(domain.SessionEpic, &epicID)
	store.CompletePhase(session.ID, "setup", `{"level":"deep"}`)
	store.CompletePhase(session.ID, "reframing", `{"problem":"test"}`)

	phases, err := store.Phases(session.ID)
	if err != nil {
		t.Fatalf("Phases: %v", err)
	}
	if len(phases) != 2 {
		t.Fatalf("len = %d, want 2", len(phases))
	}
	if phases[0].Name != "setup" {
		t.Errorf("phases[0].Name = %q, want %q", phases[0].Name, "setup")
	}
	if phases[1].Name != "reframing" {
		t.Errorf("phases[1].Name = %q, want %q", phases[1].Name, "reframing")
	}
	if phases[0].Data != `{"level":"deep"}` {
		t.Errorf("phases[0].Data = %q, want %q", phases[0].Data, `{"level":"deep"}`)
	}
}

func TestSessionStore_Phases_Empty(t *testing.T) {
	database := testDB(t)
	epicID := createEpicForTest(t, database, "test-epic")
	store := NewSessionStore(database)

	session, _, _ := store.Create(domain.SessionEpic, &epicID)
	phases, err := store.Phases(session.ID)
	if err != nil {
		t.Fatalf("Phases: %v", err)
	}
	if len(phases) != 0 {
		t.Errorf("len = %d, want 0", len(phases))
	}
}

func TestSessionStore_CurrentPhase(t *testing.T) {
	database := testDB(t)
	epicID := createEpicForTest(t, database, "test-epic")
	store := NewSessionStore(database)

	session, _, _ := store.Create(domain.SessionEpic, &epicID)

	// No phases completed — current is setup
	current, err := store.CurrentPhase(session.ID)
	if err != nil {
		t.Fatalf("CurrentPhase: %v", err)
	}
	if current != "setup" {
		t.Errorf("current = %q, want %q", current, "setup")
	}

	// Complete setup — current is reframing
	store.CompletePhase(session.ID, "setup", "{}")
	current, _ = store.CurrentPhase(session.ID)
	if current != "reframing" {
		t.Errorf("current = %q, want %q", current, "reframing")
	}
}

func TestSessionStore_CurrentPhase_AllDone(t *testing.T) {
	database := testDB(t)
	epicID := createEpicForTest(t, database, "test-epic")
	store := NewSessionStore(database)

	session, _, _ := store.Create(domain.SessionValidate, &epicID) // 3 phases
	store.CompletePhase(session.ID, "setup", "{}")
	store.CompletePhase(session.ID, "qa-dispatch", "{}")
	store.CompletePhase(session.ID, "verdict", "{}")

	current, err := store.CurrentPhase(session.ID)
	if err != nil {
		t.Fatalf("CurrentPhase: %v", err)
	}
	if current != "" {
		t.Errorf("current = %q, want empty", current)
	}
}
