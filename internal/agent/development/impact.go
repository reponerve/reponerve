package development

import (
	"context"
	"fmt"
	"sort"
	"strings"

	agentsearch "github.com/reponerve/reponerve/internal/agent/search"
	codemodels "github.com/reponerve/reponerve/internal/code/models"
	graphimpact "github.com/reponerve/reponerve/internal/graph/impact"
	"github.com/reponerve/reponerve/internal/graph/model"
)

const (
	sourceKnowledgeGraphImpact = "knowledge_graph_impact"
	sourceAgentImpactAnalysis  = "agent_impact_analysis"
)

// AnalyzeImpact assembles repository and code impact by orchestrating graph, agent, and code authorities.
func (s *Service) AnalyzeImpact(ctx context.Context, req DevelopmentRequest) (*DevelopmentImpactReport, error) {
	subject := strings.TrimSpace(req.Topic)
	if subject == "" {
		return nil, fmt.Errorf("subject cannot be empty")
	}

	resolved, err := s.router.ResolveTopic(ctx, req.RepositoryID, subject)
	if err != nil {
		return nil, err
	}

	out := &DevelopmentImpactReport{Subject: subject}
	services := map[string]struct{}{
		sourceRepositorySearch: {},
	}
	seenDecisions := make(map[string]struct{})
	seenFacts := make(map[string]struct{})
	seenEvents := make(map[string]struct{})

	appendEvidence(&out.Evidence, sourceRepositorySearch, "topic_resolution", map[string]string{
		"match_evidence": resolved.MatchEvidence,
	})

	seedIDs := make(map[string]struct{}, len(resolved.RepositoryHitIDs))
	for id := range resolved.RepositoryHitIDs {
		seedIDs[id] = struct{}{}
	}
	for _, link := range resolved.RepositoryCodeLinks {
		seedIDs[link.RepositoryEntityID] = struct{}{}
	}

	for entityID := range seedIDs {
		ref, ev, _, err := s.resolveRepositoryEntity(ctx, req.RepositoryID, entityID)
		if err != nil || ref == nil {
			continue
		}
		s.classifyImpactedEntity(out, seenDecisions, seenFacts, seenEvents, *ref)
		out.Evidence = append(out.Evidence, ev...)

		if err := s.appendGraphImpact(ctx, req.RepositoryID, ref.EntityType, ref.EntityID, out, services, seenDecisions, seenFacts, seenEvents); err != nil {
			return nil, err
		}
		if ref.EntityType == agentsearch.EntityTypeDecision {
			if err := s.appendAgentDecisionImpact(ctx, ref.EntityID, out, services, seenFacts, seenEvents); err != nil {
				return nil, err
			}
		}
	}

	entities, err := s.loadCodeEntities(ctx, req.RepositoryID, resolved.CodeEntityIDs)
	if err != nil {
		return nil, err
	}
	if len(entities) > 0 {
		services[sourceCodeIntelligence] = struct{}{}
	}
	for _, e := range entities {
		ref := codeEntityRef(e)
		switch e.EntityType {
		case codemodels.EntityTypePackage, codemodels.EntityTypeFile:
			out.DependentAreas = appendUniqueEntityRef(out.DependentAreas, ref)
		default:
			continue
		}
		appendEvidence(&out.Evidence, sourceCodeIntelligence, "dependent_area", map[string]string{
			"entity_type": e.EntityType, "qualified_name": e.QualifiedName,
		})
	}

	for _, e := range entities {
		if s.relReader == nil {
			break
		}
		outbound, err := s.relReader.ListByFromEntity(ctx, e.ID)
		if err != nil {
			return nil, err
		}
		for _, rel := range outbound {
			label := rel.RelationshipType
			if toEntity, err := s.codeEntityReader.GetByID(ctx, rel.ToEntityID); err == nil && toEntity != nil {
				label = fmt.Sprintf("%s (%s %s)", toEntity.QualifiedName, rel.RelationshipType, toEntity.QualifiedName)
			}
			out.CodeDependencies = appendUniqueRelationshipRef(out.CodeDependencies, RelationshipRef{
				RelationshipType: rel.RelationshipType,
				FromEntityID:     rel.FromEntityID,
				ToEntityID:       rel.ToEntityID,
				Label:            label,
				EvidenceJSON:     rel.EvidenceJSON,
			})
			appendEvidence(&out.Evidence, sourceCodeIntelligence, "depends_on", map[string]string{
				"from_entity_id": rel.FromEntityID,
				"to_entity_id":   rel.ToEntityID,
				"relationship":   rel.RelationshipType,
			})
			if toEntity, err := s.codeEntityReader.GetByID(ctx, rel.ToEntityID); err == nil && toEntity != nil {
				out.DependentAreas = appendUniqueEntityRef(out.DependentAreas, codeEntityRef(toEntity))
			}
		}
	}

	_, owners, ev, err := s.matchExpertise(ctx, req.RepositoryID, subject)
	if err != nil {
		return nil, err
	}
	out.Owners = appendUniqueEntityRefSlice(out.Owners, owners)
	out.Evidence = append(out.Evidence, ev...)
	if len(out.Owners) > 0 {
		services[sourceOwnershipIntelligence] = struct{}{}
	}

	links, linkEvidence, err := s.buildRepositoryCodeLinkRefs(ctx, req.RepositoryID, resolved)
	if err != nil {
		return nil, err
	}
	out.RepositoryCodeLinks = links
	if len(links) > 0 {
		services[sourceRepositoryCodeLinks] = struct{}{}
		out.Evidence = append(out.Evidence, linkEvidence...)
	}

	out.SourceServices = sortServices(services)
	sortEntityRefs(out.ImpactedDecisions)
	sortEntityRefs(out.ImpactedFacts)
	sortEntityRefs(out.ImpactedEvents)
	sortEntityRefs(out.DependentAreas)
	sortEntityRefs(out.Owners)
	sortRelationshipRefs(out.CodeDependencies)
	sortEvidence(out.Evidence)
	sortRepositoryCodeLinks(out.RepositoryCodeLinks)
	return out, nil
}

func (s *Service) classifyImpactedEntity(
	out *DevelopmentImpactReport,
	seenDecisions, seenFacts, seenEvents map[string]struct{},
	ref EntityRef,
) {
	switch ref.EntityType {
	case agentsearch.EntityTypeDecision:
		if _, ok := seenDecisions[ref.EntityID]; ok {
			return
		}
		seenDecisions[ref.EntityID] = struct{}{}
		out.ImpactedDecisions = append(out.ImpactedDecisions, ref)
	case agentsearch.EntityTypeFact:
		if _, ok := seenFacts[ref.EntityID]; ok {
			return
		}
		seenFacts[ref.EntityID] = struct{}{}
		out.ImpactedFacts = append(out.ImpactedFacts, ref)
	case agentsearch.EntityTypeEvent:
		if _, ok := seenEvents[ref.EntityID]; ok {
			return
		}
		seenEvents[ref.EntityID] = struct{}{}
		out.ImpactedEvents = append(out.ImpactedEvents, ref)
	}
}

func (s *Service) appendGraphImpact(
	ctx context.Context,
	repositoryID, entityType, entityID string,
	out *DevelopmentImpactReport,
	services map[string]struct{},
	seenDecisions, seenFacts, seenEvents map[string]struct{},
) error {
	if s.graphImpactService == nil {
		return nil
	}

	var report *graphimpact.ImpactReport
	var err error
	switch entityType {
	case agentsearch.EntityTypeDecision:
		report, err = s.graphImpactService.AnalyzeDecisionImpact(ctx, repositoryID, entityID)
	case agentsearch.EntityTypeFact:
		report, err = s.graphImpactService.AnalyzeFactImpact(ctx, repositoryID, entityID)
	case agentsearch.EntityTypeEvent:
		report, err = s.graphImpactService.AnalyzeEventImpact(ctx, repositoryID, entityID)
	default:
		return nil
	}
	if err != nil {
		return fmt.Errorf("graph impact: %w", err)
	}
	if report == nil || len(report.ImpactPaths) == 0 {
		return nil
	}

	services[sourceKnowledgeGraphImpact] = struct{}{}
	for _, impactPath := range report.ImpactPaths {
		if impactPath == nil || impactPath.Path == nil {
			continue
		}
		depth := len(impactPath.Path.Edges)
		for _, node := range impactPath.Path.Nodes {
			if node == nil {
				continue
			}
			ref := graphNodeEntityRef(node)
			s.classifyImpactedEntity(out, seenDecisions, seenFacts, seenEvents, ref)
			if ref.EntityType == agentsearch.EntityTypeExpertise {
				out.DependentAreas = appendUniqueEntityRef(out.DependentAreas, ref)
			}
			appendEvidence(&out.Evidence, sourceKnowledgeGraphImpact, "impact_chain", map[string]any{
				"depth":       depth,
				"entity_type": node.NodeType,
				"entity_id":   node.EntityID,
				"reason":      impactPath.Reason,
			})
		}
	}
	return nil
}

func (s *Service) appendAgentDecisionImpact(
	ctx context.Context,
	decisionID string,
	out *DevelopmentImpactReport,
	services map[string]struct{},
	seenFacts, seenEvents map[string]struct{},
) error {
	if s.agentImpactService == nil {
		return nil
	}

	report, err := s.agentImpactService.AnalyzeDecisionImpact(ctx, decisionID)
	if err != nil {
		return fmt.Errorf("agent impact: %w", err)
	}
	if report == nil {
		return nil
	}

	services[sourceAgentImpactAnalysis] = struct{}{}
	for _, fact := range report.Facts {
		ref := EntityRef{
			EntityType: agentsearch.EntityTypeFact,
			EntityID:   fact.ID,
			Label:      strings.TrimSpace(strings.Join([]string{fact.Subject, fact.Predicate, fact.Object}, " ")),
		}
		s.classifyImpactedEntity(out, map[string]struct{}{}, seenFacts, seenEvents, ref)
		appendEvidence(&out.Evidence, sourceAgentImpactAnalysis, "related_fact", map[string]string{
			"decision_id": decisionID, "fact_id": fact.ID,
		})
	}
	for _, event := range report.Events {
		ref := EntityRef{
			EntityType: agentsearch.EntityTypeEvent,
			EntityID:   event.ID,
			Label:      event.Title,
		}
		s.classifyImpactedEntity(out, map[string]struct{}{}, seenFacts, seenEvents, ref)
		appendEvidence(&out.Evidence, sourceAgentImpactAnalysis, "related_event", map[string]string{
			"decision_id": decisionID, "event_id": event.ID,
		})
	}
	return nil
}

func graphNodeEntityRef(node *model.GraphNode) EntityRef {
	entityType := string(node.NodeType)
	switch node.NodeType {
	case model.NodeTypeDecision:
		entityType = agentsearch.EntityTypeDecision
	case model.NodeTypeFact:
		entityType = agentsearch.EntityTypeFact
	case model.NodeTypeEvent:
		entityType = agentsearch.EntityTypeEvent
	case model.NodeTypeContributor:
		entityType = agentsearch.EntityTypeContributor
	case model.NodeTypeExpertise:
		entityType = agentsearch.EntityTypeExpertise
	}
	return EntityRef{
		EntityType: entityType,
		EntityID:   node.EntityID,
		Label:      node.EntityID,
	}
}

func appendUniqueRelationshipRef(list []RelationshipRef, ref RelationshipRef) []RelationshipRef {
	for _, existing := range list {
		if existing.RelationshipType == ref.RelationshipType &&
			existing.FromEntityID == ref.FromEntityID &&
			existing.ToEntityID == ref.ToEntityID {
			return list
		}
	}
	return append(list, ref)
}

func sortRelationshipRefs(refs []RelationshipRef) {
	sort.Slice(refs, func(i, j int) bool {
		if refs[i].RelationshipType != refs[j].RelationshipType {
			return refs[i].RelationshipType < refs[j].RelationshipType
		}
		if refs[i].FromEntityID != refs[j].FromEntityID {
			return refs[i].FromEntityID < refs[j].FromEntityID
		}
		return refs[i].ToEntityID < refs[j].ToEntityID
	})
}
