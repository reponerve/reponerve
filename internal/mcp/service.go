package mcp

import (
	"reponerve/internal/context"
	"reponerve/internal/context/render"
	"reponerve/internal/query/storage"
)

// Service aggregates the core repository intelligence capabilities (readers, generator, renderer).
type Service struct {
	DecisionReader     storage.DecisionReader
	IntentReader       storage.IntentReader
	FactReader         storage.FactReader
	EventReader        storage.EventReader
	RelationshipReader storage.RelationshipReader
	Generator          *context.Generator
	Renderer           *render.Renderer
}

// NewService creates a new Service instance aggregating the given dependencies.
func NewService(
	dr storage.DecisionReader,
	ir storage.IntentReader,
	fr storage.FactReader,
	er storage.EventReader,
	rr storage.RelationshipReader,
	g *context.Generator,
	ren *render.Renderer,
) *Service {
	return &Service{
		DecisionReader:     dr,
		IntentReader:       ir,
		FactReader:         fr,
		EventReader:        er,
		RelationshipReader: rr,
		Generator:          g,
		Renderer:           ren,
	}
}
