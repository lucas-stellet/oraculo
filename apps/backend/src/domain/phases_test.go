package domain

import "testing"

func TestPhaseIndex(t *testing.T) {
	tests := []struct {
		sessionType SessionType
		phase       string
		want        int
	}{
		{SessionEpic, "setup", 0},
		{SessionEpic, "reframing", 1},
		{SessionEpic, "artifact", 8},
		{SessionStory, "setup", 0},
		{SessionStory, "artifact", 4},
		{SessionStory, "approval", 5},
		{SessionPlan, "decomposition", 1},
		{SessionPlan, "design", 3},
		{SessionPlan, "optimization", 4},
		{SessionPlan, "artifact", 5},
		{SessionExecute, "team-assembly", 1},
		{SessionValidate, "verdict", 2},
	}
	for _, tt := range tests {
		got := PhaseIndex(tt.sessionType, tt.phase)
		if got != tt.want {
			t.Errorf("PhaseIndex(%q, %q) = %d, want %d", tt.sessionType, tt.phase, got, tt.want)
		}
	}
}

func TestPhaseIndex_Unknown(t *testing.T) {
	tests := []struct {
		sessionType SessionType
		phase       string
	}{
		{SessionEpic, "nonexistent"},
		{SessionType("bad"), "setup"},
		{SessionStory, "divergence"}, // valid for epic, not story
	}
	for _, tt := range tests {
		got := PhaseIndex(tt.sessionType, tt.phase)
		if got != -1 {
			t.Errorf("PhaseIndex(%q, %q) = %d, want -1", tt.sessionType, tt.phase, got)
		}
	}
}

func TestPhasesCount(t *testing.T) {
	tests := []struct {
		sessionType SessionType
		want        int
	}{
		{SessionEpic, 9},
		{SessionStory, 6},
		{SessionPlan, 6},
		{SessionExecute, 4},
		{SessionValidate, 3},
	}
	for _, tt := range tests {
		got := len(Phases[tt.sessionType])
		if got != tt.want {
			t.Errorf("len(Phases[%q]) = %d, want %d", tt.sessionType, got, tt.want)
		}
	}
}
