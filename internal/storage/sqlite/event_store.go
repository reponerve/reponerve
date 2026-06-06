package sqlite

import (
	"context"
	"fmt"
	"time"

	"reponerve/pkg/models"
)

// EventStore implements storage.EventStore for SQLite.
type EventStore struct {
	db *Database
}

// NewEventStore creates a new SQLite EventStore.
func NewEventStore(db *Database) *EventStore {
	return &EventStore{db: db}
}

// UpsertEvent persists or updates an extracted event record.
func (s *EventStore) UpsertEvent(ctx context.Context, event *models.Event) error {
	query := `
		INSERT INTO memory_events (id, repository_id, event_type, title, description, source_id, timestamp, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			event_type  = excluded.event_type,
			title       = excluded.title,
			description = excluded.description,
			timestamp   = excluded.timestamp
	`
	var desc interface{} = nil
	if event.Description != "" {
		desc = event.Description
	}

	_, err := s.db.ExecContext(ctx, query,
		event.ID,
		event.RepositoryID,
		event.EventType,
		event.Title,
		desc,
		event.SourceID,
		event.Timestamp,
		time.Now(),
	)
	if err != nil {
		return fmt.Errorf("failed to upsert event: %w", err)
	}
	return nil
}
