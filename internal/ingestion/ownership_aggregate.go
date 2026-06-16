package ingestion

import (
	"context"
	"fmt"

	"github.com/reponerve/reponerve/internal/ownership/expertise"
	ownerextraction "github.com/reponerve/reponerve/internal/ownership/extraction"
	querystorage "github.com/reponerve/reponerve/internal/query/storage"
	"github.com/reponerve/reponerve/pkg/models"
)

// OwnershipReaders supplies persisted repository memory for ownership recomputation.
type OwnershipReaders struct {
	Sources   querystorage.SourceReader
	Events    querystorage.EventReader
	Decisions querystorage.DecisionReader
	Facts     querystorage.FactReader
}

func (c *Coordinator) recomputeOwnership(ctx context.Context, repositoryID string) error {
	if c.ownershipReaders == nil {
		return nil
	}

	allSources, err := c.ownershipReaders.Sources.ListByRepository(ctx, repositoryID)
	if err != nil {
		return fmt.Errorf("list sources for ownership: %w", err)
	}

	commitSources := filterSourcesByType(allSources, "commit")
	contribExtractor := ownerextraction.NewExtractor()
	contribs, err := contribExtractor.Extract(ctx, commitSources)
	if err != nil {
		return fmt.Errorf("extract contributors: %w", err)
	}
	for _, contr := range contribs {
		if err := c.contributorStore.UpsertContributor(ctx, contr); err != nil {
			return fmt.Errorf("store contributor: %w", err)
		}
	}

	events, err := c.ownershipReaders.Events.ListByRepository(ctx, repositoryID)
	if err != nil {
		return fmt.Errorf("list events for expertise: %w", err)
	}
	decisions, err := c.ownershipReaders.Decisions.ListByRepository(ctx, repositoryID)
	if err != nil {
		return fmt.Errorf("list decisions for expertise: %w", err)
	}
	facts, err := c.ownershipReaders.Facts.ListByRepository(ctx, repositoryID)
	if err != nil {
		return fmt.Errorf("list facts for expertise: %w", err)
	}

	if err := c.expertiseStore.DeleteByRepository(ctx, repositoryID); err != nil {
		return fmt.Errorf("clear expertise: %w", err)
	}

	expertiseDetector := expertise.NewDetector()
	expertiseRecords, err := expertiseDetector.Detect(ctx, contribs, events, decisions, facts, commitSources)
	if err != nil {
		return fmt.Errorf("detect expertise: %w", err)
	}
	for _, exp := range expertiseRecords {
		if err := c.expertiseStore.UpsertExpertise(ctx, exp); err != nil {
			return fmt.Errorf("store expertise: %w", err)
		}
	}
	return nil
}

// RecomputeOwnership rebuilds contributor and expertise records from persisted repository memory.
func (c *Coordinator) RecomputeOwnership(ctx context.Context, repositoryID string) error {
	return c.recomputeOwnership(ctx, repositoryID)
}

func filterSourcesByType(sources []*models.Source, sourceType string) []*models.Source {
	var out []*models.Source
	for _, src := range sources {
		if src.SourceType == sourceType {
			out = append(out, src)
		}
	}
	return out
}
