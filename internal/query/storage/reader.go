package storage

import (
	"context"

	memorymodels "reponerve/internal/memory/models"
	models "reponerve/pkg/models"
)

// EventReader defines the read interface for Event memories.
type EventReader interface {
	GetByID(ctx context.Context, id string) (*models.Event, error)
	ListByRepository(ctx context.Context, repositoryID string) ([]*models.Event, error)
	ListAll(ctx context.Context) ([]*models.Event, error)
}

// DecisionReader defines the read interface for Decision memories.
type DecisionReader interface {
	GetByID(ctx context.Context, id string) (*memorymodels.Decision, error)
	ListByRepository(ctx context.Context, repositoryID string) ([]*memorymodels.Decision, error)
	ListAll(ctx context.Context) ([]*memorymodels.Decision, error)
}

// IntentReader defines the read interface for Intent memories.
type IntentReader interface {
	GetByID(ctx context.Context, id string) (*memorymodels.Intent, error)
	ListByRepository(ctx context.Context, repositoryID string) ([]*memorymodels.Intent, error)
	ListAll(ctx context.Context) ([]*memorymodels.Intent, error)
}

// FactReader defines the read interface for Fact memories.
type FactReader interface {
	GetByID(ctx context.Context, id string) (*memorymodels.Fact, error)
	ListByRepository(ctx context.Context, repositoryID string) ([]*memorymodels.Fact, error)
	ListAll(ctx context.Context) ([]*memorymodels.Fact, error)
}

// RelationshipReader defines the read interface for Relationship memories.
type RelationshipReader interface {
	GetByID(ctx context.Context, id string) (*memorymodels.Relationship, error)
	ListByRepository(ctx context.Context, repositoryID string) ([]*memorymodels.Relationship, error)
	ListAll(ctx context.Context) ([]*memorymodels.Relationship, error)
}
