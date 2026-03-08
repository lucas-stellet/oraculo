package db

import (
	"errors"
	"testing"

	"github.com/lucas/oraculo/apps/backend/src/domain"
)

func TestEpicStore_Create(t *testing.T) {
	database := testDB(t)
	store := NewEpicStore(database)

	epic, created, err := store.Create("my-epic", "a description")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if !created {
		t.Error("expected created=true on first call")
	}
	if epic.Name != "my-epic" {
		t.Errorf("Name = %q, want %q", epic.Name, "my-epic")
	}
	if epic.Description != "a description" {
		t.Errorf("Description = %q, want %q", epic.Description, "a description")
	}
}

func TestEpicStore_CreateIdempotent(t *testing.T) {
	database := testDB(t)
	store := NewEpicStore(database)

	e1, _, _ := store.Create("my-epic", "desc")
	e2, created, err := store.Create("my-epic", "desc")
	if err != nil {
		t.Fatalf("second Create: %v", err)
	}
	if created {
		t.Error("expected created=false on second call")
	}
	if e2.ID != e1.ID {
		t.Errorf("IDs differ: %d vs %d", e2.ID, e1.ID)
	}
}

func TestEpicStore_GetByName(t *testing.T) {
	database := testDB(t)
	store := NewEpicStore(database)

	store.Create("test-epic", "desc")
	epic, err := store.GetByName("test-epic")
	if err != nil {
		t.Fatalf("GetByName: %v", err)
	}
	if epic.Name != "test-epic" {
		t.Errorf("Name = %q, want %q", epic.Name, "test-epic")
	}
}

func TestEpicStore_GetByName_NotFound(t *testing.T) {
	database := testDB(t)
	store := NewEpicStore(database)

	_, err := store.GetByName("nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent epic")
	}
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got: %v", err)
	}
}

func TestEpicStore_List(t *testing.T) {
	database := testDB(t)
	store := NewEpicStore(database)

	store.Create("alpha", "")
	store.Create("beta", "")

	epics, err := store.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(epics) != 2 {
		t.Fatalf("len = %d, want 2", len(epics))
	}
}

func TestEpicStore_Update(t *testing.T) {
	database := testDB(t)
	store := NewEpicStore(database)

	store.Create("my-epic", "old")
	epic, err := store.Update("my-epic", "new description")
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if epic.Description != "new description" {
		t.Errorf("Description = %q, want %q", epic.Description, "new description")
	}
}

func TestEpicStore_Update_NotFound(t *testing.T) {
	database := testDB(t)
	store := NewEpicStore(database)

	_, err := store.Update("nonexistent", "desc")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got: %v", err)
	}
}

func TestEpicStore_Delete(t *testing.T) {
	database := testDB(t)
	store := NewEpicStore(database)

	store.Create("doomed", "")
	if err := store.Delete("doomed"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	_, err := store.GetByName("doomed")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Error("expected ErrNotFound after delete")
	}
}

func TestEpicStore_Delete_NotFound(t *testing.T) {
	database := testDB(t)
	store := NewEpicStore(database)

	err := store.Delete("nonexistent")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got: %v", err)
	}
}

func TestEpicStore_UpdateApprovalStatus(t *testing.T) {
	database := testDB(t)
	store := NewEpicStore(database)

	store.Create("my-epic", "")
	if err := store.UpdateApprovalStatus("my-epic", domain.ApprovalPending); err != nil {
		t.Fatalf("UpdateApprovalStatus: %v", err)
	}
	epic, _ := store.GetByName("my-epic")
	if epic.ApprovalStatus != domain.ApprovalPending {
		t.Errorf("ApprovalStatus = %q, want %q", epic.ApprovalStatus, domain.ApprovalPending)
	}
}
