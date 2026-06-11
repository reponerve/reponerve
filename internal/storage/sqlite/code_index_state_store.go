package sqlite

import (
	"context"
	"fmt"

	codemodels "github.com/reponerve/reponerve/internal/code/models"
)

// SQLiteCodeIndexStateStore implements storage.CodeIndexStateStore.
type SQLiteCodeIndexStateStore struct {
	db *Database
}

// NewSQLiteCodeIndexStateStore creates a new SQLiteCodeIndexStateStore.
func NewSQLiteCodeIndexStateStore(db *Database) *SQLiteCodeIndexStateStore {
	return &SQLiteCodeIndexStateStore{db: db}
}

// UpsertCodeIndexState inserts or updates code index state for a repository.
func (s *SQLiteCodeIndexStateStore) UpsertCodeIndexState(ctx context.Context, state *codemodels.CodeIndexState) error {
	query := `
		INSERT INTO code_index_state (
			repository_id, last_indexed_at, module_count, file_count,
			entity_count, relationship_count, link_count
		) VALUES (?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(repository_id) DO UPDATE SET
			last_indexed_at    = excluded.last_indexed_at,
			module_count       = excluded.module_count,
			file_count         = excluded.file_count,
			entity_count       = excluded.entity_count,
			relationship_count = excluded.relationship_count,
			link_count         = excluded.link_count
	`
	_, err := s.db.ExecContext(ctx, query,
		state.RepositoryID, state.LastIndexedAt.UTC().Format("2006-01-02T15:04:05Z"),
		state.ModuleCount, state.FileCount, state.EntityCount, state.RelationshipCount, state.LinkCount,
	)
	if err != nil {
		return fmt.Errorf("failed to upsert code index state: %w", err)
	}
	return nil
}
