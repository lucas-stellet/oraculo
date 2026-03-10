package db

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/lucas/oraculo/apps/backend/src/domain"
)

// ReviewStore provides CRUD operations for reviews backed by SQLite.
type ReviewStore struct {
	db *DB
}

// NewReviewStore returns a ReviewStore that uses the given database connection.
func NewReviewStore(db *DB) *ReviewStore {
	return &ReviewStore{db: db}
}

// scanReview reads a Review from a row scanner.
func scanReview(row interface{ Scan(...any) error }) (*domain.Review, error) {
	var (
		r         domain.Review
		createdAt string
	)
	if err := row.Scan(&r.ID, &r.VersionID, &r.VersionType, &r.Verdict, &r.Comment, &createdAt); err != nil {
		return nil, err
	}
	var err error
	if r.CreatedAt, err = time.Parse(sqliteTimeFmt, createdAt); err != nil {
		return nil, fmt.Errorf("parse created_at: %w", err)
	}
	return &r, nil
}

// Create inserts a new review and propagates the verdict to the parent entity's approval_status.
func (s *ReviewStore) Create(versionID int, versionType domain.VersionType, verdict domain.ReviewVerdict, comment string) (*domain.Review, error) {
	tx, err := s.db.conn.Begin()
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	res, err := tx.Exec(
		"INSERT INTO reviews (version_id, version_type, verdict, comment) VALUES (?, ?, ?, ?)",
		versionID, versionType, verdict, comment,
	)
	if err != nil {
		return nil, fmt.Errorf("insert review: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("last insert id: %w", err)
	}

	// Map review verdict to approval status.
	var status domain.ApprovalStatus
	switch verdict {
	case domain.ReviewApproved:
		status = domain.ApprovalApproved
	case domain.ReviewRejected:
		status = domain.ApprovalRejected
	}

	// Propagate status to parent entity.
	switch versionType {
	case domain.VersionTypeEpic:
		var epicID int
		err = tx.QueryRow("SELECT epic_id FROM epic_versions WHERE id = ?", versionID).Scan(&epicID)
		if err != nil {
			return nil, fmt.Errorf("get epic_id from version: %w", err)
		}
		_, err = tx.Exec(
			"UPDATE epics SET approval_status = ?, updated_at = datetime('now') WHERE id = ?",
			status, epicID,
		)
		if err != nil {
			return nil, fmt.Errorf("propagate epic approval status: %w", err)
		}
	case domain.VersionTypeStory:
		var storyID int
		err = tx.QueryRow("SELECT story_id FROM story_versions WHERE id = ?", versionID).Scan(&storyID)
		if err != nil {
			return nil, fmt.Errorf("get story_id from version: %w", err)
		}
		_, err = tx.Exec(
			"UPDATE stories SET approval_status = ?, updated_at = datetime('now') WHERE id = ?",
			status, storyID,
		)
		if err != nil {
			return nil, fmt.Errorf("propagate story approval status: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit tx: %w", err)
	}

	return s.GetByID(int(id))
}

// GetByID retrieves a review by its ID.
func (s *ReviewStore) GetByID(id int) (*domain.Review, error) {
	row := s.db.conn.QueryRow(
		"SELECT id, version_id, version_type, verdict, comment, created_at FROM reviews WHERE id = ?", id,
	)
	r, err := scanReview(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("review %d: %w", id, domain.ErrNotFound)
		}
		return nil, fmt.Errorf("get review: %w", err)
	}
	return r, nil
}

// ListByVersion returns all reviews for a given version ordered by created_at.
func (s *ReviewStore) ListByVersion(versionID int, versionType domain.VersionType) ([]domain.Review, error) {
	rows, err := s.db.conn.Query(
		"SELECT id, version_id, version_type, verdict, comment, created_at FROM reviews WHERE version_id = ? AND version_type = ? ORDER BY created_at",
		versionID, versionType,
	)
	if err != nil {
		return nil, fmt.Errorf("list reviews: %w", err)
	}
	defer rows.Close()

	var reviews []domain.Review
	for rows.Next() {
		r, err := scanReview(rows)
		if err != nil {
			return nil, fmt.Errorf("scan review: %w", err)
		}
		reviews = append(reviews, *r)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate reviews: %w", err)
	}
	return reviews, nil
}

// ListByStory returns all reviews for all versions of a story, ordered by created_at.
func (s *ReviewStore) ListByStory(storyID int) ([]domain.Review, error) {
	query := `
		SELECT r.id, r.version_id, r.version_type, r.verdict, r.comment, r.created_at
		FROM reviews r
		JOIN story_versions sv ON r.version_id = sv.id AND r.version_type = 'story'
		WHERE sv.story_id = ?
		ORDER BY r.created_at
	`
	rows, err := s.db.conn.Query(query, storyID)
	if err != nil {
		return nil, fmt.Errorf("list reviews by story: %w", err)
	}
	defer rows.Close()

	var reviews []domain.Review
	for rows.Next() {
		review, err := scanReview(rows)
		if err != nil {
			return nil, fmt.Errorf("scan review: %w", err)
		}
		reviews = append(reviews, *review)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate reviews: %w", err)
	}
	if reviews == nil {
		reviews = []domain.Review{}
	}
	return reviews, nil
}
