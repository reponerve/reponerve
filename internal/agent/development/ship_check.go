package development

import (
	"context"
	"fmt"
	"strconv"
	"strings"
)

// ShipCheck assembles pre-ship blockers and advisories from review and impact evidence.
func (s *Service) ShipCheck(ctx context.Context, req DevelopmentRequest) (*ShipCheckResult, error) {
	topic := strings.TrimSpace(req.Topic)
	if topic == "" {
		return nil, fmt.Errorf("topic cannot be empty")
	}

	review, err := s.PrepareReview(ctx, req)
	if err != nil {
		return nil, err
	}
	impact, err := s.AnalyzeImpact(ctx, req)
	if err != nil {
		return nil, err
	}

	out := &ShipCheckResult{
		Topic:                topic,
		ImpactedAreas:        mergeEntityRefs(review.AffectedAreas, impact.DependentAreas),
		RelatedKnowledge:     review.RelatedKnowledge,
		RecommendedReviewers: review.RecommendedReviewers,
		RecommendedNextTools: []string{"review", "analyze_topic_impact", "explain_file"},
		SourceServices:       mergeSourceServices(review.SourceServices, impact.SourceServices, []string{sourceDevelopmentDiscipline}),
	}
	out.Evidence = append(out.Evidence, review.Evidence...)
	out.Evidence = append(out.Evidence, impact.Evidence...)

	out.ShipBlockers = shipBlockersFromEvidence(topic, review, impact)
	out.Advisories = shipAdvisoriesFromEvidence(review, impact)

	if len(out.ShipBlockers) > 0 {
		out.RecommendedNextTools = []string{"review", "analyze_topic_impact"}
	}

	appendEvidence(&out.Evidence, sourceDevelopmentDiscipline, "ship_check", map[string]string{
		"blocker_count":  strconv.Itoa(len(out.ShipBlockers)),
		"advisory_count": strconv.Itoa(len(out.Advisories)),
	})

	return out, nil
}

func shipBlockersFromEvidence(
	topic string,
	review *DevelopmentReviewGuide,
	impact *DevelopmentImpactReport,
) []ShipCheckItem {
	var blockers []ShipCheckItem
	refs := mergeEntityRefs(review.RelatedKnowledge, impact.ImpactedDecisions)

	for _, ref := range refs {
		if !shipKeywordMatch(ref.Label, shipBlockerKeywords) {
			continue
		}
		blockers = append(blockers, ShipCheckItem{
			Severity: "blocker",
			Category: "migration",
			Message:  "Related ADR or knowledge mentions migration/schema change — confirm migration strategy before ship",
			Related:  []EntityRef{ref},
		})
	}

	lowerTopic := strings.ToLower(topic)
	if shipKeywordMatch(lowerTopic, shipBlockerKeywords) && len(blockers) == 0 {
		blockers = append(blockers, ShipCheckItem{
			Severity: "blocker",
			Category: "migration",
			Message:  "Change topic implies schema or migration work — document migration and rollback before ship",
		})
	}

	return blockers
}

func shipAdvisoriesFromEvidence(
	review *DevelopmentReviewGuide,
	impact *DevelopmentImpactReport,
) []ShipCheckItem {
	var advisories []ShipCheckItem

	if len(review.RequiredExpertise) > 0 && len(review.RecommendedReviewers) == 0 {
		advisories = append(advisories, ShipCheckItem{
			Severity: "advisory",
			Category: "ownership",
			Message:  "Domain expertise detected but no reviewers recommended — confirm owners before merge",
			Related:  review.RequiredExpertise,
		})
	}

	if len(impact.DependentAreas) >= 3 {
		advisories = append(advisories, ShipCheckItem{
			Severity: "advisory",
			Category: "impact",
			Message:  "Wide blast radius across multiple dependent areas — run targeted verification",
			Related:  impact.DependentAreas,
		})
	}

	if len(review.RelatedKnowledge) > 0 {
		advisories = append(advisories, ShipCheckItem{
			Severity: "advisory",
			Category: "knowledge",
			Message:  "Review related decisions and facts in structured output before ship",
			Related:  review.RelatedKnowledge,
		})
	}

	if len(impact.DependentAreas) > 0 || len(review.AffectedAreas) > 0 {
		advisories = append(advisories, ShipCheckItem{
			Severity: "advisory",
			Category: "rollback",
			Message:  "Define rollback or feature-flag plan for production-impacting changes",
		})
	}

	return advisories
}

var shipBlockerKeywords = []string{
	"migration", "schema", "database", "breaking", "deprecat", "ddl",
}

func shipKeywordMatch(text string, keywords []string) bool {
	lower := strings.ToLower(text)
	for _, kw := range keywords {
		if strings.Contains(lower, kw) {
			return true
		}
	}
	return false
}

func mergeEntityRefs(slices ...[]EntityRef) []EntityRef {
	var out []EntityRef
	for _, slice := range slices {
		out = appendUniqueEntityRefSlice(out, slice)
	}
	return out
}

func mergeSourceServices(slices ...[]string) []string {
	seen := make(map[string]struct{})
	var out []string
	for _, slice := range slices {
		for _, s := range slice {
			if s == "" {
				continue
			}
			if _, ok := seen[s]; ok {
				continue
			}
			seen[s] = struct{}{}
			out = append(out, s)
		}
	}
	return out
}
