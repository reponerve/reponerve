package sqlite

import (
	"context"
	"fmt"
	"time"

	"github.com/reponerve/reponerve/pkg/models"
)

// SourceStore implements storage.SourceStore for SQLite.
type SourceStore struct {
	db *Database
}

// NewSourceStore creates a new SQLite SourceStore.
func NewSourceStore(db *Database) *SourceStore {
	return &SourceStore{db: db}
}

// UpsertSource persists or updates a source record.
func (s *SourceStore) UpsertSource(ctx context.Context, src *models.Source) error {
	query := `
		INSERT INTO sources (id, repository_id, source_type, reference, title, author, timestamp, metadata_json, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			title = excluded.title,
			author = excluded.author,
			timestamp = excluded.timestamp,
			metadata_json = excluded.metadata_json
	`
	now := time.Now()

	var val interface{} = nil
	if src.MetadataJSON != "" {
		val = src.MetadataJSON
	}

	_, err := s.db.ExecContext(ctx, query, src.ID, src.RepositoryID, src.SourceType, src.Reference, src.Title, src.Author, src.Timestamp, val, now)
	if err != nil {
		return fmt.Errorf("failed to store source record: %w", err)
	}
	return nil
}

