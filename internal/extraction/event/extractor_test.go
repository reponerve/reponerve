package event

import (
	"context"
	"testing"
	"time"

	"reponerve/pkg/models"
)

func source(id, repoID, title string) *models.Source {
	return &models.Source{
		ID:           id,
		RepositoryID: repoID,
		SourceType:   "commit",
		Reference:    id,
		Title:        title,
		Timestamp:    time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	}
}

// ─── classifyCommit ────────────────────────────────────────────────────────

func TestClassifyCommit_AllPrefixes(t *testing.T) {
	cases := []struct {
		input     string
		wantType  string
		wantMatch bool
	}{
		{"feat: add feature", EventTypeFeatureIntroduced, true},
		{"feature: add feature", EventTypeFeatureIntroduced, true},
		{"fix: resolve bug", EventTypeDefectResolved, true},
		{"bugfix: resolve bug", EventTypeDefectResolved, true},
		{"refactor: clean up", EventTypeCodeRefactored, true},
		{"docs: update readme", EventTypeDocumentationUpdated, true},
		{"chore: bump deps", EventTypeMaintenancePerformed, true},
		{"FEAT: uppercase prefix", EventTypeFeatureIntroduced, true},
		{"wip: in progress", "", false},
		{"initial commit", "", false},
		{"", "", false},
	}

	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			gotType, gotMatch := classifyCommit(tc.input)
			if gotMatch != tc.wantMatch {
				t.Errorf("classifyCommit(%q): match = %v, want %v", tc.input, gotMatch, tc.wantMatch)
			}
			if gotType != tc.wantType {
				t.Errorf("classifyCommit(%q): type = %q, want %q", tc.input, gotType, tc.wantType)
			}
		})
	}
}

// ─── deriveTitle ───────────────────────────────────────────────────────────

func TestDeriveTitle(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{"feat(cache): introduce redis cache", "Introduce Redis Cache"},
		{"fix: resolve null pointer", "Resolve Null Pointer"},
		{"refactor: simplify auth module", "Simplify Auth Module"},
		{"docs: update api reference", "Update Api Reference"},
		{"chore: bump dependencies", "Bump Dependencies"},
		{"feat: ", ""},
		{"no colon here", "No Colon Here"},
	}

	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			got := deriveTitle(tc.input)
			if got != tc.want {
				t.Errorf("deriveTitle(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

// ─── eventID ───────────────────────────────────────────────────────────────

func TestEventID_Deterministic(t *testing.T) {
	id1 := eventID("repo-1", "abc123")
	id2 := eventID("repo-1", "abc123")
	if id1 != id2 {
		t.Errorf("eventID is not deterministic: %q != %q", id1, id2)
	}
}

func TestEventID_DifferentInputs(t *testing.T) {
	a := eventID("repo-1", "abc123")
	b := eventID("repo-1", "def456")
	c := eventID("repo-2", "abc123")
	if a == b || a == c || b == c {
		t.Errorf("eventID produced collision: a=%q b=%q c=%q", a, b, c)
	}
}

func TestEventID_HasPrefix(t *testing.T) {
	id := eventID("repo-1", "abc123")
	if len(id) < 4 || id[:4] != "evt_" {
		t.Errorf("expected eventID to start with 'evt_', got %q", id)
	}
}

// ─── Extract ───────────────────────────────────────────────────────────────

func TestExtract_QualifyingCommits(t *testing.T) {
	extractor := NewExtractor()
	sources := []*models.Source{
		source("hash1", "repo-1", "feat(cache): introduce redis"),
		source("hash2", "repo-1", "fix: resolve auth timeout"),
		source("hash3", "repo-1", "chore: update dependencies"),
	}

	events, err := extractor.Extract(context.Background(), sources)
	if err != nil {
		t.Fatalf("Extract failed: %v", err)
	}
	if len(events) != 3 {
		t.Fatalf("expected 3 events, got %d", len(events))
	}

	if events[0].EventType != EventTypeFeatureIntroduced {
		t.Errorf("expected %s, got %s", EventTypeFeatureIntroduced, events[0].EventType)
	}
	if events[1].EventType != EventTypeDefectResolved {
		t.Errorf("expected %s, got %s", EventTypeDefectResolved, events[1].EventType)
	}
	if events[2].EventType != EventTypeMaintenancePerformed {
		t.Errorf("expected %s, got %s", EventTypeMaintenancePerformed, events[2].EventType)
	}
}

func TestExtract_NonMatchingCommitsSkipped(t *testing.T) {
	extractor := NewExtractor()
	sources := []*models.Source{
		source("hash1", "repo-1", "initial commit"),
		source("hash2", "repo-1", "wip: not done yet"),
		source("hash3", "repo-1", "feat: valid commit"),
	}

	events, err := extractor.Extract(context.Background(), sources)
	if err != nil {
		t.Fatalf("Extract failed: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event (only feat: qualifies), got %d", len(events))
	}
}

func TestExtract_NonCommitSourcesSkipped(t *testing.T) {
	extractor := NewExtractor()
	sources := []*models.Source{
		{
			ID:           "adr-1",
			RepositoryID: "repo-1",
			SourceType:   "adr",
			Title:        "feat: this should be skipped",
			Timestamp:    time.Now(),
		},
		source("hash1", "repo-1", "feat: valid commit"),
	}

	events, err := extractor.Extract(context.Background(), sources)
	if err != nil {
		t.Fatalf("Extract failed: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event (adr skipped), got %d", len(events))
	}
}

func TestExtract_SourceTraceability(t *testing.T) {
	extractor := NewExtractor()
	src := source("abc123", "repo-1", "feat: add authentication")

	events, err := extractor.Extract(context.Background(), []*models.Source{src})
	if err != nil {
		t.Fatalf("Extract failed: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}

	evt := events[0]
	if evt.SourceID != "abc123" {
		t.Errorf("expected SourceID %q, got %q", "abc123", evt.SourceID)
	}
	if evt.RepositoryID != "repo-1" {
		t.Errorf("expected RepositoryID %q, got %q", "repo-1", evt.RepositoryID)
	}
	if evt.Timestamp != src.Timestamp {
		t.Errorf("expected Timestamp %v, got %v", src.Timestamp, evt.Timestamp)
	}
}

func TestExtract_EmptySources(t *testing.T) {
	extractor := NewExtractor()
	events, err := extractor.Extract(context.Background(), []*models.Source{})
	if err != nil {
		t.Fatalf("Extract failed on empty input: %v", err)
	}
	if len(events) != 0 {
		t.Errorf("expected 0 events for empty input, got %d", len(events))
	}
}

func TestExtract_TitleDerivation(t *testing.T) {
	extractor := NewExtractor()
	src := source("hash1", "repo-1", "feat(cache): introduce redis cache")

	events, _ := extractor.Extract(context.Background(), []*models.Source{src})
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].Title != "Introduce Redis Cache" {
		t.Errorf("expected title %q, got %q", "Introduce Redis Cache", events[0].Title)
	}
}

func TestExtract_IDIsStable(t *testing.T) {
	extractor := NewExtractor()
	src := source("hash1", "repo-1", "feat: stable id test")

	events1, _ := extractor.Extract(context.Background(), []*models.Source{src})
	events2, _ := extractor.Extract(context.Background(), []*models.Source{src})

	if events1[0].ID != events2[0].ID {
		t.Errorf("event ID is not stable: %q vs %q", events1[0].ID, events2[0].ID)
	}
}
