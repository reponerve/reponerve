package searchindex

import (
	"testing"
	"time"

	memorymodels "github.com/reponerve/reponerve/internal/memory/models"
	models "github.com/reponerve/reponerve/pkg/models"
)

func TestBuildDocuments(t *testing.T) {
	now := time.Now()
	docs := BuildDocuments(Input{
		RepositoryID: "repo_1",
		Events: []*models.Event{
			{ID: "evt_1", RepositoryID: "repo_1", Title: "Introduce MCP", Description: "tools", EventType: "FEATURE_INTRODUCED"},
			{ID: "evt_2", RepositoryID: "repo_2", Title: "Other repo"},
		},
		Decisions: []*memorymodels.Decision{
			{ID: "dec_1", RepositoryID: "repo_1", Title: "Use SQLite", Status: "Accepted", CreatedAt: now},
		},
		Facts: []*memorymodels.Fact{
			{ID: "fact_1", RepositoryID: "repo_1", Subject: "cache", Predicate: "uses", Object: "redis", CreatedAt: now},
		},
	})

	if len(docs) != 3 {
		t.Fatalf("expected 3 docs, got %d", len(docs))
	}
	if docs[0].EntityType != "EVENT" || docs[0].MemoryID != "evt_1" {
		t.Fatalf("unexpected first doc: %+v", docs[0])
	}
}
