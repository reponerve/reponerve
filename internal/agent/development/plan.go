package development

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/reponerve/reponerve/internal/agent/workflow"
	agentsearch "github.com/reponerve/reponerve/internal/agent/search"
	codemodels "github.com/reponerve/reponerve/internal/code/models"
	"github.com/reponerve/reponerve/internal/intelligence/learning"
	"github.com/reponerve/reponerve/internal/intelligence/reviewers"
)

const (
	sourceKnowledgeDiscovery     = "knowledge_discovery"
	sourceLearningPaths          = "learning_paths"
	sourceReviewerRecommendations = "reviewer_recommendations"
	sourceChangePlanning         = "change_planning"
	sourceWorkflowIntelligence   = "workflow_intelligence"
)

// Plan prepares implementation guidance by orchestrating upstream planning authorities.
func (s *Service) Plan(ctx context.Context, req DevelopmentRequest) (*DevelopmentPlan, error) {
	task := strings.TrimSpace(req.Topic)
	if task == "" {
		return nil, fmt.Errorf("task cannot be empty")
	}

	topic, err := s.router.ResolveTopic(ctx, req.RepositoryID, NormalizeTaskTopic(task))
	if err != nil {
		return nil, err
	}

	out := &DevelopmentPlan{
		Task:              task,
		SuggestedWorkflow: workflow.WorkflowTypeChangePreparation,
		SourceServices:    []string{sourceRepositorySearch, sourceCodeIntelligence, sourceWorkflowIntelligence},
	}

	services := map[string]struct{}{
		sourceRepositorySearch:     {},
		sourceCodeIntelligence:     {},
		sourceWorkflowIntelligence: {},
	}

	appendEvidence(&out.Evidence, sourceRepositorySearch, "topic_resolution", map[string]string{
		"match_evidence": topic.MatchEvidence,
	})

	entities, err := s.loadCodeEntities(ctx, req.RepositoryID, topic.CodeEntityIDs)
	if err != nil {
		return nil, err
	}
	for _, e := range entities {
		ref := codeEntityRef(e)
		switch e.EntityType {
		case codemodels.EntityTypePackage:
			out.ImpactedAreas = appendUniqueEntityRef(out.ImpactedAreas, ref)
		case codemodels.EntityTypeFile:
			out.ImpactedAreas = appendUniqueEntityRef(out.ImpactedAreas, ref)
			out.StartingPoints = appendUniqueEntityRef(out.StartingPoints, ref)
		case codemodels.EntityTypeStruct, codemodels.EntityTypeInterface,
			codemodels.EntityTypeTypeAlias, codemodels.EntityTypeFunction, codemodels.EntityTypeMethod:
			out.ImpactedAreas = appendUniqueEntityRef(out.ImpactedAreas, ref)
			out.StartingPoints = appendUniqueEntityRef(out.StartingPoints, ref)
		default:
			continue
		}
		appendEvidence(&out.Evidence, sourceCodeIntelligence, "impacted_area", map[string]string{
			"entity_type": e.EntityType, "qualified_name": e.QualifiedName,
		})
	}

	if len(out.StartingPoints) == 0 {
		for _, e := range entities {
			if e.EntityType != codemodels.EntityTypeFile {
				continue
			}
			out.StartingPoints = appendUniqueEntityRef(out.StartingPoints, codeEntityRef(e))
			if len(out.StartingPoints) >= 3 {
				break
			}
		}
	}

	for entityID := range topic.RepositoryHitIDs {
		ref, ev, _, err := s.resolveRepositoryEntity(ctx, req.RepositoryID, entityID)
		if err != nil || ref == nil {
			continue
		}
		switch ref.EntityType {
		case agentsearch.EntityTypeDecision:
			out.RelevantDecisions = appendUniqueEntityRef(out.RelevantDecisions, *ref)
			out.StartingPoints = appendUniqueEntityRef(out.StartingPoints, *ref)
		case agentsearch.EntityTypeFact:
			out.RelevantFacts = appendUniqueEntityRef(out.RelevantFacts, *ref)
		case agentsearch.EntityTypeExpertise:
			out.ImpactedAreas = appendUniqueEntityRef(out.ImpactedAreas, *ref)
		default:
			continue
		}
		out.Evidence = append(out.Evidence, ev...)
	}

	expertise, owners, ev, err := s.matchExpertise(ctx, req.RepositoryID, task)
	if err != nil {
		return nil, err
	}
	out.Owners = appendUniqueEntityRefSlice(out.Owners, owners)
	out.Owners = appendUniqueEntityRefSlice(out.Owners, expertise)
	out.Evidence = append(out.Evidence, ev...)
	if len(expertise) > 0 || len(owners) > 0 {
		services[sourceOwnershipIntelligence] = struct{}{}
	}

	domain := inferPlanningDomain(task, expertise)
	if s.reviewerService != nil {
		services[sourceReviewerRecommendations] = struct{}{}
		var report *reviewers.ReviewerRecommendationReport
		if domain != "" {
			report, err = s.reviewerService.RecommendDomainReviewers(ctx, req.RepositoryID, domain)
		} else {
			report, err = s.reviewerService.RecommendRepositoryReviewers(ctx, req.RepositoryID)
		}
		if err != nil {
			return nil, fmt.Errorf("reviewer recommendations: %w", err)
		}
		labels := s.contributorLabels(ctx, req.RepositoryID)
		for _, rec := range report.Recommendations {
			label := labels[rec.ContributorID]
			if label == "" {
				label = rec.ContributorID
			}
			out.Reviewers = appendUniqueEntityRef(out.Reviewers, EntityRef{
				EntityType: agentsearch.EntityTypeContributor,
				EntityID:   rec.ContributorID,
				Label:      label,
			})
			out.Evidence = append(out.Evidence, EvidenceItem{
				Source:  sourceReviewerRecommendations,
				Type:    "recommendation",
				Payload: json.RawMessage(rec.EvidenceJSON),
			})
		}
	}

	if s.learningService != nil && domain != "" {
		services[sourceLearningPaths] = struct{}{}
		path, err := s.learningService.GenerateDomainPath(ctx, req.RepositoryID, domain)
		if err != nil {
			return nil, fmt.Errorf("learning path: %w", err)
		}
		for _, step := range path.Steps {
			ref, err := s.entityRefFromLearningStep(ctx, req.RepositoryID, step)
			if err != nil || ref == nil {
				continue
			}
			out.StartingPoints = appendUniqueEntityRef(out.StartingPoints, *ref)
			out.Evidence = append(out.Evidence, EvidenceItem{
				Source:  sourceLearningPaths,
				Type:    "learning_step",
				Payload: json.RawMessage(step.EvidenceJSON),
			})
		}
	}

	if s.changePlanService != nil && len(out.RelevantDecisions) > 0 {
		services[sourceChangePlanning] = struct{}{}
		for _, decision := range out.RelevantDecisions {
			cp, err := s.changePlanService.GenerateDecisionPlan(ctx, req.RepositoryID, decision.EntityID)
			if err != nil {
				continue
			}
			for _, item := range cp.Items {
				ref := EntityRef{
					EntityType: item.EntityType,
					EntityID:   item.EntityID,
					Label:      item.EntityID,
				}
				if resolved, _, _, _ := s.resolveRepositoryEntity(ctx, req.RepositoryID, item.EntityID); resolved != nil {
					ref = *resolved
				}
				out.ImpactedAreas = appendUniqueEntityRef(out.ImpactedAreas, ref)
				out.Evidence = append(out.Evidence, EvidenceItem{
					Source:  sourceChangePlanning,
					Type:    "change_plan_item",
					Payload: json.RawMessage(item.EvidenceJSON),
				})
			}
			break
		}
	}

	links, linkEvidence, err := s.buildRepositoryCodeLinkRefs(ctx, req.RepositoryID, topic)
	if err != nil {
		return nil, err
	}
	out.RepositoryCodeLinks = links
	if len(links) > 0 {
		services[sourceRepositoryCodeLinks] = struct{}{}
		out.Evidence = append(out.Evidence, linkEvidence...)
		for _, link := range links {
			out.StartingPoints = appendUniqueEntityRef(out.StartingPoints, link.CodeEntityRef)
		}
	}

	briefingEntities := entities
	if len(briefingEntities) == 0 {
		briefingEntities, err = s.entitiesFromStartingPoints(ctx, req.RepositoryID, out.StartingPoints)
		if err != nil {
			return nil, err
		}
	}
	briefings, err := s.buildEntityBriefings(ctx, req.RepositoryID, task, briefingEntities, out.RepositoryCodeLinks)
	if err != nil {
		return nil, err
	}
	out.EntityBriefings = briefings
	out.SuggestedSteps = buildSuggestedSteps(out)

	out.SourceServices = sortServices(services)
	sortEntityRefs(out.ImpactedAreas)
	sortEntityRefs(out.RelevantDecisions)
	sortEntityRefs(out.RelevantFacts)
	sortEntityRefs(out.Owners)
	sortEntityRefs(out.Reviewers)
	sortEntityRefs(out.StartingPoints)
	sortEvidence(out.Evidence)
	sortRepositoryCodeLinks(out.RepositoryCodeLinks)
	return out, nil
}

func (s *Service) entitiesFromStartingPoints(ctx context.Context, repositoryID string, startingPoints []EntityRef) ([]*codemodels.CodeEntity, error) {
	if s.codeEntityReader == nil || len(startingPoints) == 0 {
		return nil, nil
	}
	seen := make(map[string]struct{})
	var entities []*codemodels.CodeEntity
	for _, sp := range startingPoints {
		if sp.EntityID == "" {
			continue
		}
		if _, ok := seen[sp.EntityID]; ok {
			continue
		}
		entity, err := s.codeEntityReader.GetByID(ctx, sp.EntityID)
		if err != nil || entity == nil {
			continue
		}
		seen[sp.EntityID] = struct{}{}
		entities = append(entities, entity)
	}
	return entities, nil
}

func inferPlanningDomain(task string, expertise []EntityRef) string {
	if len(expertise) > 0 {
		return expertise[0].Label
	}
	terms := extractAskTerms(task)
	if len(terms) == 0 {
		return ""
	}
	return terms[0]
}

func (s *Service) contributorLabels(ctx context.Context, repositoryID string) map[string]string {
	labels := make(map[string]string)
	if s.contributorReader == nil {
		return labels
	}
	contribs, err := s.contributorReader.ListByRepository(ctx, repositoryID)
	if err != nil {
		return labels
	}
	for _, c := range contribs {
		label := strings.TrimSpace(c.Name)
		if label == "" {
			label = strings.TrimSpace(c.Email)
		}
		if label == "" {
			label = c.ID
		}
		labels[c.ID] = label
	}
	return labels
}

func (s *Service) entityRefFromLearningStep(ctx context.Context, repositoryID string, step *learning.LearningStep) (*EntityRef, error) {
	if step == nil {
		return nil, nil
	}
	ref, _, _, err := s.resolveRepositoryEntity(ctx, repositoryID, step.EntityID)
	if err != nil || ref == nil {
		return &EntityRef{
			EntityType: step.EntityType,
			EntityID:   step.EntityID,
			Label:      step.EntityID,
		}, nil
	}
	label := fmt.Sprintf("%s (learning step %d)", ref.Label, step.Position)
	return &EntityRef{
		EntityType: ref.EntityType,
		EntityID:   ref.EntityID,
		Label:      label,
	}, nil
}

func appendUniqueEntityRef(list []EntityRef, ref EntityRef) []EntityRef {
	for _, existing := range list {
		if existing.EntityID == ref.EntityID && existing.EntityType == ref.EntityType {
			return list
		}
	}
	return append(list, ref)
}

func appendUniqueEntityRefSlice(list []EntityRef, refs []EntityRef) []EntityRef {
	for _, ref := range refs {
		list = appendUniqueEntityRef(list, ref)
	}
	return list
}
