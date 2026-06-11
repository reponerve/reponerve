package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	codemodels "github.com/reponerve/reponerve/internal/code/models"
	"github.com/reponerve/reponerve/internal/storage/sqlite"
)

// CodeEntityReader reads indexed code entities.
type CodeEntityReader interface {
	GetByID(ctx context.Context, id string) (*codemodels.CodeEntity, error)
	ListByRepository(ctx context.Context, repositoryID string) ([]*codemodels.CodeEntity, error)
	ListByFilePath(ctx context.Context, repositoryID, filePath string) ([]*codemodels.CodeEntity, error)
	ListByModulePath(ctx context.Context, repositoryID, modulePath string) ([]*codemodels.CodeEntity, error)
	ListByEntityType(ctx context.Context, repositoryID, entityType string) ([]*codemodels.CodeEntity, error)
	FindByQualifiedName(ctx context.Context, repositoryID, qualifiedName string) ([]*codemodels.CodeEntity, error)
}

// SQLiteCodeEntityReader implements CodeEntityReader.
type SQLiteCodeEntityReader struct {
	db *sqlite.Database
}

// NewSQLiteCodeEntityReader creates a code entity reader.
func NewSQLiteCodeEntityReader(db *sqlite.Database) *SQLiteCodeEntityReader {
	return &SQLiteCodeEntityReader{db: db}
}

const codeEntitySelect = `
	SELECT id, repository_id, entity_type, name, qualified_name, file_path,
	       package_path, module_path, language, start_line, end_line,
	       COALESCE(signature, ''), COALESCE(endpoint_type, ''), evidence_json, indexed_at
`

func scanCodeEntity(scanner interface {
	Scan(dest ...any) error
}) (*codemodels.CodeEntity, error) {
	var e codemodels.CodeEntity
	var indexedAt string
	if err := scanner.Scan(
		&e.ID, &e.RepositoryID, &e.EntityType, &e.Name, &e.QualifiedName, &e.FilePath,
		&e.PackagePath, &e.ModulePath, &e.Language, &e.StartLine, &e.EndLine,
		&e.Signature, &e.EndpointType, &e.EvidenceJSON, &indexedAt,
	); err != nil {
		return nil, err
	}
	if t, err := time.Parse(time.RFC3339, indexedAt); err == nil {
		e.IndexedAt = t
	}
	return &e, nil
}

func (r *SQLiteCodeEntityReader) queryEntities(ctx context.Context, query string, args ...any) ([]*codemodels.CodeEntity, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query code entities: %w", err)
	}
	defer rows.Close()

	var entities []*codemodels.CodeEntity
	for rows.Next() {
		e, err := scanCodeEntity(rows)
		if err != nil {
			return nil, fmt.Errorf("scan code entity: %w", err)
		}
		entities = append(entities, e)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return entities, nil
}

// GetByID returns one code entity by ID.
func (r *SQLiteCodeEntityReader) GetByID(ctx context.Context, id string) (*codemodels.CodeEntity, error) {
	query := codeEntitySelect + ` FROM code_entities WHERE id = ?`
	e, err := scanCodeEntity(r.db.QueryRowContext(ctx, query, id))
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("code entity not found: %s", id)
	}
	if err != nil {
		return nil, fmt.Errorf("get code entity: %w", err)
	}
	return e, nil
}

// ListByRepository returns code entities sorted deterministically.
func (r *SQLiteCodeEntityReader) ListByRepository(ctx context.Context, repositoryID string) ([]*codemodels.CodeEntity, error) {
	query := codeEntitySelect + `
		FROM code_entities
		WHERE repository_id = ?
		ORDER BY entity_type ASC, module_path ASC, file_path ASC, start_line ASC, qualified_name ASC, id ASC
	`
	return r.queryEntities(ctx, query, repositoryID)
}

// ListByFilePath returns entities associated with a file path.
func (r *SQLiteCodeEntityReader) ListByFilePath(ctx context.Context, repositoryID, filePath string) ([]*codemodels.CodeEntity, error) {
	query := codeEntitySelect + `
		FROM code_entities
		WHERE repository_id = ? AND file_path = ?
		ORDER BY entity_type ASC, start_line ASC, qualified_name ASC, id ASC
	`
	return r.queryEntities(ctx, query, repositoryID, filePath)
}

// ListByModulePath returns entities for a module path.
func (r *SQLiteCodeEntityReader) ListByModulePath(ctx context.Context, repositoryID, modulePath string) ([]*codemodels.CodeEntity, error) {
	query := codeEntitySelect + `
		FROM code_entities
		WHERE repository_id = ? AND module_path = ?
		ORDER BY entity_type ASC, file_path ASC, start_line ASC, qualified_name ASC, id ASC
	`
	return r.queryEntities(ctx, query, repositoryID, modulePath)
}

// ListByEntityType returns entities of one type for a repository.
func (r *SQLiteCodeEntityReader) ListByEntityType(ctx context.Context, repositoryID, entityType string) ([]*codemodels.CodeEntity, error) {
	query := codeEntitySelect + `
		FROM code_entities
		WHERE repository_id = ? AND entity_type = ?
		ORDER BY module_path ASC, file_path ASC, start_line ASC, qualified_name ASC, id ASC
	`
	return r.queryEntities(ctx, query, repositoryID, entityType)
}

// FindByQualifiedName returns entities with an exact qualified name.
func (r *SQLiteCodeEntityReader) FindByQualifiedName(ctx context.Context, repositoryID, qualifiedName string) ([]*codemodels.CodeEntity, error) {
	query := codeEntitySelect + `
		FROM code_entities
		WHERE repository_id = ? AND qualified_name = ?
		ORDER BY entity_type ASC, id ASC
	`
	return r.queryEntities(ctx, query, repositoryID, qualifiedName)
}
