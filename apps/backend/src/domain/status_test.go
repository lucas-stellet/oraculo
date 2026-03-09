// apps/backend/src/domain/status_test.go
package domain

import "testing"

func TestApprovalStatus_Valid(t *testing.T) {
	tests := []struct {
		s    ApprovalStatus
		want bool
	}{
		{ApprovalNone, true},
		{ApprovalPending, true},
		{ApprovalApproved, true},
		{ApprovalRejected, true},
		{ApprovalNeedsRevision, true},
		{ApprovalStatus("invalid"), false},
	}
	for _, tt := range tests {
		if got := tt.s.Valid(); got != tt.want {
			t.Errorf("ApprovalStatus(%q).Valid() = %v, want %v", tt.s, got, tt.want)
		}
	}
}
