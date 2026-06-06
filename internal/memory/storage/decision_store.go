package storage

import (
	"context"
	"fmt"

	"reponerve/internal/memory/models"
	"reponerve/internal/storage/sqlite"
)

// SQLiteDecisionStore implements storage.DecisionStore for SQLite database.
type SQLiteDecisionStore struct {
	db *sqlite.Database
}

// NewSQLiteDecisionStore creates a new SQLiteDecisionStore instance.
func NewSQLiteDecisionStore(db *sqlite.Database) *SQLiteDecisionStore {
	return &SQLiteDecisionStore{db: db}
}

// UpsertDecision persists or updates an extracted Decision memory record.
func (s *SQLiteDecisionStore) UpsertDecision(ctx context.Context, decision *models.Decision) error {
	query := `
		INSERT INTO memory_decisions (id, repository_id, title, status, source_id, created_at)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			title  = excluded.title,
			status = excluded.status
	`
	_, err := s.db.ExecContext(ctx, query,
		decision.ID,
		decision.RepositoryID,
		decision.Title,
		decision.Status,
		decision.SourceID,
		decision.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to upsert decision: %w", err)
	}
	return nil
}
