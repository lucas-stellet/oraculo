// apps/backend/src/db/story_store.go
package db

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/lucas/oraculo/apps/backend/src/domain"
)

// StoryStore provides CRUD operations for stories backed by SQLite.
// All operations are scoped by epicID.
type StoryStore struct {
	db *DB
}

// NewStoryStore returns a StoryStore that uses the given database connection.
func NewStoryStore(db *DB) *StoryStore {
	return &StoryStore{db: db}
}

// scanStory reads a Story from a row scanner, parsing SQLite TEXT timestamps.
func scanStory(row interface{ Scan(...any) error }) (*domain.Story, error) {
	var (
		s                    domain.Story
		createdAt, updatedAt string
	)
	if err := row.Scan(
		&s.ID, &s.EpicID, &s.Name, &s.Description, &s.ApprovalStatus,
		&createdAt, &updatedAt,
	); err != nil {
		return nil, err
	}
	var err error
	if s.CreatedAt, err = time.Parse(sqliteTimeFmt, createdAt); err != nil {
		return nil, fmt.Errorf("parse created_at: %w", err)
	}
	if s.UpdatedAt, err = time.Parse(sqliteTimeFmt, updatedAt); err != nil {
		return nil, fmt.Errorf("parse updated_at: %w", err)
	}
	return &s, nil
}

// Create inserts a new story or returns the existing one if (epicID, name) already exists.
// The created flag is true when a new row was inserted.
func (s *StoryStore) Create(epicID int, name, description string) (*domain.Story, bool, error) {
	res, err := s.db.conn.Exec(
		"INSERT OR IGNORE INTO stories (epic_id, name, description) VALUES (?, ?, ?)",
		epicID, name, description,
	)
	if err != nil {
		return nil, false, fmt.Errorf("create story: %w", err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return nil, false, fmt.Errorf("rows affected: %w", err)
	}
	story, err := s.GetByName(epicID, name)
	if err != nil {
		return nil, false, fmt.Errorf("fetch after create: %w", err)
	}
	return story, affected > 0, nil
}

// GetByName retrieves a story by its unique (epicID, name) pair.
// It returns domain.ErrNotFound when no matching story exists.
func (s *StoryStore) GetByName(epicID int, name string) (*domain.Story, error) {
	row := s.db.conn.QueryRow(
		"SELECT id, epic_id, name, description, approval_status, created_at, updated_at FROM stories WHERE epic_id = ? AND name = ?",
		epicID, name,
	)
	story, err := scanStory(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("story %q (epic %d): %w", name, epicID, domain.ErrNotFound)
		}
		return nil, fmt.Errorf("get story: %w", err)
	}
	return story, nil
}

// List returns all stories for the given epicID ordered by creation time.
func (s *StoryStore) List(epicID int) ([]domain.Story, error) {
	rows, err := s.db.conn.Query(
		"SELECT id, epic_id, name, description, approval_status, created_at, updated_at FROM stories WHERE epic_id = ? ORDER BY created_at",
		epicID,
	)
	if err != nil {
		return nil, fmt.Errorf("list stories: %w", err)
	}
	defer rows.Close()

	var stories []domain.Story
	for rows.Next() {
		story, err := scanStory(rows)
		if err != nil {
			return nil, fmt.Errorf("scan story: %w", err)
		}
		stories = append(stories, *story)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate stories: %w", err)
	}
	return stories, nil
}

// Update changes the description of an existing story.
// It returns domain.ErrNotFound when no matching story exists.
func (s *StoryStore) Update(epicID int, name, description string) (*domain.Story, error) {
	res, err := s.db.conn.Exec(
		"UPDATE stories SET description = ?, updated_at = datetime('now') WHERE epic_id = ? AND name = ?",
		description, epicID, name,
	)
	if err != nil {
		return nil, fmt.Errorf("update story: %w", err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("rows affected: %w", err)
	}
	if affected == 0 {
		return nil, fmt.Errorf("story %q (epic %d): %w", name, epicID, domain.ErrNotFound)
	}
	return s.GetByName(epicID, name)
}

// Delete removes a story by (epicID, name).
// It returns domain.ErrNotFound when no matching story exists.
func (s *StoryStore) Delete(epicID int, name string) error {
	res, err := s.db.conn.Exec(
		"DELETE FROM stories WHERE epic_id = ? AND name = ?",
		epicID, name,
	)
	if err != nil {
		return fmt.Errorf("delete story: %w", err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}
	if affected == 0 {
		return fmt.Errorf("story %q (epic %d): %w", name, epicID, domain.ErrNotFound)
	}
	return nil
}

// UpdateApprovalStatus sets the approval status for a story.
// It returns domain.ErrNotFound when no matching story exists.
func (s *StoryStore) UpdateApprovalStatus(epicID int, name string, status domain.ApprovalStatus) error {
	res, err := s.db.conn.Exec(
		"UPDATE stories SET approval_status = ?, updated_at = datetime('now') WHERE epic_id = ? AND name = ?",
		status, epicID, name,
	)
	if err != nil {
		return fmt.Errorf("update approval status: %w", err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}
	if affected == 0 {
		return fmt.Errorf("story %q (epic %d): %w", name, epicID, domain.ErrNotFound)
	}
	return nil
}
