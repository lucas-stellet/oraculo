// src/domain/approval_test.go
package domain

import "testing"

func TestApprovalType_Valid(t *testing.T) {
	tests := []struct {
		a    ApprovalType
		want bool
	}{
		{ApprovalQAEscalation, true},
		{ApprovalExecutionPlan, true},
		{ApprovalDesign, true},
		{ApprovalType("unknown"), false},
	}
	for _, tt := range tests {
		if got := tt.a.Valid(); got != tt.want {
			t.Errorf("ApprovalType(%q).Valid() = %v, want %v", tt.a, got, tt.want)
		}
	}
}

func TestVerdict_Valid(t *testing.T) {
	tests := []struct {
		v    Verdict
		want bool
	}{
		{VerdictApproved, true},
		{VerdictRejected, true},
		{VerdictNeedsRevision, true},
		{Verdict("maybe"), false},
	}
	for _, tt := range tests {
		if got := tt.v.Valid(); got != tt.want {
			t.Errorf("Verdict(%q).Valid() = %v, want %v", tt.v, got, tt.want)
		}
	}
}
