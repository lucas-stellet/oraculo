package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

// DB wraps an SQLite connection with auto-bootstrap behavior.
type DB struct {
	conn *sql.DB
}

// Open opens the database at .oraculo/oraculo.db relative to the current directory.
// It creates .oraculo/ if missing and runs pending migrations.
func Open() (*DB, error) {
	dir := ".oraculo"
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("create %s: %w", dir, err)
	}
	return openPath(filepath.Join(dir, "oraculo.db"))
}

// OpenMemory opens an in-memory SQLite database with all migrations applied.
// Intended for use in tests via the dbtest package.
func OpenMemory() (*DB, error) {
	return openPath(":memory:")
}

// openPath opens (or creates) a database at the given path and runs migrations.
func openPath(dsn string) (*DB, error) {
	conn, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	if _, err := conn.Exec("PRAGMA journal_mode=WAL"); err != nil {
		conn.Close()
		return nil, fmt.Errorf("set journal mode: %w", err)
	}
	if _, err := conn.Exec("PRAGMA foreign_keys=ON"); err != nil {
		conn.Close()
		return nil, fmt.Errorf("enable foreign keys: %w", err)
	}
	d := &DB{conn: conn}
	if err := d.migrate(); err != nil {
		conn.Close()
		return nil, err
	}
	return d, nil
}

// OpenMemory opens an in-memory SQLite database with all migrations applied.
// It is intended for use in tests.
func OpenMemory() (*DB, error) {
	return openPath(":memory:")
}

// Close closes the underlying database connection.
func (d *DB) Close() error {
	return d.conn.Close()
}

// Conn returns the underlying *sql.DB for use by stores.
func (d *DB) Conn() *sql.DB {
	return d.conn
}
