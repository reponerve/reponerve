package sqlite

import (
	"context"
	"fmt"

	"github.com/reponerve/reponerve/pkg/models"
)

// RepositoryStore implements storage.RepositoryStore for SQLite.
type RepositoryStore struct {
	db *Database
}

// NewRepositoryStore creates a new SQLite RepositoryStore.
func NewRepositoryStore(db *Database) *RepositoryStore {
	return &RepositoryStore{db: db}
}

// UpsertRepository persists or updates the repository metadata.
func (s *RepositoryStore) UpsertRepository(ctx context.Context, repo *models.Repository) error {
	query := `
		INSERT INTO repositories (id, name, path, default_branch, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			name = excluded.name,
			path = excluded.path,
			default_branch = excluded.default_branch,
			updated_at = excluded.updated_at
	`
	_, err := s.db.ExecContext(ctx, query, repo.ID, repo.Name, repo.Path, repo.DefaultBranch, repo.CreatedAt, repo.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to store repository metadata in sqlite: %w", err)
	}
	return nil
}
