package sqlite

import (
	"context"
	"fmt"

	codemodels "github.com/reponerve/reponerve/internal/code/models"
)

// SQLiteCodeEntityStore implements storage.CodeEntityStore.
type SQLiteCodeEntityStore struct {
	db *Database
}

// NewSQLiteCodeEntityStore creates a new SQLiteCodeEntityStore.
func NewSQLiteCodeEntityStore(db *Database) *SQLiteCodeEntityStore {
	return &SQLiteCodeEntityStore{db: db}
}

// UpsertCodeEntity inserts or updates a code entity.
func (s *SQLiteCodeEntityStore) UpsertCodeEntity(ctx context.Context, e *codemodels.CodeEntity) error {
	query := `
		INSERT INTO code_entities (
			id, repository_id, entity_type, name, qualified_name, file_path,
			package_path, module_path, language, start_line, end_line,
			signature, endpoint_type, evidence_json, indexed_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			repository_id  = excluded.repository_id,
			entity_type    = excluded.entity_type,
			name           = excluded.name,
			qualified_name = excluded.qualified_name,
			file_path      = excluded.file_path,
			package_path   = excluded.package_path,
			module_path    = excluded.module_path,
			language       = excluded.language,
			start_line     = excluded.start_line,
			end_line       = excluded.end_line,
			signature      = excluded.signature,
			endpoint_type  = excluded.endpoint_type,
			evidence_json  = excluded.evidence_json,
			indexed_at     = excluded.indexed_at
	`
	_, err := s.db.ExecContext(ctx, query,
		e.ID, e.RepositoryID, e.EntityType, e.Name, e.QualifiedName, e.FilePath,
		e.PackagePath, e.ModulePath, e.Language, e.StartLine, e.EndLine,
		e.Signature, e.EndpointType, e.EvidenceJSON, e.IndexedAt.UTC().Format("2006-01-02T15:04:05Z"),
	)
	if err != nil {
		return fmt.Errorf("failed to upsert code entity: %w", err)
	}
	return nil
}

// DeleteByRepository removes all code entities for a repository.
func (s *SQLiteCodeEntityStore) DeleteByRepository(ctx context.Context, repositoryID string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM code_entities WHERE repository_id = ?`, repositoryID)
	if err != nil {
		return fmt.Errorf("failed to delete code entities: %w", err)
	}
	return nil
}
