package intent

import (
	"context"
	"testing"
	"time"

	models "github.com/reponerve/reponerve/pkg/models"
)

func commitSource(id, repoID, title string) *models.Source {
	return &models.Source{
		ID:           id,
		RepositoryID: repoID,
		SourceType:   "commit",
		Reference:    id,
		Title:        title,
		Timestamp:    time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	}
}

func TestExtract_PositiveCases(t *testing.T) {
	extractor := NewExtractor()
	cases := []struct {
		input    string
		wantDesc string
	}{
		{"feat(cache): reduce latency", "Reduce Latency"},
		{"fix: improve reliability", "Improve Reliability"},
		{"chore: optimize deployment speed", "Optimize Deployment Speed"},
		{"docs: simplify configuration", "Simplify Configuration"},
		{"refactor: secure API access", "Secure API Access"},
	}

	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			sources := []*models.Source{
				commitSource("hash1", "repo-1", tc.input),
			}

			intents, err := extractor.Extract(context.Background(), sources)
			if err != nil {
				t.Fatalf("Extract failed: %v", err)
			}

			if len(intents) != 1 {
				t.Fatalf("expected 1 intent, got %d", len(intents))
			}

			if intents[0].Description != tc.wantDesc {
				t.Errorf("expected description %q, got %q", tc.wantDesc, intents[0].Description)
			}
		})
	}
}

func TestExtract_NegativeCases(t *testing.T) {
	extractor := NewExtractor()
	cases := []string{
		"feat: use Redis",
		"fix: create API",
		"chore: add endpoint",
		"initial commit",
	}

	for _, tc := range cases {
		t.Run(tc, func(t *testing.T) {
			sources := []*models.Source{
				commitSource("hash1", "repo-1", tc),
			}

			intents, err := extractor.Extract(context.Background(), sources)
			if err != nil {
				t.Fatalf("Extract failed: %v", err)
			}

			if len(intents) != 0 {
				t.Errorf("expected 0 intents, got %d", len(intents))
			}
		})
	}
}

func TestExtract_MultipleIntents(t *testing.T) {
	extractor := NewExtractor()
	sources := []*models.Source{
		commitSource("hash1", "repo-1", "feat: reduce latency and improve reliability"),
	}

	intents, err := extractor.Extract(context.Background(), sources)
	if err != nil {
		t.Fatalf("Extract failed: %v", err)
	}

	if len(intents) != 2 {
		t.Fatalf("expected 2 intents, got %d", len(intents))
	}

	if intents[0].Description != "Reduce Latency" {
		t.Errorf("expected first intent 'Reduce Latency', got %q", intents[0].Description)
	}
	if intents[1].Description != "Improve Reliability" {
		t.Errorf("expected second intent 'Improve Reliability', got %q", intents[1].Description)
	}
}

func TestExtract_DeterministicIDs(t *testing.T) {
	extractor := NewExtractor()
	src1 := commitSource("hash1", "repo-1", "feat: reduce latency")
	src2 := commitSource("hash1", "repo-1", "feat: reduce latency")

	intents1, _ := extractor.Extract(context.Background(), []*models.Source{src1})
	intents2, _ := extractor.Extract(context.Background(), []*models.Source{src2})

	if len(intents1) != 1 || len(intents2) != 1 {
		t.Fatalf("expected 1 intent extracted, got %d and %d", len(intents1), len(intents2))
	}

	if intents1[0].ID != intents2[0].ID {
		t.Errorf("expected deterministic IDs, got %q and %q", intents1[0].ID, intents2[0].ID)
	}

	if len(intents1[0].ID) < 7 || intents1[0].ID[:7] != "intent_" {
		t.Errorf("expected ID prefix 'intent_', got %q", intents1[0].ID)
	}
}
