package domain

import "testing"

func TestSessionType_Valid(t *testing.T) {
	tests := []struct {
		s    SessionType
		want bool
	}{
		{SessionEpic, true},
		{SessionStory, true},
		{SessionPlan, true},
		{SessionExecute, true},
		{SessionValidate, true},
		{SessionType("unknown"), false},
	}
	for _, tt := range tests {
		if got := tt.s.Valid(); got != tt.want {
			t.Errorf("SessionType(%q).Valid() = %v, want %v", tt.s, got, tt.want)
		}
	}
}

func TestSessionStatus_Valid(t *testing.T) {
	tests := []struct {
		s    SessionStatus
		want bool
	}{
		{SessionActive, true},
		{SessionCompleted, true},
		{SessionAbandoned, true},
		{SessionStatus("unknown"), false},
	}
	for _, tt := range tests {
		if got := tt.s.Valid(); got != tt.want {
			t.Errorf("SessionStatus(%q).Valid() = %v, want %v", tt.s, got, tt.want)
		}
	}
}
