package mcp

import (
	"github.com/reponerve/reponerve/internal/context"
	"github.com/reponerve/reponerve/internal/context/render"
	"github.com/reponerve/reponerve/internal/graph/impact"
	"github.com/reponerve/reponerve/internal/graph/traversal"
	"github.com/reponerve/reponerve/internal/intelligence/changeplan"
	"github.com/reponerve/reponerve/internal/intelligence/discovery"
	"github.com/reponerve/reponerve/internal/intelligence/learning"
	"github.com/reponerve/reponerve/internal/intelligence/reviewers"
	ownershipquery "github.com/reponerve/reponerve/internal/ownership/query"
	"github.com/reponerve/reponerve/internal/query/storage"
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
