// src/db/validation_store.go
package db

import (
	"fmt"
	"time"

	"github.com/lucas/oraculo/apps/backend/src/domain"
)

// ValidationStore provides operations for the validations table.
type ValidationStore struct {
	db *DB
}

// NewValidationStore returns a ValidationStore that uses the given database connection.
func NewValidationStore(db *DB) *ValidationStore {
	return &ValidationStore{db: db}
}

// Save inserts a new validation record and returns the created entity.
func (s *ValidationStore) Save(storyID int, verdict string, findings string) (*domain.Validation, error) {
	res, err := s.db.conn.Exec(
		"INSERT INTO validations (story_id, verdict, findings) VALUES (?, ?, ?)",
		storyID, verdict, findings,
	)
	if err != nil {
		return nil, fmt.Errorf("save validation: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("last insert id: %w", err)
	}

	row := s.db.conn.QueryRow(
		"SELECT id, story_id, task_id, verdict, findings, created_at FROM validations WHERE id = ?",
		id,
	)
	v, err := scanValidation(row)
	if err != nil {
		return nil, fmt.Errorf("fetch after save: %w", err)
	}
	return v, nil
}

// ListByStory returns all validations for the given storyID ordered by creation time.
func (s *ValidationStore) ListByStory(storyID int) ([]domain.Validation, error) {
	rows, err := s.db.conn.Query(
		"SELECT id, story_id, task_id, verdict, findings, created_at FROM validations WHERE story_id = ? ORDER BY created_at",
		storyID,
	)
	if err != nil {
		return nil, fmt.Errorf("list validations: %w", err)
	}
	defer rows.Close()

	var results []domain.Validation
	for rows.Next() {
		v, err := scanValidation(rows)
		if err != nil {
			return nil, fmt.Errorf("scan validation: %w", err)
		}
		results = append(results, *v)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate validations: %w", err)
	}
	if results == nil {
		results = []domain.Validation{}
	}
	return results, nil
}

// scanValidation reads a Validation from a row scanner.
func scanValidation(row interface{ Scan(...any) error }) (*domain.Validation, error) {
	var (
		v         domain.Validation
		createdAt string
	)
	if err := row.Scan(
		&v.ID, &v.StoryID, &v.TaskID, &v.Verdict, &v.Findings, &createdAt,
	); err != nil {
		return nil, err
	}
	var err error
	if v.CreatedAt, err = time.Parse(sqliteTimeFmt, createdAt); err != nil {
		return nil, fmt.Errorf("parse created_at: %w", err)
	}
	return &v, nil
}
