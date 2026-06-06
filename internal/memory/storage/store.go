package storage

import (
	"context"

	"reponerve/internal/memory/models"
)

// DecisionStore defines the persistence interface for extracted Decision records.
type DecisionStore interface {
	UpsertDecision(ctx context.Context, decision *models.Decision) error
}
