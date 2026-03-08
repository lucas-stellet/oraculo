package tools_test

import (
	"encoding/json"
	"strings"
	"testing"
)

// ----- store subcommand -----

func TestMemoryStore(t *testing.T) {
	setupTestDir(t)

	out, err := executeCmd(t,
		"tools", "memory", "store",
		"--domain", "payments",
		"--category", "pattern",
		"--finding", "Uses repository pattern",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v\noutput: %s", err, out)
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, out)
	}
	if result["Domain"] != "payments" {
		t.Errorf("Domain = %v, want %q", result["Domain"], "payments")
	}
	if result["Category"] != "pattern" {
		t.Errorf("Category = %v, want %q", result["Category"], "pattern")
	}
	if result["Finding"] != "Uses repository pattern" {
		t.Errorf("Finding = %v, want %q", result["Finding"], "Uses repository pattern")
	}
	// Default confidence should be medium.
	if result["Confidence"] != "medium" {
		t.Errorf("Confidence = %v, want %q", result["Confidence"], "medium")
	}
}

func TestMemoryStore_AllFlags(t *testing.T) {
	setupTestDir(t)

	out, err := executeCmd(t,
		"tools", "memory", "store",
		"--domain", "payments",
		"--category", "pattern",
		"--finding", "Uses repo",
		"--source", "src/repo.go",
		"--confidence", "high",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v\noutput: %s", err, out)
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, out)
	}
	if result["Domain"] != "payments" {
		t.Errorf("Domain = %v, want %q", result["Domain"], "payments")
	}
	if result["Category"] != "pattern" {
		t.Errorf("Category = %v, want %q", result["Category"], "pattern")
	}
	if result["Finding"] != "Uses repo" {
		t.Errorf("Finding = %v, want %q", result["Finding"], "Uses repo")
	}
	if result["SourceFiles"] != "src/repo.go" {
		t.Errorf("SourceFiles = %v, want %q", result["SourceFiles"], "src/repo.go")
	}
	if result["Confidence"] != "high" {
		t.Errorf("Confidence = %v, want %q", result["Confidence"], "high")
	}
}

func TestMemoryStore_MissingRequiredFlags(t *testing.T) {
	setupTestDir(t)

	// Missing --domain, --category, --finding.
	_, err := executeCmd(t, "tools", "memory", "store")
	if err == nil {
		t.Fatal("expected error for missing required flags")
	}
	// Cobra returns the missing-flag error.
	errMsg := err.Error()
	if !strings.Contains(errMsg, "required") || !strings.Contains(errMsg, "domain") {
		t.Errorf("expected error about required domain flag, got: %s", errMsg)
	}
}

func TestMemoryStore_InvalidCategory(t *testing.T) {
	setupTestDir(t)

	out, err := executeCmd(t,
		"tools", "memory", "store",
		"--domain", "payments",
		"--category", "bogus",
		"--finding", "something",
	)
	if err == nil {
		t.Fatal("expected error for invalid category")
	}

	var result map[string]string
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, out)
	}
	if result["error"] == "" {
		t.Error("expected non-empty error code")
	}
	if !strings.Contains(result["message"], "invalid category") {
		t.Errorf("message = %q, expected to contain 'invalid category'", result["message"])
	}
}

func TestMemoryStore_InvalidConfidence(t *testing.T) {
	setupTestDir(t)

	out, err := executeCmd(t,
		"tools", "memory", "store",
		"--domain", "payments",
		"--category", "pattern",
		"--finding", "something",
		"--confidence", "maybe",
	)
	if err == nil {
		t.Fatal("expected error for invalid confidence")
	}

	var result map[string]string
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, out)
	}
	if result["error"] == "" {
		t.Error("expected non-empty error code")
	}
	if !strings.Contains(result["message"], "invalid confidence") {
		t.Errorf("message = %q, expected to contain 'invalid confidence'", result["message"])
	}
}

// ----- search subcommand -----

func TestMemorySearch(t *testing.T) {
	setupTestDir(t)

	// Store a finding first.
	_, err := executeCmd(t,
		"tools", "memory", "store",
		"--domain", "payments",
		"--category", "pattern",
		"--finding", "Uses repository pattern for data access",
	)
	if err != nil {
		t.Fatalf("store: %v", err)
	}

	out, err := executeCmd(t, "tools", "memory", "search", "repository")
	if err != nil {
		t.Fatalf("search: %v\noutput: %s", err, out)
	}

	var results []map[string]any
	if err := json.Unmarshal([]byte(out), &results); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, out)
	}
	if len(results) == 0 {
		t.Fatal("expected at least one result")
	}
	if results[0]["Finding"] != "Uses repository pattern for data access" {
		t.Errorf("Finding = %v, want %q", results[0]["Finding"], "Uses repository pattern for data access")
	}
}

func TestMemorySearch_WithDomainFilter(t *testing.T) {
	setupTestDir(t)

	// Store findings in two domains.
	_, _ = executeCmd(t,
		"tools", "memory", "store",
		"--domain", "payments",
		"--category", "pattern",
		"--finding", "Payment uses repository pattern",
	)
	_, _ = executeCmd(t,
		"tools", "memory", "store",
		"--domain", "auth",
		"--category", "pattern",
		"--finding", "Auth uses repository pattern",
	)

	out, err := executeCmd(t, "tools", "memory", "search", "repository", "--domain", "payments")
	if err != nil {
		t.Fatalf("search: %v\noutput: %s", err, out)
	}

	var results []map[string]any
	if err := json.Unmarshal([]byte(out), &results); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, out)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0]["Domain"] != "payments" {
		t.Errorf("Domain = %v, want %q", results[0]["Domain"], "payments")
	}
}

func TestMemorySearch_NoResults(t *testing.T) {
	setupTestDir(t)

	out, err := executeCmd(t, "tools", "memory", "search", "nonexistent")
	if err != nil {
		t.Fatalf("search: %v\noutput: %s", err, out)
	}

	out = strings.TrimSpace(out)
	if out != "[]" {
		t.Errorf("expected empty JSON array [], got: %s", out)
	}
}

func TestMemorySearch_MissingQuery(t *testing.T) {
	setupTestDir(t)

	_, err := executeCmd(t, "tools", "memory", "search")
	if err == nil {
		t.Fatal("expected error for missing query arg")
	}
}

func TestMemorySearch_WithLimit(t *testing.T) {
	setupTestDir(t)

	// Store two findings.
	_, _ = executeCmd(t,
		"tools", "memory", "store",
		"--domain", "payments",
		"--category", "pattern",
		"--finding", "First repository pattern",
	)
	_, _ = executeCmd(t,
		"tools", "memory", "store",
		"--domain", "payments",
		"--category", "convention",
		"--finding", "Second repository convention",
	)

	out, err := executeCmd(t, "tools", "memory", "search", "repository", "--limit", "1")
	if err != nil {
		t.Fatalf("search: %v\noutput: %s", err, out)
	}

	var results []map[string]any
	if err := json.Unmarshal([]byte(out), &results); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, out)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result with limit=1, got %d", len(results))
	}
}

// ----- domains subcommand -----

func TestMemoryDomains(t *testing.T) {
	setupTestDir(t)

	// Store findings in two domains.
	_, _ = executeCmd(t,
		"tools", "memory", "store",
		"--domain", "payments",
		"--category", "pattern",
		"--finding", "something",
	)
	_, _ = executeCmd(t,
		"tools", "memory", "store",
		"--domain", "auth",
		"--category", "convention",
		"--finding", "something else",
	)

	out, err := executeCmd(t, "tools", "memory", "domains")
	if err != nil {
		t.Fatalf("domains: %v\noutput: %s", err, out)
	}

	var domains []string
	if err := json.Unmarshal([]byte(out), &domains); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, out)
	}
	if len(domains) != 2 {
		t.Fatalf("expected 2 domains, got %d", len(domains))
	}
	// Sorted alphabetically: auth, payments.
	if domains[0] != "auth" {
		t.Errorf("domains[0] = %q, want %q", domains[0], "auth")
	}
	if domains[1] != "payments" {
		t.Errorf("domains[1] = %q, want %q", domains[1], "payments")
	}
}

func TestMemoryDomains_Empty(t *testing.T) {
	setupTestDir(t)

	out, err := executeCmd(t, "tools", "memory", "domains")
	if err != nil {
		t.Fatalf("domains: %v\noutput: %s", err, out)
	}

	out = strings.TrimSpace(out)
	if out != "null" && out != "[]" {
		t.Errorf("expected null or [], got: %s", out)
	}
}
