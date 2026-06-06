package ingestion

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"reponerve/internal/extraction/decision"
	"reponerve/internal/extraction/event"
	memorystorage "reponerve/internal/memory/storage"
	"reponerve/internal/scanner/repository"
	"reponerve/internal/storage"
)

// Coordinator coordinates repository discovery and pipeline execution.
type Coordinator struct {
	discovery      *repository.GitDiscovery
	repoStore      storage.RepositoryStore
	sourceStore    storage.SourceStore
	scanStateStore storage.ScanStateStore
	eventStore     storage.EventStore
	decisionStore  memorystorage.DecisionStore
	pipeline       *Pipeline
}

// NewCoordinator creates a new Coordinator instance.
func NewCoordinator(
	discovery *repository.GitDiscovery,
	repoStore storage.RepositoryStore,
	sourceStore storage.SourceStore,
	scanStateStore storage.ScanStateStore,
	eventStore storage.EventStore,
	decisionStore memorystorage.DecisionStore,
	pipeline *Pipeline,
) *Coordinator {
	return &Coordinator{
		discovery:      discovery,
		repoStore:      repoStore,
		sourceStore:    sourceStore,
		scanStateStore: scanStateStore,
		eventStore:     eventStore,
		decisionStore:  decisionStore,
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

	// Extract and persist Events (ISSUE-011)
	eventExtractor := event.NewExtractor()
	events, err := eventExtractor.Extract(ctx, sources)
	if err != nil {
		return nil, fmt.Errorf("failed to extract events: %w", err)
	}
	for _, evt := range events {
		if err := c.eventStore.UpsertEvent(ctx, evt); err != nil {
			return nil, fmt.Errorf("failed to store event: %w", err)
		}
	}

	// Extract and persist Decisions (ISSUE-012)
	decisionExtractor := decision.NewExtractor()
	decisions, err := decisionExtractor.Extract(ctx, sources)
	if err != nil {
		return nil, fmt.Errorf("failed to extract decisions: %w", err)
	}
	for _, dec := range decisions {
		if err := c.decisionStore.UpsertDecision(ctx, dec); err != nil {
			return nil, fmt.Errorf("failed to store decision: %w", err)
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
