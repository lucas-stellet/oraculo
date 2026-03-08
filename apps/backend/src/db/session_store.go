// src/db/session_store.go
package db

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lucas/oraculo/apps/backend/src/domain"
)

// SessionStore provides operations for sessions and phases backed by SQLite.
type SessionStore struct {
	db *DB
}

// NewSessionStore returns a SessionStore that uses the given database connection.
func NewSessionStore(db *DB) *SessionStore {
	return &SessionStore{db: db}
}

// Create starts a new session. If an active session of the same type already
// exists for the given epic, it returns that session with created=false.
func (s *SessionStore) Create(sessionType domain.SessionType, epicID *int) (*domain.Session, bool, error) {
	// Check for existing active session
	if epicID != nil {
		existing, err := s.ActiveByEpic(*epicID, sessionType)
		if err == nil {
			return existing, false, nil
		}
	}
	id := uuid.New().String()
	_, err := s.db.conn.Exec(
		"INSERT INTO sessions (id, type, epic_id) VALUES (?, ?, ?)",
		id, string(sessionType), epicID,
	)
	if err != nil {
		return nil, false, fmt.Errorf("create session: %w", err)
	}
	session, err := s.Get(id)
	if err != nil {
		return nil, false, fmt.Errorf("fetch after create: %w", err)
	}
	return session, true, nil
}

func scanSession(row interface{ Scan(...any) error }) (*domain.Session, error) {
	var (
		sess      domain.Session
		epicID    sql.NullInt64
		closedAt  sql.NullString
		createdAt string
	)
	if err := row.Scan(
		&sess.ID, &sess.Type, &epicID, &sess.Status,
		&createdAt, &closedAt,
	); err != nil {
		return nil, err
	}
	var err error
	if sess.CreatedAt, err = time.Parse(sqliteTimeFmt, createdAt); err != nil {
		return nil, fmt.Errorf("parse created_at: %w", err)
	}
	if epicID.Valid {
		v := int(epicID.Int64)
		sess.EpicID = &v
	}
	if closedAt.Valid {
		t, err := time.Parse(sqliteTimeFmt, closedAt.String)
		if err != nil {
			return nil, fmt.Errorf("parse closed_at: %w", err)
		}
		sess.ClosedAt = &t
	}
	return &sess, nil
}

// Get retrieves a session by ID. Returns ErrNotFound if absent.
func (s *SessionStore) Get(id string) (*domain.Session, error) {
	row := s.db.conn.QueryRow(
		"SELECT id, type, epic_id, status, created_at, closed_at FROM sessions WHERE id = ?",
		id,
	)
	sess, err := scanSession(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("session %q: %w", id, domain.ErrNotFound)
		}
		return nil, fmt.Errorf("get session: %w", err)
	}
	return sess, nil
}

// Close marks a session as completed or abandoned.
// Returns ErrNotFound if absent, ErrInvalidTransition if already closed.
func (s *SessionStore) Close(id string, status domain.SessionStatus) error {
	res, err := s.db.conn.Exec(
		"UPDATE sessions SET status = ?, closed_at = datetime('now') WHERE id = ? AND status = 'active'",
		string(status), id,
	)
	if err != nil {
		return fmt.Errorf("close session: %w", err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}
	if affected == 0 {
		// Distinguish not found from already closed
		_, getErr := s.Get(id)
		if getErr != nil {
			return fmt.Errorf("session %q: %w", id, domain.ErrNotFound)
		}
		return fmt.Errorf("session %q already closed: %w", id, domain.ErrInvalidTransition)
	}
	return nil
}

// CompletePhase records a phase completion with its data.
// Validates that the previous phase in the sequence has been completed.
// Auto-closes the session when the last phase completes.
func (s *SessionStore) CompletePhase(sessionID, phase, data string) error {
	session, err := s.Get(sessionID)
	if err != nil {
		return err
	}
	if session.Status != domain.SessionActive {
		return fmt.Errorf("session %q is %s: %w", sessionID, session.Status, domain.ErrInvalidTransition)
	}

	idx := domain.PhaseIndex(session.Type, phase)
	if idx < 0 {
		return fmt.Errorf("phase %q is not valid for session type %q: %w", phase, session.Type, domain.ErrInvalidPhase)
	}

	// Check predecessor
	if idx > 0 {
		predecessor := domain.Phases[session.Type][idx-1]
		var count int
		err := s.db.conn.QueryRow(
			"SELECT COUNT(*) FROM session_phases WHERE session_id = ? AND phase = ?",
			sessionID, predecessor,
		).Scan(&count)
		if err != nil {
			return fmt.Errorf("check predecessor: %w", err)
		}
		if count == 0 {
			return fmt.Errorf("phase %q must be completed before %q: %w", predecessor, phase, domain.ErrInvalidTransition)
		}
	}

	_, err = s.db.conn.Exec(
		"INSERT INTO session_phases (session_id, phase, data) VALUES (?, ?, ?)",
		sessionID, phase, data,
	)
	if err != nil {
		// PRIMARY KEY violation = duplicate
		return fmt.Errorf("phase %q already completed: %w", phase, domain.ErrAlreadyExists)
	}

	// Auto-close if last phase
	phases := domain.Phases[session.Type]
	if idx == len(phases)-1 {
		return s.Close(sessionID, domain.SessionCompleted)
	}
	return nil
}

// Phases returns all completed phases for a session, ordered by completion time.
func (s *SessionStore) Phases(sessionID string) ([]domain.Phase, error) {
	rows, err := s.db.conn.Query(
		"SELECT session_id, phase, data, completed_at FROM session_phases WHERE session_id = ? ORDER BY completed_at, rowid",
		sessionID,
	)
	if err != nil {
		return nil, fmt.Errorf("list phases: %w", err)
	}
	defer rows.Close()

	var phases []domain.Phase
	for rows.Next() {
		var p domain.Phase
		var completedAt string
		if err := rows.Scan(&p.SessionID, &p.Name, &p.Data, &completedAt); err != nil {
			return nil, fmt.Errorf("scan phase: %w", err)
		}
		if p.CompletedAt, err = time.Parse(sqliteTimeFmt, completedAt); err != nil {
			return nil, fmt.Errorf("parse completed_at: %w", err)
		}
		phases = append(phases, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate phases: %w", err)
	}
	return phases, nil
}

// CurrentPhase returns the name of the next pending phase.
// If all phases are complete, returns empty string and nil error.
func (s *SessionStore) CurrentPhase(sessionID string) (string, error) {
	session, err := s.Get(sessionID)
	if err != nil {
		return "", err
	}
	completed, err := s.Phases(sessionID)
	if err != nil {
		return "", err
	}
	done := make(map[string]bool, len(completed))
	for _, p := range completed {
		done[p.Name] = true
	}
	for _, phase := range domain.Phases[session.Type] {
		if !done[phase] {
			return phase, nil
		}
	}
	return "", nil
}

// ActiveByEpic returns the active session of the given type for an epic.
// Returns ErrNotFound if no active session exists.
func (s *SessionStore) ActiveByEpic(epicID int, sessionType domain.SessionType) (*domain.Session, error) {
	row := s.db.conn.QueryRow(
		"SELECT id, type, epic_id, status, created_at, closed_at FROM sessions WHERE epic_id = ? AND type = ? AND status = 'active'",
		epicID, string(sessionType),
	)
	sess, err := scanSession(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("active session for epic %d type %q: %w", epicID, sessionType, domain.ErrNotFound)
		}
		return nil, fmt.Errorf("active by epic: %w", err)
	}
	return sess, nil
}
