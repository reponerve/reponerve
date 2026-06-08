package storage

import (
	"context"
	"fmt"

	"github.com/reponerve/reponerve/internal/memory/models"
	"github.com/reponerve/reponerve/internal/storage/sqlite"
)

// SQLiteIntentStore implements storage.IntentStore for SQLite database.
type SQLiteIntentStore struct {
	db *sqlite.Database
}

// NewSQLiteIntentStore creates a new SQLiteIntentStore instance.
func NewSQLiteIntentStore(db *sqlite.Database) *SQLiteIntentStore {
	return &SQLiteIntentStore{db: db}
}

// UpsertIntent persists or updates an extracted Intent memory record.
func (s *SQLiteIntentStore) UpsertIntent(ctx context.Context, intent *models.Intent) error {
	query := `
		INSERT INTO memory_intents (id, repository_id, description, source_id, created_at)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			description = excluded.description
	`
	_, err := s.db.ExecContext(ctx, query,
		intent.ID,
		intent.RepositoryID,
		intent.Description,
		intent.SourceID,
		intent.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to upsert intent: %w", err)
	}
	return nil
}
