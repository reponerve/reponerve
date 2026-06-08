package reviewers

import (
	"encoding/json"
	"fmt"
)

// ReviewerRecommendation represents a suggested reviewer with deterministic score and evidence.
type ReviewerRecommendation struct {
	ContributorID string  `json:"contributor_id"`
	Score         float64 `json:"score"`
	EvidenceJSON  string  `json:"evidence_json"`
	Explanation   string  `json:"explanation"`
}

// ReviewerRecommendationReport collects the generated reviewer recommendations.
type ReviewerRecommendationReport struct {
	Recommendations []*ReviewerRecommendation `json:"recommendations"`
}

// ValidateRecommendation ensures a ReviewerRecommendation has valid fields.
func ValidateRecommendation(recommendation *ReviewerRecommendation) error {
	if recommendation == nil {
		return fmt.Errorf("recommendation is nil")
	}
	if recommendation.ContributorID == "" {
		return fmt.Errorf("missing contributor ID")
	}
	if recommendation.Score < 0 {
		return fmt.Errorf("invalid score: %f (must be >= 0)", recommendation.Score)
	}
	if recommendation.EvidenceJSON == "" {
		return fmt.Errorf("missing evidence")
	}
	if !json.Valid([]byte(recommendation.EvidenceJSON)) {
		return fmt.Errorf("evidence must be valid JSON")
	}
	if recommendation.Explanation == "" {
		return fmt.Errorf("missing explanation")
	}
	return nil
}
