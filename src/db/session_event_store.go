package db

import (
	"fmt"
	"time"
)

// SessionEvent represents a lifecycle event recorded during a Claude Code session.
type SessionEvent struct {
	ID        int
	SessionID string
	EventType string
	Payload   string
	CreatedAt time.Time
}

// SessionEventStore provides persistence for session lifecycle events.
type SessionEventStore struct {
	db *DB
}

// NewSessionEventStore returns a SessionEventStore using the given database.
func NewSessionEventStore(db *DB) *SessionEventStore {
	return &SessionEventStore{db: db}
}

// Record inserts a new session event.
func (s *SessionEventStore) Record(sessionID, eventType, payload string) error {
	_, err := s.db.Conn().Exec(
		"INSERT INTO session_events (session_id, event_type, payload) VALUES (?, ?, ?)",
		sessionID, eventType, payload,
	)
	if err != nil {
		return fmt.Errorf("record session event: %w", err)
	}
	return nil
}

// ListBySession returns up to limit session events ordered by creation time.
func (s *SessionEventStore) ListBySession(sessionID string, limit int) ([]SessionEvent, error) {
	rows, err := s.db.Conn().Query(
		"SELECT id, session_id, event_type, payload, created_at FROM session_events WHERE session_id = ? ORDER BY created_at LIMIT ?",
		sessionID, limit,
	)
	if err != nil {
		return nil, fmt.Errorf("list session events: %w", err)
	}
	defer rows.Close()

	var events []SessionEvent
	for rows.Next() {
		var (
			e         SessionEvent
			createdAt string
		)
		if err := rows.Scan(&e.ID, &e.SessionID, &e.EventType, &e.Payload, &createdAt); err != nil {
			return nil, fmt.Errorf("scan session event: %w", err)
		}
		if e.CreatedAt, err = time.Parse(sqliteTimeFmt, createdAt); err != nil {
			return nil, fmt.Errorf("parse created_at: %w", err)
		}
		events = append(events, e)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate session events: %w", err)
	}
	return events, nil
}
