package ingestion

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"reponerve/internal/scanner/repository"
	"reponerve/internal/storage"
)

// Coordinator coordinates repository discovery and pipeline execution.
type Coordinator struct {
	discovery      *repository.GitDiscovery
	repoStore      storage.RepositoryStore
	sourceStore    storage.SourceStore
	scanStateStore storage.ScanStateStore
	pipeline       *Pipeline
}

// NewCoordinator creates a new Coordinator instance.
func NewCoordinator(
	discovery *repository.GitDiscovery,
	repoStore storage.RepositoryStore,
	sourceStore storage.SourceStore,
	scanStateStore storage.ScanStateStore,
	pipeline *Pipeline,
) *Coordinator {
	return &Coordinator{
		discovery:      discovery,
		repoStore:      repoStore,
		sourceStore:    sourceStore,
		scanStateStore: scanStateStore,
		pipeline:       pipeline,
	}
}

// Run discovers the repository metadata, stores it, runs all scanners, stores the discovered sources, updates scan state, and returns stats.
func (c *Coordinator) Run(ctx context.Context, path string) (*ScanResult, error) {
	startTime := time.Now()

	repo, err := c.discovery.Discover(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("failed to discover repository: %w", err)
	}

	err = c.repoStore.UpsertRepository(ctx, repo)
	if err != nil {
		return nil, fmt.Errorf("failed to store repository metadata: %w", err)
	}

	sources, err := c.pipeline.Execute(ctx, repo)
	if err != nil {
		return nil, fmt.Errorf("failed to execute ingestion pipeline: %w", err)
	}

	var commitsIndexed, adrsIndexed int
	for _, source := range sources {
		err = c.sourceStore.UpsertSource(ctx, source)
		if err != nil {
			return nil, fmt.Errorf("failed to store source record: %w", err)
		}
		if source.SourceType == "commit" {
			commitsIndexed++
		} else if source.SourceType == "adr" {
			adrsIndexed++
		}
	}

	// Update Git Scan State to HEAD
	headCmd := exec.CommandContext(ctx, "git", "rev-parse", "HEAD")
	headCmd.Dir = repo.Path
	headOut, err := headCmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get HEAD commit: %w", err)
	}
	headHash := strings.TrimSpace(string(headOut))

	err = c.scanStateStore.UpdateScanState(ctx, repo.ID, headHash)
	if err != nil {
		return nil, fmt.Errorf("failed to update scan state: %w", err)
	}

	return &ScanResult{
		RepositoryID:   repo.ID,
		CommitsIndexed: commitsIndexed,
		ADRsIndexed:    adrsIndexed,
		Duration:       time.Since(startTime),
	}, nil
}
