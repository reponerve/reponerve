package mcp

import (
	"reponerve/internal/context"
	"reponerve/internal/context/render"
	"reponerve/internal/graph/impact"
	"reponerve/internal/graph/traversal"
	"reponerve/internal/intelligence/changeplan"
	"reponerve/internal/intelligence/discovery"
	"reponerve/internal/intelligence/learning"
	"reponerve/internal/intelligence/reviewers"
	ownershipquery "reponerve/internal/ownership/query"
	"reponerve/internal/query/storage"
)

// Service aggregates the core repository intelligence capabilities (readers, generator, renderer).
type Service struct {
	DecisionReader       storage.DecisionReader
	IntentReader         storage.IntentReader
	FactReader           storage.FactReader
	EventReader          storage.EventReader
	RelationshipReader   storage.RelationshipReader
	Generator            *context.Generator
	Renderer             *render.Renderer
	OwnershipReader      *ownershipquery.Reader
	GraphTraversalEngine *traversal.Engine
	GraphImpactService   *impact.Service
	DiscoveryService     *discovery.Service
	LearningService      *learning.Service
	ReviewerService      *reviewers.Service
	ChangePlanService    *changeplan.Service
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
	or *ownershipquery.Reader,
	travEngine *traversal.Engine,
	impactSvc *impact.Service,
	discoverySvc *discovery.Service,
	learningSvc *learning.Service,
	reviewerSvc *reviewers.Service,
	changePlanSvc *changeplan.Service,
) *Service {
	return &Service{
		DecisionReader:       dr,
		IntentReader:         ir,
		FactReader:           fr,
		EventReader:          er,
		RelationshipReader:   rr,
		Generator:            g,
		Renderer:             ren,
		OwnershipReader:      or,
		GraphTraversalEngine: travEngine,
		GraphImpactService:   impactSvc,
		DiscoveryService:     discoverySvc,
		LearningService:      learningSvc,
		ReviewerService:      reviewerSvc,
		ChangePlanService:    changePlanSvc,
	}
}
