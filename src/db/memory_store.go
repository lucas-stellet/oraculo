// src/db/memory_store.go
package db

import (
	"fmt"
	"time"

	"github.com/lucas/oraculo/src/domain"
)

// MemoryStore provides operations for the knowledge table backed by SQLite with FTS5 search.
type MemoryStore struct {
	db *DB
}

// NewMemoryStore returns a MemoryStore that uses the given database connection.
func NewMemoryStore(db *DB) *MemoryStore {
	return &MemoryStore{db: db}
}

// Store inserts a new knowledge finding and returns the created entity.
func (s *MemoryStore) Store(domainName string, category domain.Category, finding, sourceFiles string, confidence domain.Confidence) (*domain.Knowledge, error) {
	res, err := s.db.conn.Exec(
		"INSERT INTO knowledge (domain, category, finding, source_files, confidence) VALUES (?, ?, ?, ?, ?)",
		domainName, category, finding, sourceFiles, confidence,
	)
	if err != nil {
		return nil, fmt.Errorf("store knowledge: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("last insert id: %w", err)
	}

	row := s.db.conn.QueryRow(
		"SELECT id, domain, category, finding, source_files, confidence, created_at FROM knowledge WHERE id = ?",
		id,
	)
	k, err := scanKnowledge(row)
	if err != nil {
		return nil, fmt.Errorf("fetch after store: %w", err)
	}
	return k, nil
}

// Search performs a full-text search over the knowledge table using FTS5.
// If domainFilter is non-empty, results are restricted to that domain.
// If limit is greater than zero, at most limit results are returned.
func (s *MemoryStore) Search(query, domainFilter string, limit int) ([]domain.Knowledge, error) {
	baseQuery := `SELECT k.id, k.domain, k.category, k.finding, k.source_files, k.confidence, k.created_at
		FROM knowledge k
		WHERE k.id IN (SELECT rowid FROM knowledge_fts WHERE knowledge_fts MATCH ?)`

	args := []any{query}

	if domainFilter != "" {
		baseQuery += " AND k.domain = ?"
		args = append(args, domainFilter)
	}

	if limit > 0 {
		baseQuery += " LIMIT ?"
		args = append(args, limit)
	}

	rows, err := s.db.conn.Query(baseQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("search knowledge: %w", err)
	}
	defer rows.Close()

	var results []domain.Knowledge
	for rows.Next() {
		k, err := scanKnowledge(rows)
		if err != nil {
			return nil, fmt.Errorf("scan knowledge: %w", err)
		}
		results = append(results, *k)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate knowledge: %w", err)
	}
	if results == nil {
		results = []domain.Knowledge{}
	}
	return results, nil
}

// Domains returns all distinct domain names from the knowledge table, sorted alphabetically.
func (s *MemoryStore) Domains() ([]string, error) {
	rows, err := s.db.conn.Query("SELECT DISTINCT domain FROM knowledge ORDER BY domain")
	if err != nil {
		return nil, fmt.Errorf("list domains: %w", err)
	}
	defer rows.Close()

	var domains []string
	for rows.Next() {
		var d string
		if err := rows.Scan(&d); err != nil {
			return nil, fmt.Errorf("scan domain: %w", err)
		}
		domains = append(domains, d)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate domains: %w", err)
	}
	return domains, nil
}

// scanKnowledge reads a Knowledge from a row scanner, parsing the SQLite TEXT timestamp.
func scanKnowledge(row interface{ Scan(...any) error }) (*domain.Knowledge, error) {
	var (
		k         domain.Knowledge
		createdAt string
	)
	if err := row.Scan(
		&k.ID, &k.Domain, &k.Category, &k.Finding,
		&k.SourceFiles, &k.Confidence, &createdAt,
	); err != nil {
		return nil, err
	}
	var err error
	if k.CreatedAt, err = time.Parse(sqliteTimeFmt, createdAt); err != nil {
		return nil, fmt.Errorf("parse created_at: %w", err)
	}
	return &k, nil
}
