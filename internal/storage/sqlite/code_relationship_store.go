package sqlite

import (
	"context"
	"fmt"

	codemodels "github.com/reponerve/reponerve/internal/code/models"
)

// SQLiteCodeRelationshipStore implements storage.CodeRelationshipStore.
type SQLiteCodeRelationshipStore struct {
	db *Database
}

// NewSQLiteCodeRelationshipStore creates a new SQLiteCodeRelationshipStore.
func NewSQLiteCodeRelationshipStore(db *Database) *SQLiteCodeRelationshipStore {
	return &SQLiteCodeRelationshipStore{db: db}
}

// UpsertCodeRelationship inserts or updates a code relationship.
func (s *SQLiteCodeRelationshipStore) UpsertCodeRelationship(ctx context.Context, rel *codemodels.CodeRelationship) error {
	query := `
		INSERT INTO code_relationships (
			id, repository_id, from_entity_id, to_entity_id,
			relationship_type, evidence_json, indexed_at
		) VALUES (?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			repository_id     = excluded.repository_id,
			from_entity_id    = excluded.from_entity_id,
			to_entity_id      = excluded.to_entity_id,
			relationship_type = excluded.relationship_type,
			evidence_json     = excluded.evidence_json,
			indexed_at        = excluded.indexed_at
	`
	_, err := s.db.ExecContext(ctx, query,
		rel.ID, rel.RepositoryID, rel.FromEntityID, rel.ToEntityID,
		rel.RelationshipType, rel.EvidenceJSON, rel.IndexedAt.UTC().Format("2006-01-02T15:04:05Z"),
	)
	if err != nil {
		return fmt.Errorf("failed to upsert code relationship: %w", err)
	}
	return nil
}

// DeleteByRepository removes all code relationships for a repository.
func (s *SQLiteCodeRelationshipStore) DeleteByRepository(ctx context.Context, repositoryID string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM code_relationships WHERE repository_id = ?`, repositoryID)
	if err != nil {
		return fmt.Errorf("failed to delete code relationships: %w", err)
	}
	return nil
}
