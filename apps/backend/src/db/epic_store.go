package db

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/lucas/oraculo/apps/backend/src/domain"
)

const sqliteTimeFmt = "2006-01-02 15:04:05"

// EpicStore provides CRUD operations for epics backed by SQLite.
type EpicStore struct {
	db *DB
}

// NewEpicStore returns an EpicStore that uses the given database connection.
func NewEpicStore(db *DB) *EpicStore {
	return &EpicStore{db: db}
}

// scanEpic reads an Epic from a row scanner, parsing SQLite TEXT timestamps.
func scanEpic(row interface{ Scan(...any) error }) (*domain.Epic, error) {
	var (
		e                    domain.Epic
		createdAt, updatedAt string
	)
	if err := row.Scan(
		&e.ID, &e.Name, &e.Description, &e.ApprovalStatus,
		&createdAt, &updatedAt,
	); err != nil {
		return nil, err
	}
	var err error
	if e.CreatedAt, err = time.Parse(sqliteTimeFmt, createdAt); err != nil {
		return nil, fmt.Errorf("parse created_at: %w", err)
	}
	if e.UpdatedAt, err = time.Parse(sqliteTimeFmt, updatedAt); err != nil {
		return nil, fmt.Errorf("parse updated_at: %w", err)
	}
	return &e, nil
}

// Create inserts a new epic or returns the existing one if the name already exists.
// The created flag is true when a new row was inserted.
func (s *EpicStore) Create(name, description string) (*domain.Epic, bool, error) {
	res, err := s.db.conn.Exec(
		"INSERT OR IGNORE INTO epics (name, description) VALUES (?, ?)",
		name, description,
	)
	if err != nil {
		return nil, false, fmt.Errorf("create epic: %w", err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return nil, false, fmt.Errorf("rows affected: %w", err)
	}
	epic, err := s.GetByName(name)
	if err != nil {
		return nil, false, fmt.Errorf("fetch after create: %w", err)
	}
	return epic, affected > 0, nil
}

// GetByName retrieves an epic by its unique name.
// It returns domain.ErrNotFound when no matching epic exists.
func (s *EpicStore) GetByName(name string) (*domain.Epic, error) {
	row := s.db.conn.QueryRow(
		"SELECT id, name, description, approval_status, created_at, updated_at FROM epics WHERE name = ?",
		name,
	)
	epic, err := scanEpic(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("epic %q: %w", name, domain.ErrNotFound)
		}
		return nil, fmt.Errorf("get epic: %w", err)
	}
	return epic, nil
}

// List returns all epics ordered by creation time.
func (s *EpicStore) List() ([]domain.Epic, error) {
	rows, err := s.db.conn.Query(
		"SELECT id, name, description, approval_status, created_at, updated_at FROM epics ORDER BY created_at",
	)
	if err != nil {
		return nil, fmt.Errorf("list epics: %w", err)
	}
	defer rows.Close()

	var epics []domain.Epic
	for rows.Next() {
		epic, err := scanEpic(rows)
		if err != nil {
			return nil, fmt.Errorf("scan epic: %w", err)
		}
		epics = append(epics, *epic)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate epics: %w", err)
	}
	return epics, nil
}

// Update changes the description of an existing epic.
// It returns domain.ErrNotFound when no matching epic exists.
func (s *EpicStore) Update(name, description string) (*domain.Epic, error) {
	res, err := s.db.conn.Exec(
		"UPDATE epics SET description = ?, updated_at = datetime('now') WHERE name = ?",
		description, name,
	)
	if err != nil {
		return nil, fmt.Errorf("update epic: %w", err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("rows affected: %w", err)
	}
	if affected == 0 {
		return nil, fmt.Errorf("epic %q: %w", name, domain.ErrNotFound)
	}
	return s.GetByName(name)
}

// Delete removes an epic by name.
// It returns domain.ErrNotFound when no matching epic exists.
func (s *EpicStore) Delete(name string) error {
	res, err := s.db.conn.Exec("DELETE FROM epics WHERE name = ?", name)
	if err != nil {
		return fmt.Errorf("delete epic: %w", err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}
	if affected == 0 {
		return fmt.Errorf("epic %q: %w", name, domain.ErrNotFound)
	}
	return nil
}

// UpdateApprovalStatus sets the approval status for an epic.
// It returns domain.ErrNotFound when no matching epic exists.
func (s *EpicStore) UpdateApprovalStatus(name string, status domain.ApprovalStatus) error {
	res, err := s.db.conn.Exec(
		"UPDATE epics SET approval_status = ?, updated_at = datetime('now') WHERE name = ?",
		status, name,
	)
	if err != nil {
		return fmt.Errorf("update approval status: %w", err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}
	if affected == 0 {
		return fmt.Errorf("epic %q: %w", name, domain.ErrNotFound)
	}
	return nil
}

// ListSummaries returns all epics with aggregated phase, story, and task data.
func (s *EpicStore) ListSummaries() ([]domain.EpicSummary, error) {
	query := `
		SELECT
			e.id, e.name, e.description, e.approval_status, e.created_at, e.updated_at,
			COALESCE(active_s.type, last_s.type, '') AS session_type,
			CASE
				WHEN active_s.id IS NOT NULL THEN 'active'
				WHEN last_s.id IS NOT NULL THEN 'completed'
				ELSE 'pending'
			END AS phase_status,
			COALESCE(story_counts.cnt, 0) AS story_count,
			COALESCE(task_counts.total, 0) AS task_count,
			COALESCE(task_counts.completed, 0) AS completed_task_count
		FROM epics e
		LEFT JOIN sessions active_s
			ON active_s.epic_id = e.id AND active_s.status = 'active'
		LEFT JOIN (
			SELECT epic_id, id, type
			FROM sessions
			WHERE status = 'completed'
			ORDER BY closed_at DESC
		) last_s ON last_s.epic_id = e.id AND active_s.id IS NULL
		LEFT JOIN (
			SELECT epic_id, COUNT(*) AS cnt
			FROM stories
			GROUP BY epic_id
		) story_counts ON story_counts.epic_id = e.id
		LEFT JOIN (
			SELECT s.epic_id,
				COUNT(t.id) AS total,
				COUNT(CASE WHEN t.status = 'completed' THEN 1 END) AS completed
			FROM tasks t
			JOIN stories s ON t.story_id = s.id
			GROUP BY s.epic_id
		) task_counts ON task_counts.epic_id = e.id
		ORDER BY e.created_at
	`

	rows, err := s.db.conn.Query(query)
	if err != nil {
		return nil, fmt.Errorf("list epic summaries: %w", err)
	}
	defer rows.Close()

	var summaries []domain.EpicSummary
	for rows.Next() {
		var (
			es                   domain.EpicSummary
			createdAt, updatedAt string
			sessionType          string
		)
		if err := rows.Scan(
			&es.ID, &es.Name, &es.Description, &es.ApprovalStatus,
			&createdAt, &updatedAt,
			&sessionType, &es.PhaseStatus,
			&es.StoryCount, &es.TaskCount, &es.CompletedTaskCount,
		); err != nil {
			return nil, fmt.Errorf("scan epic summary: %w", err)
		}
		var parseErr error
		if es.CreatedAt, parseErr = time.Parse(sqliteTimeFmt, createdAt); parseErr != nil {
			return nil, fmt.Errorf("parse created_at: %w", parseErr)
		}
		if es.UpdatedAt, parseErr = time.Parse(sqliteTimeFmt, updatedAt); parseErr != nil {
			return nil, fmt.Errorf("parse updated_at: %w", parseErr)
		}

		if sessionType != "" {
			es.Phase = domain.SessionTypeToPhase[domain.SessionType(sessionType)]
		}
		if es.Phase == "" {
			es.Phase = "discover"
		}

		summaries = append(summaries, es)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate epic summaries: %w", err)
	}
	if summaries == nil {
		summaries = []domain.EpicSummary{}
	}
	return summaries, nil
}
