package development

// Agent context completeness levels for MCP responses.
const (
	CompletenessFull          = "full"
	CompletenessPartial       = "partial"
	CompletenessRetrievalOnly = "retrieval_only"
)

// Shared guidance so weak models and strong models follow the same evidence contract.
var guidanceEvidenceOnly = []string{
	"Cite only facts present in structured — do not invent paths, types, or ADRs.",
	"If a fact is missing, say so and use recommended_next_tools — do not guess.",
}

var guidanceTokenDiscipline = []string{
	"Prefer structured fields over bulk file reads when completeness is full.",
}

// AgentContextMeta tells AI consumers how to use a Development Experience payload.
type AgentContextMeta struct {
	Kind                 string   `json:"kind"`
	Completeness         string   `json:"completeness"`
	MustUseBeforeEdit    bool     `json:"must_use_before_edit"`
	GuidanceForAgent     []string `json:"guidance_for_agent"`
	RecommendedNextTools []string `json:"recommended_next_tools,omitempty"`
}

// BuildAgentContextMeta derives agent instructions from a structured DE payload.
func BuildAgentContextMeta(structured any) AgentContextMeta {
	switch v := structured.(type) {
	case *DevelopmentAnswer:
		return metaFromAnswer(v)
	case DevelopmentAnswer:
		return metaFromAnswer(&v)
	case *DevelopmentExplanation:
		return metaFromExplanation(v)
	case DevelopmentExplanation:
		return metaFromExplanation(&v)
	case *DevelopmentPlan:
		return metaFromPlan(v)
	case DevelopmentPlan:
		return metaFromPlan(&v)
	case *DevelopmentImpactReport:
		return metaFromImpact(v)
	case DevelopmentImpactReport:
		return metaFromImpact(&v)
	case *DevelopmentReviewGuide:
		return metaFromReview(v)
	case DevelopmentReviewGuide:
		return metaFromReview(&v)
	case *DevelopmentOnboardingGuide:
		return metaFromOnboarding(v)
	case DevelopmentOnboardingGuide:
		return metaFromOnboarding(&v)
	default:
		return AgentContextMeta{
			Kind:             "unknown",
			Completeness:     CompletenessPartial,
			GuidanceForAgent: []string{"Read structured payload before answering or editing."},
		}
	}
}

func metaFromAnswer(a *DevelopmentAnswer) AgentContextMeta {
	if a == nil {
		return AgentContextMeta{Kind: "answer", Completeness: CompletenessPartial}
	}

	meta := AgentContextMeta{
		Kind:             a.AnswerType,
		Completeness:     completenessForAnswer(a),
		GuidanceForAgent: append([]string{
			"Read structured.entity_briefings and related before synthesizing.",
			"Do not edit code from search hit counts alone.",
		}, guidanceEvidenceOnly...),
	}
	meta.GuidanceForAgent = append(meta.GuidanceForAgent, guidanceTokenDiscipline...)

	switch a.AnswerType {
	case answerTypeConceptExplanation:
		meta.MustUseBeforeEdit = len(a.EntityBriefings) > 0
		meta.GuidanceForAgent = append(meta.GuidanceForAgent,
			"Anchor explanations on entity_briefings (role, defined_in, fields, producers, consumers).",
		)
		if len(a.EntityBriefings) > 1 {
			meta.GuidanceForAgent = append(meta.GuidanceForAgent,
				"Disambiguate homonyms explicitly; edit only the briefing that matches the task.",
			)
		}
		if meta.Completeness == CompletenessFull {
			meta.RecommendedNextTools = []string{"plan", "analyze_topic_impact"}
		}
	case answerTypeSearchSummary:
		meta.Completeness = CompletenessRetrievalOnly
		meta.GuidanceForAgent = append([]string{
			"This is retrieval-only context, not software understanding.",
			"Do not answer confidently or edit code from this response.",
			"Re-query with ask (What is X?) or explain before editing code.",
		}, guidanceEvidenceOnly...)
		meta.RecommendedNextTools = []string{"ask", "explain", "plan"}
	case answerTypeDecisionRationale:
		meta.MustUseBeforeEdit = true
		meta.GuidanceForAgent = append(meta.GuidanceForAgent,
			"Treat related decisions as constraints on implementation.",
		)
		meta.RecommendedNextTools = []string{"trace_decision", "plan"}
	case answerTypeDependency:
		meta.RecommendedNextTools = []string{"analyze_topic_impact", "explain"}
	case answerTypeTaskPlan:
		meta.MustUseBeforeEdit = true
		meta.Completeness = completenessForPlan(a.Plan)
		meta.GuidanceForAgent = append([]string{
			"Task intake: follow suggested_steps and plan.starting_points.",
			"Implement only within scoped impacted_areas.",
		}, guidanceEvidenceOnly...)
		meta.GuidanceForAgent = append(meta.GuidanceForAgent, guidanceTokenDiscipline...)
		meta.RecommendedNextTools = []string{"analyze_topic_impact", "explain_file", "review"}
	}

	return meta
}

func completenessForAnswer(a *DevelopmentAnswer) string {
	if a.AnswerType == answerTypeTaskPlan {
		return completenessForPlan(a.Plan)
	}
	if len(a.EntityBriefings) > 0 {
		return CompletenessFull
	}
	if a.AnswerType == answerTypeSearchSummary {
		return CompletenessRetrievalOnly
	}
	if len(a.Related) > 0 || len(a.Evidence) > 0 {
		return CompletenessPartial
	}
	return CompletenessPartial
}

func metaFromExplanation(e *DevelopmentExplanation) AgentContextMeta {
	if e == nil {
		return AgentContextMeta{Kind: "unified_explanation", Completeness: CompletenessPartial}
	}

	meta := AgentContextMeta{
		Kind: "unified_explanation",
		GuidanceForAgent: append([]string{
			"Combine entity_briefings with code_context and repository_context.",
			"Honor repository_code_links when connecting decisions to code.",
		}, guidanceEvidenceOnly...),
	}
	meta.GuidanceForAgent = append(meta.GuidanceForAgent, guidanceTokenDiscipline...)

	if len(e.EntityBriefings) > 0 {
		meta.Completeness = CompletenessFull
		meta.MustUseBeforeEdit = true
		if len(e.EntityBriefings) > 1 {
			meta.GuidanceForAgent = append(meta.GuidanceForAgent,
				"Compare briefings before choosing a symbol to edit.",
			)
		}
		meta.RecommendedNextTools = []string{"plan", "analyze_topic_impact"}
	} else if hasExplanationContext(e) {
		meta.Completeness = CompletenessPartial
		meta.RecommendedNextTools = []string{"explain_struct", "explain_function", "ask"}
	} else {
		meta.Completeness = CompletenessRetrievalOnly
		meta.RecommendedNextTools = []string{"ask", "explain"}
	}

	return meta
}

func hasExplanationContext(e *DevelopmentExplanation) bool {
	if e.CodeContext != nil && hasCodeContent(e.CodeContext) {
		return true
	}
	if e.RepositoryContext != nil && hasRepositoryContent(e.RepositoryContext) {
		return true
	}
	return len(e.RepositoryCodeLinks) > 0
}

func metaFromPlan(p *DevelopmentPlan) AgentContextMeta {
	if p == nil {
		return AgentContextMeta{Kind: "plan", Completeness: CompletenessPartial}
	}

	meta := AgentContextMeta{
		Kind:              "plan",
		Completeness:      CompletenessFull,
		MustUseBeforeEdit: true,
		GuidanceForAgent: append([]string{
			"Task intake: treat this plan as the scope boundary for pasted assignments.",
			"Implement only within impacted_areas and starting_points.",
			"Apply relevant_decisions as architectural constraints.",
			"Run explain/ask on unknown terms before editing if briefings are absent.",
		}, guidanceEvidenceOnly...),
		RecommendedNextTools: []string{"analyze_topic_impact", "explain_file", "review"},
	}
	meta.GuidanceForAgent = append(meta.GuidanceForAgent, guidanceTokenDiscipline...)
	meta.Completeness = completenessForPlan(p)
	return meta
}

func completenessForPlan(p *DevelopmentPlan) string {
	if p == nil {
		return CompletenessPartial
	}
	if len(p.EntityBriefings) > 0 || len(p.StartingPoints) > 0 {
		return CompletenessFull
	}
	if len(p.ImpactedAreas) > 0 || len(p.RelevantDecisions) > 0 {
		return CompletenessPartial
	}
	return CompletenessPartial
}

func metaFromImpact(r *DevelopmentImpactReport) AgentContextMeta {
	if r == nil {
		return AgentContextMeta{Kind: "impact", Completeness: CompletenessPartial}
	}

	meta := AgentContextMeta{
		Kind:              "impact",
		Completeness:      CompletenessFull,
		MustUseBeforeEdit: true,
		GuidanceForAgent: append([]string{
			"Update dependent_areas and code_dependencies when refactoring.",
			"Re-run impact after substantive changes.",
		}, guidanceEvidenceOnly...),
		RecommendedNextTools: []string{"plan", "review"},
	}
	if len(r.DependentAreas) == 0 && len(r.CodeDependencies) == 0 {
		meta.Completeness = CompletenessPartial
	}
	return meta
}

func metaFromOnboarding(g *DevelopmentOnboardingGuide) AgentContextMeta {
	if g == nil {
		return AgentContextMeta{Kind: "onboarding", Completeness: CompletenessPartial}
	}

	meta := AgentContextMeta{
		Kind: "onboarding",
		GuidanceForAgent: append([]string{
			"First-day context: read key_decisions and orientation before exploring files.",
			"If assignment_plan is present, follow its suggested_steps for the task.",
		}, guidanceEvidenceOnly...),
		RecommendedNextTools: []string{"explain", "plan", "list_decisions"},
	}
	if g.AssignmentPlan != nil {
		meta.MustUseBeforeEdit = true
		meta.Completeness = completenessForPlan(g.AssignmentPlan)
		meta.RecommendedNextTools = []string{"analyze_topic_impact", "explain_file", "review"}
	} else if len(g.EntityBriefings) > 0 {
		meta.Completeness = CompletenessFull
	} else if len(g.KeyDecisions) > 0 || g.Orientation != nil {
		meta.Completeness = CompletenessPartial
	}
	meta.GuidanceForAgent = append(meta.GuidanceForAgent, guidanceTokenDiscipline...)
	return meta
}

func metaFromReview(g *DevelopmentReviewGuide) AgentContextMeta {
	if g == nil {
		return AgentContextMeta{Kind: "review", Completeness: CompletenessPartial}
	}

	return AgentContextMeta{
		Kind:         "review",
		Completeness: CompletenessFull,
		MustUseBeforeEdit: false,
		GuidanceForAgent: []string{
			"Verify affected_areas and related_knowledge before merge.",
			"Involve recommended_reviewers when expertise is required.",
		},
	}
}
