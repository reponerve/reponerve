package development

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/reponerve/reponerve/internal/agent/workflow"
	agentsearch "github.com/reponerve/reponerve/internal/agent/search"
	codemodels "github.com/reponerve/reponerve/internal/code/models"
	"github.com/reponerve/reponerve/internal/intelligence/reviewers"
)

// PrepareReview assembles a review guide by orchestrating reviewer, ownership, and search authorities.
func (s *Service) PrepareReview(ctx context.Context, req DevelopmentRequest) (*DevelopmentReviewGuide, error) {
	topic := strings.TrimSpace(req.Topic)
	if topic == "" {
		return nil, fmt.Errorf("topic cannot be empty")
	}

	resolved, err := s.router.ResolveTopic(ctx, req.RepositoryID, topic)
	if err != nil {
		return nil, err
	}

	out := &DevelopmentReviewGuide{
		Topic:             topic,
		SuggestedWorkflow: workflow.WorkflowTypeReviewPreparation,
	}

	services := map[string]struct{}{
		sourceRepositorySearch:     {},
		sourceWorkflowIntelligence: {},
	}

	appendEvidence(&out.Evidence, sourceRepositorySearch, "topic_resolution", map[string]string{
		"match_evidence": resolved.MatchEvidence,
	})

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
			out.AffectedAreas = appendUniqueEntityRef(out.AffectedAreas, ref)
		default:
			continue
		}
		appendEvidence(&out.Evidence, sourceCodeIntelligence, "affected_area", map[string]string{
			"entity_type": e.EntityType, "qualified_name": e.QualifiedName,
		})
	}

	for entityID := range resolved.RepositoryHitIDs {
		ref, ev, _, err := s.resolveRepositoryEntity(ctx, req.RepositoryID, entityID)
		if err != nil || ref == nil {
			continue
		}
		switch ref.EntityType {
		case agentsearch.EntityTypeDecision, agentsearch.EntityTypeFact, agentsearch.EntityTypeEvent:
			out.RelatedKnowledge = appendUniqueEntityRef(out.RelatedKnowledge, *ref)
		case agentsearch.EntityTypeExpertise:
			out.RequiredExpertise = appendUniqueEntityRef(out.RequiredExpertise, *ref)
			out.AffectedAreas = appendUniqueEntityRef(out.AffectedAreas, *ref)
		default:
			continue
		}
		out.Evidence = append(out.Evidence, ev...)
	}

	expertise, _, ev, err := s.matchExpertise(ctx, req.RepositoryID, topic)
	if err != nil {
		return nil, err
	}
	out.RequiredExpertise = appendUniqueEntityRefSlice(out.RequiredExpertise, expertise)
	out.Evidence = append(out.Evidence, ev...)
	if len(expertise) > 0 {
		services[sourceOwnershipIntelligence] = struct{}{}
	}

	domain := inferPlanningDomain(topic, expertise)
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
			display := label
			if rec.Explanation != "" {
				display = fmt.Sprintf("%s (%s, score: %.0f)", label, rec.Explanation, rec.Score)
			}
			out.RecommendedReviewers = appendUniqueEntityRef(out.RecommendedReviewers, EntityRef{
				EntityType: agentsearch.EntityTypeContributor,
				EntityID:   rec.ContributorID,
				Label:      display,
			})
			out.Evidence = append(out.Evidence, EvidenceItem{
				Source:  sourceReviewerRecommendations,
				Type:    "recommendation",
				Payload: json.RawMessage(rec.EvidenceJSON),
			})
		}
	}

	links, linkEvidence, err := s.buildRepositoryCodeLinkRefs(ctx, req.RepositoryID, resolved)
	if err != nil {
		return nil, err
	}
	out.RepositoryCodeLinks = links
	if len(links) > 0 {
		services[sourceRepositoryCodeLinks] = struct{}{}
		out.Evidence = append(out.Evidence, linkEvidence...)
		for _, link := range links {
			out.AffectedAreas = appendUniqueEntityRef(out.AffectedAreas, link.CodeEntityRef)
			out.RelatedKnowledge = appendUniqueEntityRef(out.RelatedKnowledge, link.RepositoryEntityRef)
		}
	}

	out.SourceServices = sortServices(services)
	sortEntityRefs(out.RecommendedReviewers)
	sortEntityRefs(out.RequiredExpertise)
	sortEntityRefs(out.AffectedAreas)
	sortEntityRefs(out.RelatedKnowledge)
	sortEvidence(out.Evidence)
	sortRepositoryCodeLinks(out.RepositoryCodeLinks)
	appendReviewDiscipline(out)
	return out, nil
}
