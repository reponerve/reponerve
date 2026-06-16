package sqlite

import (
	"context"
	"database/sql"
	"fmt"

	codemodels "github.com/reponerve/reponerve/internal/code/models"
)

// ReplaceCodeIndex atomically replaces code entities and relationships for one repository.
func (db *Database) ReplaceCodeIndex(
	ctx context.Context,
	repositoryID string,
	entities []*codemodels.CodeEntity,
	rels []*codemodels.CodeRelationship,
) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin code index transaction: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `DELETE FROM repository_code_relationships WHERE repository_id = ?`, repositoryID); err != nil {
		return fmt.Errorf("delete repository code links: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM code_relationships WHERE repository_id = ?`, repositoryID); err != nil {
		return fmt.Errorf("delete code relationships: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM code_entities WHERE repository_id = ?`, repositoryID); err != nil {
		return fmt.Errorf("delete code entities: %w", err)
	}

	entityStmt, err := tx.PrepareContext(ctx, `
		INSERT INTO code_entities (
			id, repository_id, entity_type, name, qualified_name, file_path,
			package_path, module_path, language, start_line, end_line,
			signature, endpoint_type, evidence_json, indexed_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("prepare code entity insert: %w", err)
	}
	defer entityStmt.Close()

	for _, e := range entities {
		if _, err := entityStmt.ExecContext(ctx,
			e.ID, e.RepositoryID, e.EntityType, e.Name, e.QualifiedName, e.FilePath,
			e.PackagePath, e.ModulePath, e.Language, e.StartLine, e.EndLine,
			e.Signature, e.EndpointType, e.EvidenceJSON, e.IndexedAt.UTC().Format("2006-01-02T15:04:05Z"),
		); err != nil {
			return fmt.Errorf("insert code entity %s: %w", e.QualifiedName, err)
		}
	}

	relStmt, err := tx.PrepareContext(ctx, `
		INSERT INTO code_relationships (
			id, repository_id, from_entity_id, to_entity_id, relationship_type, evidence_json, indexed_at
		) VALUES (?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("prepare code relationship insert: %w", err)
	}
	defer relStmt.Close()

	for _, rel := range rels {
		if _, err := relStmt.ExecContext(ctx,
			rel.ID, rel.RepositoryID, rel.FromEntityID, rel.ToEntityID,
			rel.RelationshipType, rel.EvidenceJSON, rel.IndexedAt.UTC().Format("2006-01-02T15:04:05Z"),
		); err != nil {
			return fmt.Errorf("insert code relationship %s: %w", rel.RelationshipType, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit code index transaction: %w", err)
	}
	return nil
}

// WithTransaction runs fn inside a SQLite transaction.
func (db *Database) WithTransaction(ctx context.Context, fn func(*sql.Tx) error) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if err := fn(tx); err != nil {
		return err
	}
	return tx.Commit()
}
