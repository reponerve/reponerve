package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

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

// MergeCodeIndexModules replaces code entities and relationships for specific module paths.
func (db *Database) MergeCodeIndexModules(
	ctx context.Context,
	repositoryID string,
	modulePaths []string,
	entities []*codemodels.CodeEntity,
	rels []*codemodels.CodeRelationship,
) error {
	if len(modulePaths) == 0 {
		return fmt.Errorf("module paths required for scoped merge")
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin scoped code index transaction: %w", err)
	}
	defer tx.Rollback()

	placeholders := strings.Repeat("?,", len(modulePaths))
	placeholders = placeholders[:len(placeholders)-1]

	deleteRepoCodeLinks := fmt.Sprintf(`
		DELETE FROM repository_code_relationships
		WHERE repository_id = ?
		  AND code_entity_id IN (SELECT id FROM code_entities WHERE repository_id = ? AND module_path IN (%s))`, placeholders)
	repoCodeArgs := []interface{}{repositoryID, repositoryID}
	repoCodeArgs = append(repoCodeArgs, modulePathArgs(modulePaths)...)
	if _, err := tx.ExecContext(ctx, deleteRepoCodeLinks, repoCodeArgs...); err != nil {
		return fmt.Errorf("delete scoped repository-code links: %w", err)
	}

	deleteRels := fmt.Sprintf(`
		DELETE FROM code_relationships
		WHERE repository_id = ?
		  AND (
		    from_entity_id IN (SELECT id FROM code_entities WHERE repository_id = ? AND module_path IN (%s))
		    OR to_entity_id IN (SELECT id FROM code_entities WHERE repository_id = ? AND module_path IN (%s))
		  )`, placeholders, placeholders)
	relArgs := []interface{}{repositoryID, repositoryID}
	relArgs = append(relArgs, modulePathArgs(modulePaths)...)
	relArgs = append(relArgs, repositoryID)
	relArgs = append(relArgs, modulePathArgs(modulePaths)...)
	if _, err := tx.ExecContext(ctx, deleteRels, relArgs...); err != nil {
		return fmt.Errorf("delete scoped code relationships: %w", err)
	}

	deleteEntities := fmt.Sprintf(`DELETE FROM code_entities WHERE repository_id = ? AND module_path IN (%s)`, placeholders)
	entArgs := []interface{}{repositoryID}
	entArgs = append(entArgs, modulePathArgs(modulePaths)...)
	if _, err := tx.ExecContext(ctx, deleteEntities, entArgs...); err != nil {
		return fmt.Errorf("delete scoped code entities: %w", err)
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
		return fmt.Errorf("commit scoped code index transaction: %w", err)
	}
	return nil
}

func modulePathArgs(modulePaths []string) []interface{} {
	out := make([]interface{}, len(modulePaths))
	for i, mp := range modulePaths {
		out[i] = mp
	}
	return out
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
