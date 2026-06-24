package development

import (
	"testing"
)

func TestPruneStructured_CapsRelated(t *testing.T) {
	refs := make([]EntityRef, 0, 25)
	for i := 0; i < 25; i++ {
		refs = append(refs, EntityRef{EntityType: "FILE", EntityID: "id", Label: "f"})
	}
	answer := &DevelopmentAnswer{Related: refs}
	out, report := PruneStructured(answer)
	pruned := out.(*DevelopmentAnswer)
	if len(pruned.Related) != MaxRelatedRefs {
		t.Fatalf("related len=%d want %d", len(pruned.Related), MaxRelatedRefs)
	}
	if !report.Truncated || report.OmittedCounts["related"] != 10 {
		t.Fatalf("report=%+v", report)
	}
}

func TestEffectiveTokenBudget_Default(t *testing.T) {
	if got := (OutputOptions{}).EffectiveTokenBudget(); got != DefaultTokenBudget {
		t.Fatalf("got %d want %d", got, DefaultTokenBudget)
	}
}

func TestApplyPruneReport_SetsMeta(t *testing.T) {
	meta := AgentContextMeta{Completeness: CompletenessFull}
	ApplyPruneReport(&meta, PruneReport{
		Truncated:       true,
		TruncatedFields: []string{"related"},
		OmittedCounts:   map[string]int{"related": 5},
	})
	if !meta.Truncated || meta.Completeness != CompletenessPartial {
		t.Fatalf("meta=%+v", meta)
	}
	if len(meta.PreferNarrowTools) == 0 {
		t.Fatal("expected prefer_narrow_tools")
	}
}
