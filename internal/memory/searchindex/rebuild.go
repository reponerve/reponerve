package searchindex

import (
	"context"
	"fmt"

	"github.com/reponerve/reponerve/internal/query/storage"
	searchstorage "github.com/reponerve/reponerve/internal/storage"
)

// RebuildFromRepository rebuilds the FTS index from all persisted repository memory.
func RebuildFromRepository(
	ctx context.Context,
	repositoryID string,
	eventReader storage.EventReader,
	decisionReader storage.DecisionReader,
	factReader storage.FactReader,
	memorySearchStore searchstorage.MemorySearchStore,
) error {
	events, err := eventReader.ListByRepository(ctx, repositoryID)
	if err != nil {
		return fmt.Errorf("failed to list events for search index: %w", err)
	}

	decisions, err := decisionReader.ListByRepository(ctx, repositoryID)
	if err != nil {
		return fmt.Errorf("failed to list decisions for search index: %w", err)
	}

	facts, err := factReader.ListByRepository(ctx, repositoryID)
	if err != nil {
		return fmt.Errorf("failed to list facts for search index: %w", err)
	}

	docs := BuildDocuments(Input{
		RepositoryID: repositoryID,
		Events:       events,
		Decisions:    decisions,
		Facts:        facts,
	})

	if err := memorySearchStore.Rebuild(ctx, repositoryID, docs); err != nil {
		return fmt.Errorf("failed to rebuild memory search index: %w", err)
	}
	return nil
}
