package development

import (
	"context"
	"fmt"
	"regexp"
	"strings"
)

var taskTicketPrefix = regexp.MustCompile(`(?i)^[A-Z][A-Z0-9]+-\d+`)
var taskTicketStrip = regexp.MustCompile(`(?i)^[A-Z][A-Z0-9]+-\d+:\s*`)

// NormalizeTaskTopic removes ticket prefixes so repository search can resolve the topic.
func NormalizeTaskTopic(input string) string {
	return strings.TrimSpace(taskTicketStrip.ReplaceAllString(strings.TrimSpace(input), ""))
}

var taskIntakePrefixes = []string{
	"add ", "implement ", "fix ", "refactor ", "update ", "create ",
	"build ", "integrate ", "remove ", "migrate ", "support ", "introduce ",
}

// LooksLikeTaskDescription reports whether input reads as an implementation assignment.
func LooksLikeTaskDescription(input string) bool {
	input = strings.TrimSpace(input)
	if input == "" {
		return false
	}
	lower := strings.ToLower(input)
	for _, prefix := range taskIntakePrefixes {
		if strings.HasPrefix(lower, prefix) {
			return true
		}
	}
	if taskTicketPrefix.MatchString(input) {
		return true
	}
	if strings.Count(input, "\n") >= 2 {
		return true
	}
	if strings.Contains(lower, "as a user") || strings.Contains(lower, "acceptance criteria") {
		return true
	}
	// Pasted ticket / multi-line assignment without a leading verb.
	return len(strings.Fields(input)) >= 12
}

func (s *Service) answerTaskPlan(ctx context.Context, repositoryID, question string) (*DevelopmentAnswer, error) {
	plan, err := s.Plan(ctx, DevelopmentRequest{
		RepositoryID: repositoryID,
		Topic:        question,
	})
	if err != nil {
		return nil, fmt.Errorf("task plan failed: %w", err)
	}
	out := taskAnswerFromPlan(question, plan)
	sortEntityRefs(out.Related)
	sortEvidence(out.Evidence)
	return out, nil
}

func taskAnswerFromPlan(question string, plan *DevelopmentPlan) *DevelopmentAnswer {
	out := &DevelopmentAnswer{
		Question:        question,
		AnswerType:      answerTypeTaskPlan,
		Plan:            plan,
		EntityBriefings: plan.EntityBriefings,
		Evidence:        plan.Evidence,
		SourceServices:  plan.SourceServices,
	}

	var summaryParts []string
	summaryParts = append(summaryParts, fmt.Sprintf("Task plan for: %s", plan.Task))
	if len(plan.SuggestedSteps) > 0 {
		summaryParts = append(summaryParts, "Suggested steps:")
		for _, step := range plan.SuggestedSteps {
			summaryParts = append(summaryParts, "  "+step)
		}
	}
	if briefSummary := summarizeBriefings(plan.EntityBriefings); briefSummary != "" {
		summaryParts = append(summaryParts, briefSummary)
	}
	if len(plan.RelevantDecisions) > 0 {
		labels := make([]string, 0, len(plan.RelevantDecisions))
		for _, d := range plan.RelevantDecisions {
			labels = append(labels, d.Label)
		}
		summaryParts = append(summaryParts, fmt.Sprintf("Relevant decisions:\n  - %s", strings.Join(labels, "\n  - ")))
	}
	if len(summaryParts) == 1 && len(plan.StartingPoints) > 0 {
		labels := make([]string, 0, len(plan.StartingPoints))
		for _, sp := range plan.StartingPoints {
			labels = append(labels, sp.Label)
		}
		summaryParts = append(summaryParts, fmt.Sprintf("Starting points: %s", strings.Join(labels, ", ")))
	}
	out.Summary = strings.Join(summaryParts, "\n\n")

	seen := make(map[string]struct{})
	for _, ref := range plan.StartingPoints {
		key := ref.EntityType + ":" + ref.EntityID
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out.Related = append(out.Related, ref)
	}
	for _, ref := range plan.ImpactedAreas {
		key := ref.EntityType + ":" + ref.EntityID
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out.Related = append(out.Related, ref)
	}
	return out
}

func buildSuggestedSteps(plan *DevelopmentPlan) []string {
	if plan == nil {
		return nil
	}
	var steps []string
	n := 1

	if len(plan.RelevantDecisions) > 0 {
		steps = append(steps, fmt.Sprintf("%d. Review relevant decisions — they constrain implementation.", n))
		n++
	}
	if len(plan.EntityBriefings) > 0 {
		steps = append(steps, fmt.Sprintf("%d. Read entity_briefings for roles, defined_in, and relationships.", n))
		n++
	}
	if len(plan.StartingPoints) > 0 {
		labels := make([]string, 0, min(4, len(plan.StartingPoints)))
		for i, sp := range plan.StartingPoints {
			if i >= 4 {
				break
			}
			label := strings.TrimSpace(sp.Label)
			if label == "" {
				label = sp.EntityID
			}
			labels = append(labels, label)
		}
		suffix := ""
		if len(plan.StartingPoints) > len(labels) {
			suffix = fmt.Sprintf(" (+%d more)", len(plan.StartingPoints)-len(labels))
		}
		steps = append(steps, fmt.Sprintf("%d. Begin at starting points: %s%s.", n, strings.Join(labels, ", "), suffix))
		n++
	}
	steps = append(steps, fmt.Sprintf("%d. Run analyze_topic_impact before broad refactors.", n))
	n++
	steps = append(steps, fmt.Sprintf("%d. Run review on the task topic before merge.", n))
	return steps
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
