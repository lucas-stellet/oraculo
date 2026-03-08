// src/db/memory_store_test.go
package db

import (
	"testing"

	"github.com/lucas/oraculo/apps/backend/src/domain"
)

func TestMemoryStore_Store(t *testing.T) {
	database := testDB(t)
	store := NewMemoryStore(database)

	k, err := store.Store("payments", domain.CategoryPattern, "Uses repository pattern", "src/repo.go", domain.ConfidenceHigh)
	if err != nil {
		t.Fatalf("Store: %v", err)
	}
	if k.Domain != "payments" {
		t.Errorf("Domain = %q, want %q", k.Domain, "payments")
	}
	if k.Category != domain.CategoryPattern {
		t.Errorf("Category = %q, want %q", k.Category, domain.CategoryPattern)
	}
	if k.Finding != "Uses repository pattern" {
		t.Errorf("Finding = %q", k.Finding)
	}
	if k.Confidence != domain.ConfidenceHigh {
		t.Errorf("Confidence = %q, want %q", k.Confidence, domain.ConfidenceHigh)
	}
}

func TestMemoryStore_Search(t *testing.T) {
	database := testDB(t)
	store := NewMemoryStore(database)

	store.Store("payments", domain.CategoryPattern, "Uses repository pattern", "", domain.ConfidenceMedium)
	store.Store("auth", domain.CategoryConvention, "JWT tokens for authentication", "", domain.ConfidenceHigh)

	results, err := store.Search("repository", "", 10)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("len = %d, want 1", len(results))
	}
	if results[0].Domain != "payments" {
		t.Errorf("Domain = %q, want %q", results[0].Domain, "payments")
	}
}

func TestMemoryStore_Search_WithDomainFilter(t *testing.T) {
	database := testDB(t)
	store := NewMemoryStore(database)

	store.Store("payments", domain.CategoryPattern, "Uses repository pattern", "", domain.ConfidenceMedium)
	store.Store("auth", domain.CategoryPattern, "Uses repository pattern too", "", domain.ConfidenceMedium)

	results, err := store.Search("repository", "payments", 10)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("len = %d, want 1", len(results))
	}
}

func TestMemoryStore_Search_WithLimit(t *testing.T) {
	database := testDB(t)
	store := NewMemoryStore(database)

	store.Store("a", domain.CategoryPattern, "pattern one", "", domain.ConfidenceMedium)
	store.Store("b", domain.CategoryPattern, "pattern two", "", domain.ConfidenceMedium)
	store.Store("c", domain.CategoryPattern, "pattern three", "", domain.ConfidenceMedium)

	results, err := store.Search("pattern", "", 2)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("len = %d, want 2", len(results))
	}
}

func TestMemoryStore_Search_NoMatch(t *testing.T) {
	database := testDB(t)
	store := NewMemoryStore(database)

	store.Store("payments", domain.CategoryPattern, "Uses repository pattern", "", domain.ConfidenceMedium)

	results, err := store.Search("nonexistent", "", 10)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("len = %d, want 0", len(results))
	}
}

func TestMemoryStore_Domains(t *testing.T) {
	database := testDB(t)
	store := NewMemoryStore(database)

	store.Store("payments", domain.CategoryPattern, "finding 1", "", domain.ConfidenceMedium)
	store.Store("auth", domain.CategoryConvention, "finding 2", "", domain.ConfidenceMedium)
	store.Store("payments", domain.CategoryConstraint, "finding 3", "", domain.ConfidenceMedium)

	domains, err := store.Domains()
	if err != nil {
		t.Fatalf("Domains: %v", err)
	}
	if len(domains) != 2 {
		t.Fatalf("len = %d, want 2", len(domains))
	}
	// Should be sorted
	if domains[0] != "auth" || domains[1] != "payments" {
		t.Errorf("domains = %v, want [auth payments]", domains)
	}
}
