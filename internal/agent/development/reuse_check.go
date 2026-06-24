package development

import (
	"context"
	"fmt"
	"sort"
	"strings"

	agentsearch "github.com/reponerve/reponerve/internal/agent/search"
	codemodels "github.com/reponerve/reponerve/internal/code/models"
)

const sourceDevelopmentDiscipline = "development_discipline"

// ReuseCheck finds existing symbols and knowledge to prefer before writing new code.
func (s *Service) ReuseCheck(ctx context.Context, req DevelopmentRequest) (*ReuseCheckResult, error) {
	intent := strings.TrimSpace(req.Topic)
	if intent == "" {
		return nil, fmt.Errorf("intent cannot be empty")
	}

	resolved, err := s.router.ResolveTopic(ctx, req.RepositoryID, NormalizeTaskTopic(intent))
	if err != nil {
		return nil, err
	}

	out := &ReuseCheckResult{
		Intent: intent,
		RecommendedNextTools: []string{
			"explain_function", "explain_file", "plan",
		},
		SourceServices: []string{sourceDevelopmentDiscipline, sourceRepositorySearch, sourceCodeIntelligence},
	}

	appendEvidence(&out.Evidence, sourceRepositorySearch, "topic_resolution", map[string]string{
		"match_evidence": resolved.MatchEvidence,
	})

	entities, err := s.loadCodeEntities(ctx, req.RepositoryID, resolved.CodeEntityIDs)
	if err != nil {
		return nil, err
	}

	links := repositoryCodeLinkRefs(resolved.RepositoryCodeLinks)
	briefings, err := s.buildEntityBriefings(ctx, req.RepositoryID, intent, entities, links)
	if err != nil {
		return nil, err
	}

	seen := make(map[string]struct{})
	for _, b := range briefings {
		key := b.QualifiedName + "|" + b.EntityType
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out.ReuseCandidates = append(out.ReuseCandidates, ReuseCandidate{
			QualifiedName: b.QualifiedName,
			EntityType:    b.EntityType,
			DefinedIn:     b.DefinedIn,
			Role:          b.Role,
			Rank:          reuseRank(b.EntityType),
		})
		appendEvidence(&out.Evidence, sourceCodeIntelligence, "reuse_candidate", map[string]string{
			"qualified_name": b.QualifiedName,
			"entity_type":    b.EntityType,
			"defined_in":     b.DefinedIn,
		})
	}

	for entityID := range resolved.RepositoryHitIDs {
		ref, ev, _, err := s.resolveRepositoryEntity(ctx, req.RepositoryID, entityID)
		if err != nil || ref == nil {
			continue
		}
		if ref.EntityType == agentsearch.EntityTypeDecision {
			out.RelatedDecisions = appendUniqueEntityRef(out.RelatedDecisions, *ref)
			key := ref.EntityID + "|decision"
			if _, ok := seen[key]; !ok {
				seen[key] = struct{}{}
				out.ReuseCandidates = append(out.ReuseCandidates, ReuseCandidate{
					QualifiedName: ref.Label,
					EntityType:    ref.EntityType,
					DefinedIn:     ref.EntityID,
					Role:          "existing decision pattern",
					Rank:          10,
				})
			}
		}
		out.Evidence = append(out.Evidence, ev...)
	}

	sort.Slice(out.ReuseCandidates, func(i, j int) bool {
		if out.ReuseCandidates[i].Rank != out.ReuseCandidates[j].Rank {
			return out.ReuseCandidates[i].Rank < out.ReuseCandidates[j].Rank
		}
		return strings.ToLower(out.ReuseCandidates[i].QualifiedName) <
			strings.ToLower(out.ReuseCandidates[j].QualifiedName)
	})

	if len(out.ReuseCandidates) == 0 {
		out.RecommendedNextTools = []string{"ask", "plan", "explain"}
		appendEvidence(&out.Evidence, sourceDevelopmentDiscipline, "reuse_status", map[string]string{
			"status": "no_candidates",
		})
	} else {
		appendEvidence(&out.Evidence, sourceDevelopmentDiscipline, "reuse_status", map[string]string{
			"status":     "candidates_found",
			"candidate_count": fmt.Sprintf("%d", len(out.ReuseCandidates)),
		})
	}

	return out, nil
}

func reuseRank(entityType string) int {
	switch entityType {
	case codemodels.EntityTypeFunction, codemodels.EntityTypeMethod:
		return 1
	case codemodels.EntityTypeStruct, codemodels.EntityTypeInterface:
		return 2
	case codemodels.EntityTypeTypeAlias:
		return 3
	case codemodels.EntityTypeFile:
		return 4
	case codemodels.EntityTypePackage:
		return 5
	default:
		return 6
	}
}

func repositoryCodeLinkRefs(links []*codemodels.RepositoryCodeRelationship) []RepositoryCodeLinkRef {
	if len(links) == 0 {
		return nil
	}
	out := make([]RepositoryCodeLinkRef, 0, len(links))
	for _, link := range links {
		if link == nil {
			continue
		}
		out = append(out, RepositoryCodeLinkRef{
			RelationshipType: link.RelationshipType,
			RepositoryEntityRef: EntityRef{
				EntityType: link.RepositoryEntityType,
				EntityID:   link.RepositoryEntityID,
			},
			CodeEntityRef: EntityRef{
				EntityType: link.CodeEntityType,
				EntityID:   link.CodeEntityID,
			},
			EvidenceJSON: link.EvidenceJSON,
		})
	}
	return out
}
