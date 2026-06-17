package compression

import (
	"testing"

	memorymodels "github.com/reponerve/reponerve/internal/memory/models"
)

func TestScoreTextTopicMatch(t *testing.T) {
	tokens := topicTokens("sqlite storage")
	if scoreText(tokens, "Local-first SQLite storage") < 10 {
		t.Fatal("expected sqlite match")
	}
	if scoreText(tokens, "Unrelated topic") != 0 {
		t.Fatal("expected no match")
	}
}

func TestPackByTokenBudget(t *testing.T) {
	decisions := []*memorymodels.Decision{
		{ID: "d1", Title: "SQLite local-first"},
		{ID: "d2", Title: "Redis caching"},
	}
	scores := map[string]int{"d1": 20, "d2": 5}
	out, _, _, _ := packByTokenBudget(decisions, nil, nil, nil, scores, 5)
	if len(out) != 1 || out[0].ID != "d1" {
		t.Fatalf("expected highest-scored decision within budget, got %+v", out)
	}
}
