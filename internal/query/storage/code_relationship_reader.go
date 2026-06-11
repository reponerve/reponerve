package storage

import (
	"context"
	"fmt"
	"time"

	codemodels "github.com/reponerve/reponerve/internal/code/models"
	"github.com/reponerve/reponerve/internal/storage/sqlite"
)

// CodeRelationshipReader reads code relationships.
type CodeRelationshipReader interface {
	ListByFromEntity(ctx context.Context, entityID string) ([]*codemodels.CodeRelationship, error)
	ListByToEntity(ctx context.Context, entityID string) ([]*codemodels.CodeRelationship, error)
	ListByRepository(ctx context.Context, repositoryID string) ([]*codemodels.CodeRelationship, error)
}

// SQLiteCodeRelationshipReader implements CodeRelationshipReader.
type SQLiteCodeRelationshipReader struct {
	db *sqlite.Database
}

// NewSQLiteCodeRelationshipReader creates a code relationship reader.
func NewSQLiteCodeRelationshipReader(db *sqlite.Database) *SQLiteCodeRelationshipReader {
	return &SQLiteCodeRelationshipReader{db: db}
}

func scanCodeRelationship(scanner interface {
	Scan(dest ...any) error
}) (*codemodels.CodeRelationship, error) {
	var rel codemodels.CodeRelationship
	var indexedAt string
	if err := scanner.Scan(
		&rel.ID, &rel.RepositoryID, &rel.FromEntityID, &rel.ToEntityID,
		&rel.RelationshipType, &rel.EvidenceJSON, &indexedAt,
	); err != nil {
		return nil, err
	}
	if t, err := time.Parse(time.RFC3339, indexedAt); err == nil {
		rel.IndexedAt = t
	}
	return &rel, nil
}

func (r *SQLiteCodeRelationshipReader) queryRelationships(ctx context.Context, query string, args ...any) ([]*codemodels.CodeRelationship, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query code relationships: %w", err)
	}
	defer rows.Close()

	var rels []*codemodels.CodeRelationship
	for rows.Next() {
		rel, err := scanCodeRelationship(rows)
		if err != nil {
			return nil, fmt.Errorf("scan code relationship: %w", err)
		}
		rels = append(rels, rel)
	}
	return rels, rows.Err()
}

// ListByFromEntity returns outbound relationships from an entity.
func (r *SQLiteCodeRelationshipReader) ListByFromEntity(ctx context.Context, entityID string) ([]*codemodels.CodeRelationship, error) {
	query := `
		SELECT id, repository_id, from_entity_id, to_entity_id, relationship_type, evidence_json, indexed_at
		FROM code_relationships
		WHERE from_entity_id = ?
		ORDER BY relationship_type ASC, to_entity_id ASC, id ASC
	`
	return r.queryRelationships(ctx, query, entityID)
}

// ListByToEntity returns inbound relationships to an entity.
func (r *SQLiteCodeRelationshipReader) ListByToEntity(ctx context.Context, entityID string) ([]*codemodels.CodeRelationship, error) {
	query := `
		SELECT id, repository_id, from_entity_id, to_entity_id, relationship_type, evidence_json, indexed_at
		FROM code_relationships
		WHERE to_entity_id = ?
		ORDER BY relationship_type ASC, from_entity_id ASC, id ASC
	`
	return r.queryRelationships(ctx, query, entityID)
}

// ListByRepository returns all code relationships for a repository.
func (r *SQLiteCodeRelationshipReader) ListByRepository(ctx context.Context, repositoryID string) ([]*codemodels.CodeRelationship, error) {
	query := `
		SELECT id, repository_id, from_entity_id, to_entity_id, relationship_type, evidence_json, indexed_at
		FROM code_relationships
		WHERE repository_id = ?
		ORDER BY relationship_type ASC, from_entity_id ASC, to_entity_id ASC, id ASC
	`
	return r.queryRelationships(ctx, query, repositoryID)
}
