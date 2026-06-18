package development

import (
	"strings"
	"testing"
)

func TestToCompactShortensHeaders(t *testing.T) {
	in := "ENTITY BRIEFINGS\n  foo [bar]\n\nREPOSITORY CONTEXT\n  decision"
	out := ToCompact(in)
	if strings.Contains(out, "ENTITY BRIEFINGS") {
		t.Fatalf("expected header shortened: %q", out)
	}
	if !strings.Contains(out, "BRIEF") || !strings.Contains(out, "REPO") {
		t.Fatalf("expected compact headers: %q", out)
	}
}

func TestToCompactReducesSize(t *testing.T) {
	in := strings.Repeat("ENTITY BRIEFINGS\n  internal/foo/bar.go [FILE]\n  Related decisions: ADR-1\n\n", 20)
	proseLen := len(in)
	compactLen := len(ToCompact(in))
	if compactLen >= proseLen {
		t.Fatalf("compact should shrink prose: prose=%d compact=%d", proseLen, compactLen)
	}
	if float64(compactLen)/float64(proseLen) > 0.50 {
		t.Fatalf("expected >=50%% reduction, got %d/%d (%.0f%%)", compactLen, proseLen, 100*float64(compactLen)/float64(proseLen))
	}
}

func TestToCompactRealisticDEOutput(t *testing.T) {
	related := make([]EntityRef, 0, 30)
	for i := 0; i < 30; i++ {
		related = append(related, EntityRef{
			EntityType: "FUNCTION",
			Label:      "internal/storage/sqlite/helper_" + itoa(i),
		})
	}
	evidence := make([]EvidenceItem, 0, 20)
	for i := 0; i < 20; i++ {
		evidence = append(evidence, EvidenceItem{Source: "code_intelligence", Type: "code_entity"})
	}
	answer := &DevelopmentAnswer{
		Question:       "Why do we use SQLite?",
		AnswerType:     "decision_rationale",
		Summary:        "Relevant decisions for \"SQLite\":\n  - 1. Local-first SQLite storage — Software understanding must remain available offline and under contributor control.",
		Related:        related,
		Evidence:       evidence,
		SourceServices: []string{"repository_search", "architectural_guidance", "code_intelligence"},
	}
	prose := FormatAnswer(answer)
	compact := ToCompact(prose)
	ratio := float64(len(compact)) / float64(len(prose))
	if ratio > 0.50 {
		t.Fatalf("expected >=50%% reduction on realistic answer, got %d/%d (%.0f%%)", len(compact), len(prose), ratio*100)
	}
	if !strings.Contains(compact, "EV:") {
		t.Fatalf("expected collapsed evidence line, got %q", compact)
	}
}

func TestTruncateToTokenBudget(t *testing.T) {
	text := strings.Repeat("word ", 100)
	out := TruncateToTokenBudget(text, 10)
	if len(out) > 10*4+3 {
		t.Fatalf("truncate exceeded budget: len=%d", len(out))
	}
}
