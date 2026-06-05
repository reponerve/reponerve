package scanner

import (
	"context"
	"reponerve/pkg/models"
)

// SourceScanner represents a common interface for all repository scanners (Git, ADR, Docs, etc.)
type SourceScanner interface {
	Scan(ctx context.Context, repo *models.Repository) ([]*models.Source, error)
}
