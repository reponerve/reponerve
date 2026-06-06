package storage

import (
	"context"
	"time"

	"reponerve/pkg/models"
)

// RepositoryStore defines persistence operations for repository metadata.
type RepositoryStore interface {
	UpsertRepository(ctx context.Context, repo *models.Repository) error
}

// ScanState holds the scan state for a repository.
type ScanState struct {
	RepositoryID   string
	LastScanCommit string
	UpdatedAt      time.Time
}

// SourceStore defines persistence operations for sources.
type SourceStore interface {
	UpsertSource(ctx context.Context, source *models.Source) error
}

// ScanStateStore defines persistence operations for scanning state.
type ScanStateStore interface {
	GetScanState(ctx context.Context, repoID string) (*ScanState, error)
	UpdateScanState(ctx context.Context, repoID string, commitHash string) error
}

// EventStore defines persistence operations for extracted events.
type EventStore interface {
	UpsertEvent(ctx context.Context, event *models.Event) error
}
