package sqlite

import (
	"context"
	"fmt"

	"reponerve/pkg/models"
)

// SQLiteContributorStore implements storage.ContributorStore for SQLite.
type SQLiteContributorStore struct {
	db *Database
}

// NewSQLiteContributorStore creates a new SQLiteContributorStore.
func NewSQLiteContributorStore(db *Database) *SQLiteContributorStore {
	return &SQLiteContributorStore{db: db}
}

// UpsertContributor inserts or updates a contributor record.
func (s *SQLiteContributorStore) UpsertContributor(ctx context.Context, c *models.Contributor) error {
	query := `
		INSERT INTO contributors (id, repository_id, name, email, first_seen, last_seen, commit_count)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			repository_id = excluded.repository_id,
			name          = excluded.name,
			email         = excluded.email,
			first_seen    = excluded.first_seen,
			last_seen     = excluded.last_seen,
			commit_count  = excluded.commit_count
	`
	_, err := s.db.ExecContext(ctx, query,
		c.ID,
		c.RepositoryID,
		c.Name,
		c.Email,
		c.FirstSeen,
		c.LastSeen,
		c.CommitCount,
	)
	if err != nil {
		return fmt.Errorf("failed to upsert contributor in sqlite: %w", err)
	}
	return nil
}
