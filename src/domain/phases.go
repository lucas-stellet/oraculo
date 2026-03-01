// src/domain/phases.go
package domain

// Phases defines the ordered sequence for each session type.
// The CLI enforces that phase N completes only if phase N-1 exists.
var Phases = map[SessionType][]string{
	SessionEpic:     {"setup", "reframing", "divergence", "codebase", "convergence", "assumptions", "stress-test", "exit-gate", "artifact"},
	SessionStory:    {"setup", "reframing", "assumptions", "exit-gate", "artifact"},
	SessionPlan:     {"setup", "decomposition", "dependencies", "optimization", "artifact"},
	SessionExecute:  {"setup", "team-assembly", "monitoring", "completion"},
	SessionValidate: {"setup", "qa-dispatch", "verdict"},
}

// PhaseIndex returns the position of phase in the sequence for the given
// session type. It returns -1 if the session type or phase is not recognized.
func PhaseIndex(sessionType SessionType, phase string) int {
	phases, ok := Phases[sessionType]
	if !ok {
		return -1
	}
	for i, p := range phases {
		if p == phase {
			return i
		}
	}
	return -1
}
