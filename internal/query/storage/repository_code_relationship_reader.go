package storage

import (
	"context"
	"fmt"
	"time"

	codemodels "github.com/reponerve/reponerve/internal/code/models"
	"github.com/reponerve/reponerve/internal/storage/sqlite"
)

// RepositoryCodeRelationshipReader reads repository-code links.
type RepositoryCodeRelationshipReader interface {
	ListByRepositoryEntity(ctx context.Context, repositoryID, repositoryEntityID string) ([]*codemodels.RepositoryCodeRelationship, error)
	ListByCodeEntity(ctx context.Context, repositoryID, codeEntityID string) ([]*codemodels.RepositoryCodeRelationship, error)
	ListByRepository(ctx context.Context, repositoryID string) ([]*codemodels.RepositoryCodeRelationship, error)
}

// SQLiteRepositoryCodeRelationshipReader implements RepositoryCodeRelationshipReader.
type SQLiteRepositoryCodeRelationshipReader struct {
	db *sqlite.Database
}

// NewSQLiteRepositoryCodeRelationshipReader creates a repository-code relationship reader.
func NewSQLiteRepositoryCodeRelationshipReader(db *sqlite.Database) *SQLiteRepositoryCodeRelationshipReader {
	return &SQLiteRepositoryCodeRelationshipReader{db: db}
}

func scanRepositoryCodeRelationship(scanner interface {
	Scan(dest ...any) error
}) (*codemodels.RepositoryCodeRelationship, error) {
	var rel codemodels.RepositoryCodeRelationship
	var indexedAt string
	if err := scanner.Scan(
		&rel.ID, &rel.RepositoryID, &rel.RepositoryEntityID, &rel.RepositoryEntityType,
		&rel.CodeEntityID, &rel.CodeEntityType, &rel.RelationshipType, &rel.EvidenceJSON, &indexedAt,
	); err != nil {
		return nil, err
	}
	if t, err := time.Parse(time.RFC3339, indexedAt); err == nil {
		rel.IndexedAt = t
	}
	return &rel, nil
}

func (r *SQLiteRepositoryCodeRelationshipReader) queryLinks(ctx context.Context, query string, args ...any) ([]*codemodels.RepositoryCodeRelationship, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query repository-code relationships: %w", err)
	}
	defer rows.Close()

	var links []*codemodels.RepositoryCodeRelationship
	for rows.Next() {
		rel, err := scanRepositoryCodeRelationship(rows)
		if err != nil {
			return nil, fmt.Errorf("scan repository-code relationship: %w", err)
		}
		links = append(links, rel)
	}
	return links, rows.Err()
}

// ListByRepositoryEntity returns links for one repository memory entity.
func (r *SQLiteRepositoryCodeRelationshipReader) ListByRepositoryEntity(ctx context.Context, repositoryID, repositoryEntityID string) ([]*codemodels.RepositoryCodeRelationship, error) {
	query := `
		SELECT id, repository_id, repository_entity_id, repository_entity_type,
		       code_entity_id, code_entity_type, relationship_type, evidence_json, indexed_at
		FROM repository_code_relationships
		WHERE repository_id = ? AND repository_entity_id = ?
		ORDER BY relationship_type ASC, code_entity_type ASC, code_entity_id ASC, id ASC
	`
	return r.queryLinks(ctx, query, repositoryID, repositoryEntityID)
}

// ListByCodeEntity returns links for one code entity.
func (r *SQLiteRepositoryCodeRelationshipReader) ListByCodeEntity(ctx context.Context, repositoryID, codeEntityID string) ([]*codemodels.RepositoryCodeRelationship, error) {
	query := `
		SELECT id, repository_id, repository_entity_id, repository_entity_type,
		       code_entity_id, code_entity_type, relationship_type, evidence_json, indexed_at
		FROM repository_code_relationships
		WHERE repository_id = ? AND code_entity_id = ?
		ORDER BY relationship_type ASC, repository_entity_type ASC, repository_entity_id ASC, id ASC
	`
	return r.queryLinks(ctx, query, repositoryID, codeEntityID)
}

// ListByRepository returns all repository-code links.
func (r *SQLiteRepositoryCodeRelationshipReader) ListByRepository(ctx context.Context, repositoryID string) ([]*codemodels.RepositoryCodeRelationship, error) {
	query := `
		SELECT id, repository_id, repository_entity_id, repository_entity_type,
		       code_entity_id, code_entity_type, relationship_type, evidence_json, indexed_at
		FROM repository_code_relationships
		WHERE repository_id = ?
		ORDER BY relationship_type ASC, repository_entity_type ASC, repository_entity_id ASC,
		         code_entity_type ASC, code_entity_id ASC, id ASC
	`
	return r.queryLinks(ctx, query, repositoryID)
}
