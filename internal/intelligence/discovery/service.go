package discovery

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/reponerve/reponerve/internal/graph/impact"
	"github.com/reponerve/reponerve/internal/graph/model"
	"github.com/reponerve/reponerve/internal/graph/relationships"
	"github.com/reponerve/reponerve/internal/graph/traversal"
	"github.com/reponerve/reponerve/internal/query/storage"
	models "github.com/reponerve/reponerve/pkg/models"
)

// Service provides repository knowledge discovery capabilities.
type Service struct {
	decisionReader       storage.DecisionReader
	factReader           storage.FactReader
	eventReader          storage.EventReader
	contribReader        storage.ContributorReader
	expertiseReader      storage.ExpertiseReader
	relationshipReader   storage.RelationshipReader
	relationshipEngine   *relationships.Engine
	graphTraversalEngine *traversal.Engine
	graphImpactService   *impact.Service
}

// NewService constructs a new discovery Service.
func NewService(
	dr storage.DecisionReader,
	fr storage.FactReader,
	er storage.EventReader,
	cr storage.ContributorReader,
	expr storage.ExpertiseReader,
	rr storage.RelationshipReader,
	relEngine *relationships.Engine,
	travEngine *traversal.Engine,
	impactSvc *impact.Service,
) *Service {
	return &Service{
		decisionReader:       dr,
		factReader:           fr,
		eventReader:          er,
		contribReader:        cr,
		expertiseReader:      expr,
		relationshipReader:   rr,
		relationshipEngine:   relEngine,
		graphTraversalEngine: travEngine,
		graphImpactService:   impactSvc,
	}
}

// Discover generates a KnowledgeDiscoveryReport identifying important repository entities.
func (s *Service) Discover(ctx context.Context, repositoryID string) (*KnowledgeDiscoveryReport, error) {
	if repositoryID == "" {
		return nil, fmt.Errorf("repository ID cannot be empty")
	}

	// 1. Load all relevant entity records from memory readers
	decs, err := s.decisionReader.ListByRepository(ctx, repositoryID)
	if err != nil {
		return nil, fmt.Errorf("failed to list decisions: %w", err)
	}

	intents, err := s.relationshipEngine.IntentReader().ListByRepository(ctx, repositoryID)
	if err != nil {
		return nil, fmt.Errorf("failed to list intents: %w", err)
	}

	facts, err := s.factReader.ListByRepository(ctx, repositoryID)
	if err != nil {
		return nil, fmt.Errorf("failed to list facts: %w", err)
	}

	events, err := s.eventReader.ListByRepository(ctx, repositoryID)
	if err != nil {
		return nil, fmt.Errorf("failed to list events: %w", err)
	}

	contribs, err := s.contribReader.ListByRepository(ctx, repositoryID)
	if err != nil {
		return nil, fmt.Errorf("failed to list contributors: %w", err)
	}

	exps, err := s.expertiseReader.ListByRepository(ctx, repositoryID)
	if err != nil {
		return nil, fmt.Errorf("failed to list expertise: %w", err)
	}

	// 2. Build map of EntityID -> GraphNodeID using model.NodeID
	entityToNodeID := make(map[string]string)
	for _, d := range decs {
		entityToNodeID[d.ID] = model.NodeID(repositoryID, model.NodeTypeDecision, d.ID)
	}
	for _, i := range intents {
		entityToNodeID[i.ID] = model.NodeID(repositoryID, model.NodeTypeIntent, i.ID)
	}
	for _, f := range facts {
		entityToNodeID[f.ID] = model.NodeID(repositoryID, model.NodeTypeFact, f.ID)
	}
	for _, ev := range events {
		entityToNodeID[ev.ID] = model.NodeID(repositoryID, model.NodeTypeEvent, ev.ID)
	}
	for _, c := range contribs {
		entityToNodeID[c.ID] = model.NodeID(repositoryID, model.NodeTypeContributor, c.ID)
	}
	for _, exp := range exps {
		entityToNodeID[exp.ID] = model.NodeID(repositoryID, model.NodeTypeExpertise, exp.ID)
	}

	// 3. Count graph relationships (stored + derived) per node
	graphCounts := make(map[string]int)

	// Inbound/outbound stored memory relationships
	rels, err := s.relationshipReader.ListByRepository(ctx, repositoryID)
	if err != nil {
		return nil, fmt.Errorf("failed to list relationships: %w", err)
	}
	for _, rel := range rels {
		fromNodeID, existsFrom := entityToNodeID[rel.FromID]
		toNodeID, existsTo := entityToNodeID[rel.ToID]
		if existsFrom && existsTo {
			graphCounts[fromNodeID]++
			graphCounts[toNodeID]++
		}
	}

	// Contributor-expertise stored relationships
	for _, exp := range exps {
		if exp.ContributorID != "" {
			fromNodeID, existsFrom := entityToNodeID[exp.ID]
			toNodeID, existsTo := entityToNodeID[exp.ContributorID]
			if existsFrom && existsTo {
				graphCounts[fromNodeID]++
				graphCounts[toNodeID]++
			}
		}
	}

	// Derived relationships
	derivedRels, err := s.relationshipEngine.Generate(ctx, repositoryID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate derived relationships: %w", err)
	}
	for _, dr := range derivedRels {
		if dr.Edge != nil {
			graphCounts[dr.Edge.FromNodeID]++
			graphCounts[dr.Edge.ToNodeID]++
		}
	}

	var items []*DiscoveryItem

	// 4. Generate DiscoveryItems for Decisions
	for _, d := range decs {
		nodeID := entityToNodeID[d.ID]
		graphCount := graphCounts[nodeID]

		impactPaths := 0
		report, err := s.graphImpactService.AnalyzeDecisionImpact(ctx, repositoryID, d.ID)
		if err == nil && report != nil {
			impactPaths = len(report.ImpactPaths)
		}

		evidence := map[string]int{
			"graph_relationships": graphCount,
			"impact_paths":        impactPaths,
		}
		evBytes, _ := json.Marshal(evidence)

		item := &DiscoveryItem{
			EntityType:   EntityTypeDecision,
			EntityID:     d.ID,
			Score:        float64(graphCount + impactPaths),
			EvidenceJSON: string(evBytes),
			Explanation:  fmt.Sprintf("This decision participates in %d graph relationships and %d impact paths.", graphCount, impactPaths),
		}

		if err := ValidateItem(item); err == nil {
			items = append(items, item)
		}
	}

	// 5. Generate DiscoveryItems for Facts
	for _, f := range facts {
		nodeID := entityToNodeID[f.ID]
		graphCount := graphCounts[nodeID]

		impactPaths := 0
		report, err := s.graphImpactService.AnalyzeFactImpact(ctx, repositoryID, f.ID)
		if err == nil && report != nil {
			impactPaths = len(report.ImpactPaths)
		}

		evidence := map[string]int{
			"graph_relationships": graphCount,
			"impact_paths":        impactPaths,
		}
		evBytes, _ := json.Marshal(evidence)

		item := &DiscoveryItem{
			EntityType:   EntityTypeFact,
			EntityID:     f.ID,
			Score:        float64(graphCount + impactPaths),
			EvidenceJSON: string(evBytes),
			Explanation:  fmt.Sprintf("This fact participates in %d repository knowledge chains.", graphCount+impactPaths),
		}

		if err := ValidateItem(item); err == nil {
			items = append(items, item)
		}
	}

	// 6. Generate DiscoveryItems for Events
	for _, ev := range events {
		nodeID := entityToNodeID[ev.ID]
		graphCount := graphCounts[nodeID]

		impactPaths := 0
		report, err := s.graphImpactService.AnalyzeEventImpact(ctx, repositoryID, ev.ID)
		if err == nil && report != nil {
			impactPaths = len(report.ImpactPaths)
		}

		evidence := map[string]int{
			"graph_relationships": graphCount,
			"impact_paths":        impactPaths,
		}
		evBytes, _ := json.Marshal(evidence)

		item := &DiscoveryItem{
			EntityType:   EntityTypeEvent,
			EntityID:     ev.ID,
			Score:        float64(graphCount + impactPaths),
			EvidenceJSON: string(evBytes),
			Explanation:  fmt.Sprintf("This event participates in %d graph relationships and %d impact paths.", graphCount, impactPaths),
		}

		if err := ValidateItem(item); err == nil {
			items = append(items, item)
		}
	}

	// 7. Generate DiscoveryItems for Contributors
	for _, c := range contribs {
		var cExps []*models.Expertise
		domainsMap := make(map[string]bool)

		for _, exp := range exps {
			if exp.ContributorID == c.ID {
				cExps = append(cExps, exp)
				domainsMap[exp.Domain] = true
			}
		}

		expCount := len(cExps)
		domainCount := len(domainsMap)

		evidence := map[string]int{
			"expertise_count": expCount,
			"domains":         domainCount,
		}
		evBytes, _ := json.Marshal(evidence)

		item := &DiscoveryItem{
			EntityType:   EntityTypeContributor,
			EntityID:     c.ID,
			Score:        float64(expCount + domainCount),
			EvidenceJSON: string(evBytes),
			Explanation:  fmt.Sprintf("This contributor owns expertise across %d repository domains.", domainCount),
		}

		if err := ValidateItem(item); err == nil {
			items = append(items, item)
		}
	}

	// 8. Deterministic sorting: Score desc -> EntityType asc -> EntityID asc
	sort.Slice(items, func(i, j int) bool {
		if items[i].Score != items[j].Score {
			return items[i].Score > items[j].Score
		}
		if items[i].EntityType != items[j].EntityType {
			return items[i].EntityType < items[j].EntityType
		}
		return items[i].EntityID < items[j].EntityID
	})

	return &KnowledgeDiscoveryReport{Items: items}, nil
}
