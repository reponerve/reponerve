package relationships

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"reponerve/internal/graph/model"
	"reponerve/internal/query/storage"
	models "reponerve/pkg/models"
)

// Engine derives graph relationships from existing repository memory and ownership intelligence.
type Engine struct {
	decisionReader     storage.DecisionReader
	intentReader       storage.IntentReader
	factReader         storage.FactReader
	eventReader        storage.EventReader
	relationshipReader storage.RelationshipReader
	contribReader      storage.ContributorReader
	expertiseReader    storage.ExpertiseReader
	sourceReader       storage.SourceReader
}

// NewEngine creates a new Graph Relationship Engine instance.
func NewEngine(
	dr storage.DecisionReader,
	ir storage.IntentReader,
	fr storage.FactReader,
	er storage.EventReader,
	rr storage.RelationshipReader,
	cr storage.ContributorReader,
	expr storage.ExpertiseReader,
	sr storage.SourceReader,
) *Engine {
	return &Engine{
		decisionReader:     dr,
		intentReader:       ir,
		factReader:         fr,
		eventReader:        er,
		relationshipReader: rr,
		contribReader:      cr,
		expertiseReader:    expr,
		sourceReader:       sr,
	}
}

func (e *Engine) DecisionReader() storage.DecisionReader { return e.decisionReader }
func (e *Engine) IntentReader() storage.IntentReader         { return e.intentReader }
func (e *Engine) FactReader() storage.FactReader             { return e.factReader }
func (e *Engine) EventReader() storage.EventReader           { return e.eventReader }
func (e *Engine) RelationshipReader() storage.RelationshipReader {
	return e.relationshipReader
}
func (e *Engine) ContributorReader() storage.ContributorReader { return e.contribReader }
func (e *Engine) ExpertiseReader() storage.ExpertiseReader     { return e.expertiseReader }
func (e *Engine) SourceReader() storage.SourceReader           { return e.sourceReader }

type adrMetadata struct {
	Content string `json:"content"`
	Status  string `json:"status"`
	Path    string `json:"path"`
}

type decisionDependencyEvidence struct {
	MatchType string `json:"match_type"`
	Value     string `json:"value"`
}

type factSupportEvidence struct {
	MatchingValue string `json:"matching_value"`
}

type domainRelationEvidence struct {
	ContributorID    string  `json:"contributor_id"`
	ContributorName  string  `json:"contributor_name"`
	ContributorEmail string  `json:"contributor_email"`
	DomainA          string  `json:"domain_a"`
	ScoreA           float64 `json:"score_a"`
	DomainB          string  `json:"domain_b"`
	ScoreB           float64 `json:"score_b"`
}

// Generate scans repository memory and extracts all deterministic derived relationships.
func (e *Engine) Generate(ctx context.Context, repositoryID string) ([]*DerivedRelationship, error) {
	var list []*DerivedRelationship

	// Load memory elements
	allDecisions, err := e.decisionReader.ListByRepository(ctx, repositoryID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch decisions: %w", err)
	}

	allSources, err := e.sourceReader.ListByRepository(ctx, repositoryID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch sources: %w", err)
	}
	sourcesMap := make(map[string]*models.Source)
	for _, src := range allSources {
		sourcesMap[src.ID] = src
	}

	allFacts, err := e.factReader.ListByRepository(ctx, repositoryID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch facts: %w", err)
	}

	allExpertise, err := e.expertiseReader.ListByRepository(ctx, repositoryID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch expertise: %w", err)
	}

	allContributors, err := e.contribReader.ListByRepository(ctx, repositoryID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch contributors: %w", err)
	}
	contribsMap := make(map[string]*models.Contributor)
	for _, c := range allContributors {
		contribsMap[c.ID] = c
	}

	// 1. Derive DECISION_DEPENDS_ON_DECISION relationships
	for _, decA := range allDecisions {
		srcA, exists := sourcesMap[decA.SourceID]
		if !exists || srcA.MetadataJSON == "" {
			continue
		}

		var meta adrMetadata
		if err := json.Unmarshal([]byte(srcA.MetadataJSON), &meta); err != nil || meta.Content == "" {
			continue
		}

		contentLower := strings.ToLower(meta.Content)

		for _, decB := range allDecisions {
			if decA.ID == decB.ID {
				continue
			}

			srcB := sourcesMap[decB.SourceID]
			matched := false
			matchType := ""
			matchVal := ""

			idLower := strings.ToLower(decB.ID)
			refLower := ""
			if srcB != nil {
				refLower = strings.ToLower(srcB.Reference)
			}

			// Prefix checks for explicit references (depends on, see, depends on:) or bracketed markdown references
			containsExplicitLinkToID := strings.Contains(contentLower, "["+idLower+"]")
			containsExplicitLinkToRef := refLower != "" && strings.Contains(contentLower, "]("+refLower+")")

			containsPrefixToID := strings.Contains(contentLower, "see "+idLower) ||
				strings.Contains(contentLower, "depends on "+idLower) ||
				strings.Contains(contentLower, "depends on: "+idLower)

			containsPrefixToRef := refLower != "" && (
				strings.Contains(contentLower, "see "+refLower) ||
				strings.Contains(contentLower, "depends on "+refLower) ||
				strings.Contains(contentLower, "depends on: "+refLower))

			if containsExplicitLinkToID || containsExplicitLinkToRef || containsPrefixToID || containsPrefixToRef {
				matched = true
				matchType = "explicit_reference"
				if containsExplicitLinkToID || containsPrefixToID {
					matchVal = decB.ID
				} else {
					matchVal = srcB.Reference
				}
			} else if refLower != "" && strings.Contains(contentLower, refLower) {
				matched = true
				matchType = "source_reference"
				matchVal = srcB.Reference
			} else if strings.Contains(contentLower, idLower) {
				matched = true
				matchType = "decision_id"
				matchVal = decB.ID
			}

			if matched {
				evidence := decisionDependencyEvidence{
					MatchType: matchType,
					Value:     matchVal,
				}
				evBytes, err := json.Marshal(evidence)
				if err != nil {
					return nil, fmt.Errorf("failed to marshal decision dependency evidence: %w", err)
				}

				explanation := fmt.Sprintf("Decision %s depends on Decision %s because the source of Decision %s contains a reference of type %q: %q.",
					decA.Title, decB.Title, decA.Title, matchType, matchVal)

				fromNodeID := model.NodeID(repositoryID, model.NodeTypeDecision, decA.ID)
				toNodeID := model.NodeID(repositoryID, model.NodeTypeDecision, decB.ID)

				edge := model.NewEdge(repositoryID, fromNodeID, toNodeID, "DECISION_DEPENDS_ON_DECISION", model.CategoryDerived, string(evBytes))

				list = append(list, &DerivedRelationship{
					Edge:        edge,
					Evidence:    evBytes,
					Explanation: explanation,
				})
			}
		}
	}

	// 2. Derive FACT_SUPPORTS_FACT relationships
	for _, factA := range allFacts {
		if factA.Object == "" {
			continue
		}
		for _, factB := range allFacts {
			if factA.ID == factB.ID || factB.Subject == "" {
				continue
			}

			if strings.EqualFold(factA.Object, factB.Subject) {
				evidence := factSupportEvidence{
					MatchingValue: factA.Object,
				}
				evBytes, err := json.Marshal(evidence)
				if err != nil {
					return nil, fmt.Errorf("failed to marshal fact support evidence: %w", err)
				}

				explanation := fmt.Sprintf("Fact %s supports Fact %s because the object of Fact A (%q) matches the subject of Fact B (%q).",
					factA.ID, factB.ID, factA.Object, factB.Subject)

				fromNodeID := model.NodeID(repositoryID, model.NodeTypeFact, factA.ID)
				toNodeID := model.NodeID(repositoryID, model.NodeTypeFact, factB.ID)

				edge := model.NewEdge(repositoryID, fromNodeID, toNodeID, "FACT_SUPPORTS_FACT", model.CategoryDerived, string(evBytes))

				list = append(list, &DerivedRelationship{
					Edge:        edge,
					Evidence:    evBytes,
					Explanation: explanation,
				})
			}
		}
	}

	// 3. Derive DOMAIN_RELATES_TO_DOMAIN relationships
	// Group expertise records by ContributorID
	contribExps := make(map[string][]*models.Expertise)
	for _, exp := range allExpertise {
		contribExps[exp.ContributorID] = append(contribExps[exp.ContributorID], exp)
	}

	for cID, exps := range contribExps {
		c, hasContrib := contribsMap[cID]
		var cName, cEmail string
		var ident string = cID
		if hasContrib {
			cName = c.Name
			cEmail = c.Email
			if c.Name != "" {
				ident = c.Name
			} else if c.Email != "" {
				ident = c.Email
			}
		}

		for i := 0; i < len(exps); i++ {
			for j := i + 1; j < len(exps); j++ {
				expA := exps[i]
				expB := exps[j]

				// Ensure we only link different domains and sort alphabetically to prevent duplicates
				if expA.Domain == expB.Domain {
					continue
				}

				first := expA
				second := expB
				if expA.Domain > expB.Domain {
					first = expB
					second = expA
				}

				evidence := domainRelationEvidence{
					ContributorID:    cID,
					ContributorName:  cName,
					ContributorEmail: cEmail,
					DomainA:          first.Domain,
					ScoreA:           first.Score,
					DomainB:          second.Domain,
					ScoreB:           second.Score,
				}
				evBytes, err := json.Marshal(evidence)
				if err != nil {
					return nil, fmt.Errorf("failed to marshal domain relation evidence: %w", err)
				}

				explanation := fmt.Sprintf("Domain %q relates to Domain %q because contributor %s has active expertise in both domains.",
					first.Domain, second.Domain, ident)

				fromNodeID := model.NodeID(repositoryID, model.NodeTypeExpertise, first.ID)
				toNodeID := model.NodeID(repositoryID, model.NodeTypeExpertise, second.ID)

				edge := model.NewEdge(repositoryID, fromNodeID, toNodeID, "DOMAIN_RELATES_TO_DOMAIN", model.CategoryDerived, string(evBytes))

				list = append(list, &DerivedRelationship{
					Edge:        edge,
					Evidence:    evBytes,
					Explanation: explanation,
				})
			}
		}
	}

	// Validate all generated relationships
	for _, rel := range list {
		if err := validateDerivedRelationship(rel); err != nil {
			return nil, fmt.Errorf("generated relationship validation failed: %w", err)
		}
	}

	// Sort deterministically: EdgeType ascending -> FromNodeID ascending -> ToNodeID ascending
	sort.Slice(list, func(i, j int) bool {
		if list[i].Edge.EdgeType != list[j].Edge.EdgeType {
			return list[i].Edge.EdgeType < list[j].Edge.EdgeType
		}
		if list[i].Edge.FromNodeID != list[j].Edge.FromNodeID {
			return list[i].Edge.FromNodeID < list[j].Edge.FromNodeID
		}
		return list[i].Edge.ToNodeID < list[j].Edge.ToNodeID
	})

	return list, nil
}

func validateDerivedRelationship(rel *DerivedRelationship) error {
	if rel == nil {
		return fmt.Errorf("derived relationship is nil")
	}
	if rel.Edge == nil {
		return fmt.Errorf("missing edge")
	}
	if err := model.ValidateEdge(rel.Edge); err != nil {
		return fmt.Errorf("invalid edge: %w", err)
	}
	if len(rel.Evidence) == 0 {
		return fmt.Errorf("missing evidence")
	}
	if rel.Explanation == "" {
		return fmt.Errorf("missing explanation")
	}
	return nil
}
