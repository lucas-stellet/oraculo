package db

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/lucas/oraculo/src/domain"
)

// VersionStore provides CRUD operations for epic and story versions backed by SQLite.
type VersionStore struct {
	db *DB
}

// NewVersionStore returns a VersionStore that uses the given database connection.
func NewVersionStore(db *DB) *VersionStore {
	return &VersionStore{db: db}
}

// scanEpicVersion reads an EpicVersion from a row scanner.
func scanEpicVersion(row interface{ Scan(...any) error }) (*domain.EpicVersion, error) {
	var (
		v         domain.EpicVersion
		createdAt string
	)
	if err := row.Scan(&v.ID, &v.EpicID, &v.Number, &v.Content, &createdAt); err != nil {
		return nil, err
	}
	var err error
	if v.CreatedAt, err = time.Parse(sqliteTimeFmt, createdAt); err != nil {
		return nil, fmt.Errorf("parse created_at: %w", err)
	}
	return &v, nil
}

// scanStoryVersion reads a StoryVersion from a row scanner.
func scanStoryVersion(row interface{ Scan(...any) error }) (*domain.StoryVersion, error) {
	var (
		v         domain.StoryVersion
		createdAt string
	)
	if err := row.Scan(&v.ID, &v.StoryID, &v.Number, &v.Content, &createdAt); err != nil {
		return nil, err
	}
	var err error
	if v.CreatedAt, err = time.Parse(sqliteTimeFmt, createdAt); err != nil {
		return nil, fmt.Errorf("parse created_at: %w", err)
	}
	return &v, nil
}

// CreateEpicVersion creates a new version for an epic, auto-incrementing the version number.
// It also sets the epic's approval_status to 'pending'.
func (s *VersionStore) CreateEpicVersion(epicID int, content string) (*domain.EpicVersion, error) {
	tx, err := s.db.conn.Begin()
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	var maxNum sql.NullInt64
	err = tx.QueryRow("SELECT MAX(number) FROM epic_versions WHERE epic_id = ?", epicID).Scan(&maxNum)
	if err != nil {
		return nil, fmt.Errorf("get max version number: %w", err)
	}
	nextNum := 1
	if maxNum.Valid {
		nextNum = int(maxNum.Int64) + 1
	}

	res, err := tx.Exec(
		"INSERT INTO epic_versions (epic_id, number, content) VALUES (?, ?, ?)",
		epicID, nextNum, content,
	)
	if err != nil {
		return nil, fmt.Errorf("insert epic version: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("last insert id: %w", err)
	}

	_, err = tx.Exec(
		"UPDATE epics SET approval_status = 'pending', updated_at = datetime('now') WHERE id = ?",
		epicID,
	)
	if err != nil {
		return nil, fmt.Errorf("update epic approval status: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit tx: %w", err)
	}

	return s.GetEpicVersion(int(id))
}

// GetEpicVersion retrieves an epic version by its ID.
func (s *VersionStore) GetEpicVersion(id int) (*domain.EpicVersion, error) {
	row := s.db.conn.QueryRow(
		"SELECT id, epic_id, number, content, created_at FROM epic_versions WHERE id = ?", id,
	)
	v, err := scanEpicVersion(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("epic version %d: %w", id, domain.ErrNotFound)
		}
		return nil, fmt.Errorf("get epic version: %w", err)
	}
	return v, nil
}

// ListEpicVersions returns all versions for an epic ordered by number.
func (s *VersionStore) ListEpicVersions(epicID int) ([]domain.EpicVersion, error) {
	rows, err := s.db.conn.Query(
		"SELECT id, epic_id, number, content, created_at FROM epic_versions WHERE epic_id = ? ORDER BY number",
		epicID,
	)
	if err != nil {
		return nil, fmt.Errorf("list epic versions: %w", err)
	}
	defer rows.Close()

	var versions []domain.EpicVersion
	for rows.Next() {
		v, err := scanEpicVersion(rows)
		if err != nil {
			return nil, fmt.Errorf("scan epic version: %w", err)
		}
		versions = append(versions, *v)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate epic versions: %w", err)
	}
	return versions, nil
}

// LatestEpicVersion returns the version with the highest number for an epic.
func (s *VersionStore) LatestEpicVersion(epicID int) (*domain.EpicVersion, error) {
	row := s.db.conn.QueryRow(
		"SELECT id, epic_id, number, content, created_at FROM epic_versions WHERE epic_id = ? ORDER BY number DESC LIMIT 1",
		epicID,
	)
	v, err := scanEpicVersion(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("epic %d latest version: %w", epicID, domain.ErrNotFound)
		}
		return nil, fmt.Errorf("latest epic version: %w", err)
	}
	return v, nil
}

// CreateStoryVersion creates a new version for a story, auto-incrementing the version number.
// It also sets the story's approval_status to 'pending'.
func (s *VersionStore) CreateStoryVersion(storyID int, content string) (*domain.StoryVersion, error) {
	tx, err := s.db.conn.Begin()
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	var maxNum sql.NullInt64
	err = tx.QueryRow("SELECT MAX(number) FROM story_versions WHERE story_id = ?", storyID).Scan(&maxNum)
	if err != nil {
		return nil, fmt.Errorf("get max version number: %w", err)
	}
	nextNum := 1
	if maxNum.Valid {
		nextNum = int(maxNum.Int64) + 1
	}

	res, err := tx.Exec(
		"INSERT INTO story_versions (story_id, number, content) VALUES (?, ?, ?)",
		storyID, nextNum, content,
	)
	if err != nil {
		return nil, fmt.Errorf("insert story version: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("last insert id: %w", err)
	}

	_, err = tx.Exec(
		"UPDATE stories SET approval_status = 'pending', updated_at = datetime('now') WHERE id = ?",
		storyID,
	)
	if err != nil {
		return nil, fmt.Errorf("update story approval status: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit tx: %w", err)
	}

	return s.GetStoryVersion(int(id))
}

// GetStoryVersion retrieves a story version by its ID.
func (s *VersionStore) GetStoryVersion(id int) (*domain.StoryVersion, error) {
	row := s.db.conn.QueryRow(
		"SELECT id, story_id, number, content, created_at FROM story_versions WHERE id = ?", id,
	)
	v, err := scanStoryVersion(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("story version %d: %w", id, domain.ErrNotFound)
		}
		return nil, fmt.Errorf("get story version: %w", err)
	}
	return v, nil
}

// ListStoryVersions returns all versions for a story ordered by number.
func (s *VersionStore) ListStoryVersions(storyID int) ([]domain.StoryVersion, error) {
	rows, err := s.db.conn.Query(
		"SELECT id, story_id, number, content, created_at FROM story_versions WHERE story_id = ? ORDER BY number",
		storyID,
	)
	if err != nil {
		return nil, fmt.Errorf("list story versions: %w", err)
	}
	defer rows.Close()

	var versions []domain.StoryVersion
	for rows.Next() {
		v, err := scanStoryVersion(rows)
		if err != nil {
			return nil, fmt.Errorf("scan story version: %w", err)
		}
		versions = append(versions, *v)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate story versions: %w", err)
	}
	return versions, nil
}

// LatestStoryVersion returns the version with the highest number for a story.
func (s *VersionStore) LatestStoryVersion(storyID int) (*domain.StoryVersion, error) {
	row := s.db.conn.QueryRow(
		"SELECT id, story_id, number, content, created_at FROM story_versions WHERE story_id = ? ORDER BY number DESC LIMIT 1",
		storyID,
	)
	v, err := scanStoryVersion(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("story %d latest version: %w", storyID, domain.ErrNotFound)
		}
		return nil, fmt.Errorf("latest story version: %w", err)
	}
	return v, nil
}
