package ingestion

import "context"

// CodeLinker links repository memory entities to indexed code entities.
type CodeLinker interface {
	Link(ctx context.Context, repositoryID string) error
}
