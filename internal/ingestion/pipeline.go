package ingestion

import (
	"context"
	"fmt"

	"github.com/reponerve/reponerve/pkg/models"
)

// Pipeline orchestrates the execution of all registered scanners.
type Pipeline struct {
	registry *Registry
}

// NewPipeline creates a new Pipeline instance.
func NewPipeline(registry *Registry) *Pipeline {
	return &Pipeline{registry: registry}
}

// Execute runs all registered scanners and returns the combined list of sources.
func (p *Pipeline) Execute(ctx context.Context, repo *models.Repository) ([]*models.Source, error) {
	var allSources []*models.Source
	for _, entry := range p.registry.Scanners() {
		sources, err := entry.Scanner.Scan(ctx, repo)
		if err != nil {
			return nil, fmt.Errorf("scanner %q failed: %w", entry.Name, err)
		}
		allSources = append(allSources, sources...)
	}
	return allSources, nil
}
