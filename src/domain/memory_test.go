// src/domain/memory_test.go
package domain

import "testing"

func TestCategory_Valid(t *testing.T) {
	tests := []struct {
		c    Category
		want bool
	}{
		{CategoryPattern, true},
		{CategoryConvention, true},
		{CategoryConstraint, true},
		{CategoryDependency, true},
		{CategoryTest, true},
		{CategoryArchitecture, true},
		{Category("bogus"), false},
	}
	for _, tt := range tests {
		if got := tt.c.Valid(); got != tt.want {
			t.Errorf("Category(%q).Valid() = %v, want %v", tt.c, got, tt.want)
		}
	}
}

func TestConfidence_Valid(t *testing.T) {
	tests := []struct {
		c    Confidence
		want bool
	}{
		{ConfidenceHigh, true},
		{ConfidenceMedium, true},
		{ConfidenceLow, true},
		{Confidence("none"), false},
	}
	for _, tt := range tests {
		if got := tt.c.Valid(); got != tt.want {
			t.Errorf("Confidence(%q).Valid() = %v, want %v", tt.c, got, tt.want)
		}
	}
}
