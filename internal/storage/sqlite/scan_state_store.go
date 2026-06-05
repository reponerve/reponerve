package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"reponerve/internal/storage"
)

// ScanStateStore implements storage.ScanStateStore for SQLite.
type ScanStateStore struct {
	db *Database
}

// NewScanStateStore creates a new SQLite ScanStateStore.
func NewScanStateStore(db *Database) *ScanStateStore {
	return &ScanStateStore{db: db}
}

// GetScanState retrieves the scan state for a repository.
func (s *ScanStateStore) GetScanState(ctx context.Context, repoID string) (*storage.ScanState, error) {
	var state storage.ScanState
	query := "SELECT repository_id, last_scan_commit, updated_at FROM scan_state WHERE repository_id = ?"
	err := s.db.QueryRowContext(ctx, query, repoID).Scan(&state.RepositoryID, &state.LastScanCommit, &state.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf("failed to query scan state: %w", err)
	}
	return &state, nil
}

// UpdateScanState stores or updates the scan state for a repository.
func (s *ScanStateStore) UpdateScanState(ctx context.Context, repoID string, commitHash string) error {
	query := `
		INSERT INTO scan_state (repository_id, last_scan_commit, updated_at)
		VALUES (?, ?, ?)
		ON CONFLICT(repository_id) DO UPDATE SET
			last_scan_commit = excluded.last_scan_commit,
			updated_at = excluded.updated_at
	`
	_, err := s.db.ExecContext(ctx, query, repoID, commitHash, time.Now())
	if err != nil {
		return fmt.Errorf("failed to update scan state: %w", err)
	}
	return nil
}
