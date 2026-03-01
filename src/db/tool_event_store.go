package db

import (
	"fmt"
	"time"
)

// ToolEvent represents a single tool invocation recorded during a session.
type ToolEvent struct {
	ID        int
	SessionID string
	ToolName  string
	FilePath  string
	Timestamp time.Time
}

// ToolEventStore provides persistence operations for tool events backed by SQLite.
type ToolEventStore struct {
	db *DB
}

// NewToolEventStore returns a ToolEventStore that uses the given database connection.
func NewToolEventStore(db *DB) *ToolEventStore {
	return &ToolEventStore{db: db}
}

// Record inserts a new tool event with timestamp set to now.
func (s *ToolEventStore) Record(sessionID, toolName, filePath string) error {
	now := time.Now().UTC().Format(sqliteTimeFmt)
	_, err := s.db.conn.Exec(
		"INSERT INTO tool_events (session_id, tool_name, file_path, timestamp) VALUES (?, ?, ?, ?)",
		sessionID, toolName, filePath, now,
	)
	if err != nil {
		return fmt.Errorf("record tool event: %w", err)
	}
	return nil
}

// ListBySession returns up to limit tool events for the given session, ordered by timestamp.
func (s *ToolEventStore) ListBySession(sessionID string, limit int) ([]ToolEvent, error) {
	rows, err := s.db.conn.Query(
		"SELECT id, session_id, tool_name, COALESCE(file_path, ''), timestamp FROM tool_events WHERE session_id = ? ORDER BY timestamp LIMIT ?",
		sessionID, limit,
	)
	if err != nil {
		return nil, fmt.Errorf("list tool events: %w", err)
	}
	defer rows.Close()

	var events []ToolEvent
	for rows.Next() {
		var (
			e         ToolEvent
			timestamp string
		)
		if err := rows.Scan(&e.ID, &e.SessionID, &e.ToolName, &e.FilePath, &timestamp); err != nil {
			return nil, fmt.Errorf("scan tool event: %w", err)
		}
		if e.Timestamp, err = time.Parse(sqliteTimeFmt, timestamp); err != nil {
			return nil, fmt.Errorf("parse timestamp: %w", err)
		}
		events = append(events, e)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate tool events: %w", err)
	}
	return events, nil
}
