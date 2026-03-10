package db

import (
	"database/sql"
	"fmt"
	"time"
)

// Agent represents a running or completed agent within a session.
type Agent struct {
	ID        int        `json:"id"`
	SessionID string     `json:"session_id"`
	Name      string     `json:"name"`
	Type      string     `json:"type"`
	Status    string     `json:"status"`
	StartedAt time.Time  `json:"started_at"`
	StoppedAt *time.Time `json:"stopped_at"`
	TaskID    *int       `json:"task_id"`
}

// AgentStore provides persistence operations for agents backed by SQLite.
type AgentStore struct {
	db *DB
}

// NewAgentStore returns an AgentStore that uses the given database connection.
func NewAgentStore(db *DB) *AgentStore {
	return &AgentStore{db: db}
}

// scanAgent reads an Agent from a row scanner, parsing SQLite TEXT timestamps.
func scanAgent(row interface{ Scan(...any) error }) (*Agent, error) {
	var (
		a         Agent
		startedAt string
		stoppedAt sql.NullString
		taskID    sql.NullInt64
	)
	if err := row.Scan(
		&a.ID, &a.SessionID, &a.Name, &a.Type, &a.Status,
		&startedAt, &stoppedAt, &taskID,
	); err != nil {
		return nil, err
	}
	var err error
	if a.StartedAt, err = time.Parse(sqliteTimeFmt, startedAt); err != nil {
		return nil, fmt.Errorf("parse started_at: %w", err)
	}
	if stoppedAt.Valid {
		t, err := time.Parse(sqliteTimeFmt, stoppedAt.String)
		if err != nil {
			return nil, fmt.Errorf("parse stopped_at: %w", err)
		}
		a.StoppedAt = &t
	}
	if taskID.Valid {
		id := int(taskID.Int64)
		a.TaskID = &id
	}
	return &a, nil
}

// Start inserts a new agent with status "active" and started_at set to now.
// taskID is optional; pass nil to create an agent not associated with any task.
func (s *AgentStore) Start(sessionID, name, agentType string, taskID *int) (*Agent, error) {
	now := time.Now().UTC().Format(sqliteTimeFmt)
	var res sql.Result
	var err error
	if taskID != nil {
		res, err = s.db.conn.Exec(
			"INSERT INTO agents (session_id, name, type, status, started_at, task_id) VALUES (?, ?, ?, 'active', ?, ?)",
			sessionID, name, agentType, now, *taskID,
		)
	} else {
		res, err = s.db.conn.Exec(
			"INSERT INTO agents (session_id, name, type, status, started_at) VALUES (?, ?, ?, 'active', ?)",
			sessionID, name, agentType, now,
		)
	}
	if err != nil {
		return nil, fmt.Errorf("start agent: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("last insert id: %w", err)
	}
	return s.getByID(int(id))
}

// Stop updates the agent's status and sets stopped_at to now.
func (s *AgentStore) Stop(id int, status string) (*Agent, error) {
	now := time.Now().UTC().Format(sqliteTimeFmt)
	res, err := s.db.conn.Exec(
		"UPDATE agents SET status = ?, stopped_at = ? WHERE id = ?",
		status, now, id,
	)
	if err != nil {
		return nil, fmt.Errorf("stop agent: %w", err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("rows affected: %w", err)
	}
	if affected == 0 {
		return nil, fmt.Errorf("agent %d: %w", id, sql.ErrNoRows)
	}
	return s.getByID(id)
}

// ListBySession returns all agents for the given session.
func (s *AgentStore) ListBySession(sessionID string) ([]Agent, error) {
	rows, err := s.db.conn.Query(
		"SELECT id, session_id, name, type, status, started_at, stopped_at, task_id FROM agents WHERE session_id = ? ORDER BY started_at",
		sessionID,
	)
	if err != nil {
		return nil, fmt.Errorf("list agents: %w", err)
	}
	defer rows.Close()

	var agents []Agent
	for rows.Next() {
		agent, err := scanAgent(rows)
		if err != nil {
			return nil, fmt.Errorf("scan agent: %w", err)
		}
		agents = append(agents, *agent)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate agents: %w", err)
	}
	return agents, nil
}

// getByID retrieves an agent by its primary key.
func (s *AgentStore) getByID(id int) (*Agent, error) {
	row := s.db.conn.QueryRow(
		"SELECT id, session_id, name, type, status, started_at, stopped_at, task_id FROM agents WHERE id = ?",
		id,
	)
	agent, err := scanAgent(row)
	if err != nil {
		return nil, fmt.Errorf("get agent %d: %w", id, err)
	}
	return agent, nil
}
