package sqlite

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

// Database wraps sql.DB to provide a connection abstraction.
type Database struct {
	*sql.DB
}

// Open opens a connection to the SQLite database.
// It ensures parent directories exist, and enables foreign keys and WAL mode.
func Open(dbPath string) (*Database, error) {
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory for database: %w", err)
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	db.SetMaxOpenConns(1)

	_, err = db.Exec("PRAGMA foreign_keys = ON; PRAGMA journal_mode = WAL; PRAGMA busy_timeout = 5000;")
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to configure database pragmas: %w", err)
	}

	return &Database{DB: db}, nil
}
