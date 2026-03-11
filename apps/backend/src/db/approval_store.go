package db

import (
	"crypto/rand"
	"database/sql"
	"fmt"
	"time"

	"github.com/lucas/oraculo/apps/backend/src/domain"
)

// ApprovalStore provides CRUD operations for approvals backed by SQLite.
type ApprovalStore struct {
	db *DB
}

// NewApprovalStore returns an ApprovalStore that uses the given database connection.
func NewApprovalStore(db *DB) *ApprovalStore {
	return &ApprovalStore{db: db}
}

// scanApproval reads an Approval from a row scanner, handling nullable fields.
func scanApproval(row interface{ Scan(...any) error }) (*domain.Approval, error) {
	var (
		a           domain.Approval
		epicID      sql.NullInt64
		storyID     sql.NullInt64
		decidedAt   sql.NullString
		requestedAt string
	)
	if err := row.Scan(
		&a.ID, &a.Type, &epicID, &storyID,
		&a.Content, &a.PreviousVersion, &a.Status,
		&a.VerdictComment, &requestedAt, &decidedAt,
	); err != nil {
		return nil, err
	}
	var err error
	if a.RequestedAt, err = time.Parse(sqliteTimeFmt, requestedAt); err != nil {
		return nil, fmt.Errorf("parse requested_at: %w", err)
	}
	if epicID.Valid {
		v := int(epicID.Int64)
		a.EpicID = &v
	}
	if storyID.Valid {
		v := int(storyID.Int64)
		a.StoryID = &v
	}
	if decidedAt.Valid {
		t, err := time.Parse(sqliteTimeFmt, decidedAt.String)
		if err != nil {
			return nil, fmt.Errorf("parse decided_at: %w", err)
		}
		a.DecidedAt = &t
	}
	return &a, nil
}

// Request creates a new approval with status "pending" and an auto-generated UUID.
func (s *ApprovalStore) Request(approvalType domain.ApprovalType, epicID, storyID *int, content string) (*domain.Approval, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return nil, fmt.Errorf("generate uuid: %w", err)
	}
	id := fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])

	var epicVal, storyVal any
	if epicID != nil {
		epicVal = *epicID
	}
	if storyID != nil {
		storyVal = *storyID
	}

	_, err := s.db.conn.Exec(
		`INSERT INTO approvals (id, type, epic_id, story_id, content, status)
		 VALUES (?, ?, ?, ?, ?, 'pending')`,
		id, approvalType, epicVal, storyVal, content,
	)
	if err != nil {
		return nil, fmt.Errorf("create approval: %w", err)
	}
	return s.GetByID(id)
}

// GetByID retrieves an approval by its UUID.
// It returns domain.ErrNotFound when no matching approval exists.
func (s *ApprovalStore) GetByID(id string) (*domain.Approval, error) {
	row := s.db.conn.QueryRow(
		`SELECT id, type, epic_id, story_id, content, previous_version, status,
		        verdict_comment, requested_at, decided_at
		 FROM approvals WHERE id = ?`, id,
	)
	approval, err := scanApproval(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("approval %q: %w", id, domain.ErrNotFound)
		}
		return nil, fmt.Errorf("get approval: %w", err)
	}
	return approval, nil
}

// List returns all approvals. If pendingOnly is true, only pending approvals are returned.
func (s *ApprovalStore) List(pendingOnly bool) ([]domain.Approval, error) {
	query := `SELECT id, type, epic_id, story_id, content, previous_version, status,
	                 verdict_comment, requested_at, decided_at
	          FROM approvals`
	if pendingOnly {
		query += " WHERE status = 'pending'"
	}
	query += " ORDER BY requested_at"

	rows, err := s.db.conn.Query(query)
	if err != nil {
		return nil, fmt.Errorf("list approvals: %w", err)
	}
	defer rows.Close()

	var approvals []domain.Approval
	for rows.Next() {
		approval, err := scanApproval(rows)
		if err != nil {
			return nil, fmt.Errorf("scan approval: %w", err)
		}
		approvals = append(approvals, *approval)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate approvals: %w", err)
	}
	return approvals, nil
}

// ListByEpic returns approvals for a specific epic, optionally filtered to pending only.
func (s *ApprovalStore) ListByEpic(epicID int, pendingOnly bool) ([]domain.Approval, error) {
	query := `SELECT id, type, epic_id, story_id, content, previous_version, status,
	                 verdict_comment, requested_at, decided_at
	          FROM approvals WHERE epic_id = ?`
	if pendingOnly {
		query += " AND status = 'pending'"
	}
	query += " ORDER BY requested_at"

	rows, err := s.db.conn.Query(query, epicID)
	if err != nil {
		return nil, fmt.Errorf("list approvals by epic: %w", err)
	}
	defer rows.Close()

	var approvals []domain.Approval
	for rows.Next() {
		approval, err := scanApproval(rows)
		if err != nil {
			return nil, fmt.Errorf("scan approval: %w", err)
		}
		approvals = append(approvals, *approval)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate approvals: %w", err)
	}
	if approvals == nil {
		approvals = []domain.Approval{}
	}
	return approvals, nil
}

// Verdict records a decision on a pending approval.
// It returns domain.ErrInvalidTransition if the approval is not pending.
// If verdict is "needs_revision", the current content is copied to previous_version.
func (s *ApprovalStore) Verdict(id string, verdict domain.Verdict, comment string) (*domain.Approval, error) {
	current, err := s.GetByID(id)
	if err != nil {
		return nil, err
	}
	if current.Status != domain.ApprovalPending {
		return nil, fmt.Errorf("approval %q status %q: %w", id, current.Status, domain.ErrInvalidTransition)
	}

	// Map verdict to the corresponding approval status.
	var status domain.ApprovalStatus
	switch verdict {
	case domain.VerdictApproved:
		status = domain.ApprovalApproved
	case domain.VerdictRejected:
		status = domain.ApprovalRejected
	case domain.VerdictNeedsRevision:
		status = domain.ApprovalNeedsRevision
	}

	// If needs_revision, copy current content to previous_version.
	previousVersion := current.PreviousVersion
	if verdict == domain.VerdictNeedsRevision {
		previousVersion = current.Content
	}

	tx, err := s.db.conn.Begin()
	if err != nil {
		return nil, fmt.Errorf("begin verdict tx: %w", err)
	}
	defer tx.Rollback()

	_, err = tx.Exec(
		`UPDATE approvals
		 SET status = ?, verdict_comment = ?, previous_version = ?, decided_at = datetime('now')
		 WHERE id = ?`,
		status, comment, previousVersion, id,
	)
	if err != nil {
		return nil, fmt.Errorf("verdict approval: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit verdict tx: %w", err)
	}
	return s.GetByID(id)
}

// CreateComment inserts a new inline comment on an approval and returns the created record.
func (s *ApprovalStore) CreateComment(approvalID string, selectedText string, comment string) (*domain.ApprovalComment, error) {
	res, err := s.db.conn.Exec(
		`INSERT INTO approval_comments (approval_id, selected_text, comment) VALUES (?, ?, ?)`,
		approvalID, selectedText, comment,
	)
	if err != nil {
		return nil, fmt.Errorf("create approval comment: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("last insert id: %w", err)
	}

	var c domain.ApprovalComment
	var createdAt string
	err = s.db.conn.QueryRow(
		`SELECT id, selected_text, comment, created_at FROM approval_comments WHERE id = ?`, id,
	).Scan(&c.ID, &c.SelectedText, &c.Comment, &createdAt)
	if err != nil {
		return nil, fmt.Errorf("get created comment: %w", err)
	}
	if c.CreatedAt, err = time.Parse(sqliteTimeFmt, createdAt); err != nil {
		return nil, fmt.Errorf("parse comment created_at: %w", err)
	}
	return &c, nil
}

// ListComments returns all comments for an approval, ordered by created_at ascending.
func (s *ApprovalStore) ListComments(approvalID string) ([]domain.ApprovalComment, error) {
	rows, err := s.db.conn.Query(
		`SELECT id, selected_text, comment, created_at FROM approval_comments
		 WHERE approval_id = ? ORDER BY created_at`,
		approvalID,
	)
	if err != nil {
		return nil, fmt.Errorf("list approval comments: %w", err)
	}
	defer rows.Close()

	var comments []domain.ApprovalComment
	for rows.Next() {
		var c domain.ApprovalComment
		var createdAt string
		if err := rows.Scan(&c.ID, &c.SelectedText, &c.Comment, &createdAt); err != nil {
			return nil, fmt.Errorf("scan approval comment: %w", err)
		}
		if c.CreatedAt, err = time.Parse(sqliteTimeFmt, createdAt); err != nil {
			return nil, fmt.Errorf("parse comment created_at: %w", err)
		}
		comments = append(comments, c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate approval comments: %w", err)
	}
	if comments == nil {
		comments = []domain.ApprovalComment{}
	}
	return comments, nil
}

// DeleteComment deletes a single comment by its ID.
func (s *ApprovalStore) DeleteComment(id int) error {
	_, err := s.db.conn.Exec(`DELETE FROM approval_comments WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete approval comment %d: %w", id, err)
	}
	return nil
}

// DeleteCommentsByApproval deletes all comments for a given approval.
// Intended for use when an approval is approved and comments are no longer relevant.
func (s *ApprovalStore) DeleteCommentsByApproval(approvalID string) error {
	_, err := s.db.conn.Exec(`DELETE FROM approval_comments WHERE approval_id = ?`, approvalID)
	if err != nil {
		return fmt.Errorf("delete comments for approval %q: %w", approvalID, err)
	}
	return nil
}
