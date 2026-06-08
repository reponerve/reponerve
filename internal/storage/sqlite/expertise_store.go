package sqlite

import (
	"context"
	"fmt"

	"github.com/reponerve/reponerve/pkg/models"
)

// SQLiteExpertiseStore implements storage.ExpertiseStore for SQLite.
type SQLiteExpertiseStore struct {
	db *Database
}

// NewSQLiteExpertiseStore creates a new SQLiteExpertiseStore.
func NewSQLiteExpertiseStore(db *Database) *SQLiteExpertiseStore {
	return &SQLiteExpertiseStore{db: db}
}

// UpsertExpertise inserts or updates an expertise record.
func (s *SQLiteExpertiseStore) UpsertExpertise(ctx context.Context, e *models.Expertise) error {
	query := `
		INSERT INTO expertise (id, repository_id, contributor_id, domain, score, evidence_json)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			repository_id  = excluded.repository_id,
			contributor_id = excluded.contributor_id,
			domain         = excluded.domain,
			score          = excluded.score,
			evidence_json  = excluded.evidence_json
	`
	_, err := s.db.ExecContext(ctx, query,
		e.ID,
		e.RepositoryID,
		e.ContributorID,
		e.Domain,
		e.Score,
		e.EvidenceJSON,
	)
	if err != nil {
		return fmt.Errorf("failed to upsert expertise in sqlite: %w", err)
	}
	return nil
}
