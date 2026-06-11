package storage

import (
	"context"
	"time"

	codemodels "github.com/reponerve/reponerve/internal/code/models"
	"github.com/reponerve/reponerve/pkg/models"
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

// ContributorStore defines persistence operations for contributors.
type ContributorStore interface {
	UpsertContributor(ctx context.Context, contributor *models.Contributor) error
}

// ExpertiseStore defines persistence operations for expertise.
type ExpertiseStore interface {
	UpsertExpertise(ctx context.Context, expertise *models.Expertise) error
}

// CodeEntityStore defines persistence operations for code entities.
type CodeEntityStore interface {
	UpsertCodeEntity(ctx context.Context, entity *codemodels.CodeEntity) error
	DeleteByRepository(ctx context.Context, repositoryID string) error
}

// CodeRelationshipStore defines persistence operations for code relationships.
type CodeRelationshipStore interface {
	UpsertCodeRelationship(ctx context.Context, rel *codemodels.CodeRelationship) error
	DeleteByRepository(ctx context.Context, repositoryID string) error
}

// RepositoryCodeRelationshipStore defines persistence for repository-code links.
type RepositoryCodeRelationshipStore interface {
	UpsertRepositoryCodeRelationship(ctx context.Context, rel *codemodels.RepositoryCodeRelationship) error
	DeleteByRepository(ctx context.Context, repositoryID string) error
}

// CodeIndexStateStore defines persistence for code index state.
type CodeIndexStateStore interface {
	GetByRepository(ctx context.Context, repositoryID string) (*codemodels.CodeIndexState, error)
	UpsertCodeIndexState(ctx context.Context, state *codemodels.CodeIndexState) error
	UpdateLinkCount(ctx context.Context, repositoryID string, linkCount int) error
}
