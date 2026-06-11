package sqlite

import (
	"context"
	"fmt"

	codemodels "github.com/reponerve/reponerve/internal/code/models"
)

// SQLiteRepositoryCodeRelationshipStore implements storage.RepositoryCodeRelationshipStore.
type SQLiteRepositoryCodeRelationshipStore struct {
	db *Database
}

// NewSQLiteRepositoryCodeRelationshipStore creates a new store.
func NewSQLiteRepositoryCodeRelationshipStore(db *Database) *SQLiteRepositoryCodeRelationshipStore {
	return &SQLiteRepositoryCodeRelationshipStore{db: db}
}

// UpsertRepositoryCodeRelationship inserts or updates a repository-code link.
func (s *SQLiteRepositoryCodeRelationshipStore) UpsertRepositoryCodeRelationship(ctx context.Context, rel *codemodels.RepositoryCodeRelationship) error {
	query := `
		INSERT INTO repository_code_relationships (
			id, repository_id, repository_entity_id, repository_entity_type,
			code_entity_id, code_entity_type, relationship_type, evidence_json, indexed_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			repository_id          = excluded.repository_id,
			repository_entity_id   = excluded.repository_entity_id,
			repository_entity_type = excluded.repository_entity_type,
			code_entity_id         = excluded.code_entity_id,
			code_entity_type       = excluded.code_entity_type,
			relationship_type      = excluded.relationship_type,
			evidence_json          = excluded.evidence_json,
			indexed_at             = excluded.indexed_at
	`
	_, err := s.db.ExecContext(ctx, query,
		rel.ID, rel.RepositoryID, rel.RepositoryEntityID, rel.RepositoryEntityType,
		rel.CodeEntityID, rel.CodeEntityType, rel.RelationshipType, rel.EvidenceJSON,
		rel.IndexedAt.UTC().Format("2006-01-02T15:04:05Z"),
	)
	if err != nil {
		return fmt.Errorf("failed to upsert repository-code relationship: %w", err)
	}
	return nil
}

// DeleteByRepository removes all repository-code links for a repository.
func (s *SQLiteRepositoryCodeRelationshipStore) DeleteByRepository(ctx context.Context, repositoryID string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM repository_code_relationships WHERE repository_id = ?`, repositoryID)
	if err != nil {
		return fmt.Errorf("failed to delete repository-code relationships: %w", err)
	}
	return nil
}
