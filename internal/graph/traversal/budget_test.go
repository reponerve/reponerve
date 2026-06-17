package traversal

import (
	"context"
	"testing"

	"github.com/reponerve/reponerve/internal/graph/model"
	"github.com/reponerve/reponerve/internal/graph/relationships"
	memorymodels "github.com/reponerve/reponerve/internal/memory/models"
)

func TestTraverseWithBudget(t *testing.T) {
	ctx := context.Background()
	intents := []*memorymodels.Intent{{ID: "intent_1", RepositoryID: "repo_1"}}
	decisions := []*memorymodels.Decision{{ID: "dec_1", RepositoryID: "repo_1"}, {ID: "dec_2", RepositoryID: "repo_1"}}
	rels := []*memorymodels.Relationship{
		{ID: "r1", RepositoryID: "repo_1", FromID: "intent_1", ToID: "dec_1", Type: "INTENT_DRIVES_DECISION"},
		{ID: "r2", RepositoryID: "repo_1", FromID: "dec_1", ToID: "dec_2", Type: "DECISION_DEPENDS_ON_DECISION"},
	}
	relEngine := relationships.NewEngine(
		&mockDecisionReader{decisions: decisions},
		&mockIntentReader{intents: intents},
		&mockFactReader{},
		&mockEventReader{},
		&mockRelationshipReader{rels: rels},
		&mockContributorReader{},
		&mockExpertiseReader{},
		&mockSourceReader{},
	)
	engine := NewEngine(relEngine)
	start := model.NodeID("repo_1", model.NodeTypeIntent, "intent_1")

	result, err := engine.TraverseWithBudget(ctx, "repo_1", start, BudgetTraversalOptions{
		TraversalOptions: TraversalOptions{IncludeStored: true},
		TokenBudget:      50,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Nodes) < 2 {
		t.Fatalf("expected budget traversal to include neighbors, got %d nodes", len(result.Nodes))
	}
	if result.TokensUsed <= 0 || result.TokensUsed > 50 {
		t.Fatalf("unexpected tokens used: %d", result.TokensUsed)
	}
}
