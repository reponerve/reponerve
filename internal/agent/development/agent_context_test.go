package development_test

import (
	"testing"

	"github.com/reponerve/reponerve/internal/agent/development"
)

func TestBuildAgentContextMeta_ConceptExplanationFull(t *testing.T) {
	answer := &development.DevelopmentAnswer{
		AnswerType: "concept_explanation",
		EntityBriefings: []development.EntityBriefing{
			{QualifiedName: "pkg.Foo", DefinedIn: "pkg/foo.go:1-10"},
		},
	}

	meta := development.BuildAgentContextMeta(answer)
	if meta.Completeness != development.CompletenessFull {
		t.Fatalf("completeness: got %q want full", meta.Completeness)
	}
	if !meta.MustUseBeforeEdit {
		t.Fatal("expected must_use_before_edit for concept with briefings")
	}
	if len(meta.RecommendedNextTools) == 0 {
		t.Fatal("expected recommended_next_tools")
	}
}

func TestBuildAgentContextMeta_SearchSummaryRetrievalOnly(t *testing.T) {
	answer := &development.DevelopmentAnswer{
		AnswerType: "search_summary",
		Summary:    "170 hits",
	}

	meta := development.BuildAgentContextMeta(answer)
	if meta.Completeness != development.CompletenessRetrievalOnly {
		t.Fatalf("completeness: got %q want retrieval_only", meta.Completeness)
	}
	if meta.MustUseBeforeEdit {
		t.Fatal("search_summary must not set must_use_before_edit")
	}
}

func TestBuildAgentContextMeta_PlanRequiresEditGate(t *testing.T) {
	plan := &development.DevelopmentPlan{
		Task: "Add feature",
		StartingPoints: []development.EntityRef{
			{EntityType: "FILE", Label: "internal/foo/bar.go"},
		},
	}

	meta := development.BuildAgentContextMeta(plan)
	if !meta.MustUseBeforeEdit {
		t.Fatal("plan must set must_use_before_edit")
	}
	if meta.Completeness != development.CompletenessFull {
		t.Fatalf("completeness: got %q want full", meta.Completeness)
	}
}

func TestNewMCPResult_IncludesAgentMeta(t *testing.T) {
	answer := &development.DevelopmentAnswer{
		AnswerType: "concept_explanation",
		EntityBriefings: []development.EntityBriefing{
			{QualifiedName: "pkg.Foo"},
		},
	}

	result := development.NewMCPResult("formatted", answer)
	if result.Agent.Kind != "concept_explanation" {
		t.Fatalf("agent.kind: got %q", result.Agent.Kind)
	}
	if result.Agent.Completeness != development.CompletenessFull {
		t.Fatalf("agent.completeness: got %q", result.Agent.Completeness)
	}
}
