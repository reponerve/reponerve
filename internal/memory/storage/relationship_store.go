package storage

import (
	"context"
	"fmt"

	"github.com/reponerve/reponerve/internal/memory/models"
	"github.com/reponerve/reponerve/internal/storage/sqlite"
)

// SQLiteRelationshipStore implements storage.RelationshipStore for SQLite database.
type SQLiteRelationshipStore struct {
	db *sqlite.Database
}

// NewSQLiteRelationshipStore creates a new SQLiteRelationshipStore instance.
func NewSQLiteRelationshipStore(db *sqlite.Database) *SQLiteRelationshipStore {
	return &SQLiteRelationshipStore{db: db}
}

// UpsertRelationship persists or updates a Relationship memory record.
func (s *SQLiteRelationshipStore) UpsertRelationship(ctx context.Context, rel *models.Relationship) error {
	query := `
		INSERT INTO memory_relationships (id, repository_id, from_id, to_id, relationship_type, created_at)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO NOTHING
	`
	_, err := s.db.ExecContext(ctx, query,
		rel.ID,
		rel.RepositoryID,
		rel.FromID,
		rel.ToID,
		rel.Type,
		rel.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to upsert relationship: %w", err)
	}
	return nil
}
