package storage

import (
	"context"

	"github.com/reponerve/reponerve/internal/memory/models"
)

// DecisionStore defines the persistence interface for extracted Decision records.
type DecisionStore interface {
	UpsertDecision(ctx context.Context, decision *models.Decision) error
}

// IntentStore defines the persistence interface for extracted Intent records.
type IntentStore interface {
	UpsertIntent(ctx context.Context, intent *models.Intent) error
}

// FactStore defines the persistence interface for extracted Fact records.
type FactStore interface {
	UpsertFact(ctx context.Context, fact *models.Fact) error
}

// RelationshipStore defines the persistence interface for memory relationships.
type RelationshipStore interface {
	UpsertRelationship(ctx context.Context, relationship *models.Relationship) error
}
