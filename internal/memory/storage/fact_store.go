package storage

import (
	"context"
	"fmt"

	"github.com/reponerve/reponerve/internal/memory/models"
	"github.com/reponerve/reponerve/internal/storage/sqlite"
)

// SQLiteFactStore implements storage.FactStore for SQLite database.
type SQLiteFactStore struct {
	db *sqlite.Database
}

// NewSQLiteFactStore creates a new SQLiteFactStore instance.
func NewSQLiteFactStore(db *sqlite.Database) *SQLiteFactStore {
	return &SQLiteFactStore{db: db}
}

// UpsertFact persists or updates an extracted Fact memory record.
func (s *SQLiteFactStore) UpsertFact(ctx context.Context, fact *models.Fact) error {
	query := `
		INSERT INTO memory_facts (id, repository_id, subject, predicate, object, source_id, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			subject   = excluded.subject,
			predicate = excluded.predicate,
			object    = excluded.object
	`
	_, err := s.db.ExecContext(ctx, query,
		fact.ID,
		fact.RepositoryID,
		fact.Subject,
		fact.Predicate,
		fact.Object,
		fact.SourceID,
		fact.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to upsert fact: %w", err)
	}
	return nil
}

// DeleteFact removes a fact memory record by ID.
func (s *SQLiteFactStore) DeleteFact(ctx context.Context, id string) error {
	if _, err := s.db.ExecContext(ctx, `DELETE FROM memory_facts WHERE id = ?`, id); err != nil {
		return fmt.Errorf("failed to delete fact: %w", err)
	}
	return nil
}
