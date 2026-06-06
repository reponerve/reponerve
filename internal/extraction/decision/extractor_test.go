package decision

import (
	"context"
	"testing"
	"time"

	models "reponerve/pkg/models"
)

func adrSource(id, repoID, title, status string, validJSON bool) *models.Source {
	metaJSON := ""
	if status != "" {
		if validJSON {
			metaJSON = `{"status": "` + status + `"}`
		} else {
			metaJSON = `{"status": "` + status + `"` // missing closing brace
		}
	} else if !validJSON {
		metaJSON = `{"invalid": json`
	}

	return &models.Source{
		ID:           id,
		RepositoryID: repoID,
		SourceType:   "adr",
		Reference:    "docs/adr/" + id + ".md",
		Title:        title,
		Timestamp:    time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		MetadataJSON: metaJSON,
	}
}

func TestExtract_AcceptedADR(t *testing.T) {
	extractor := NewExtractor()
	sources := []*models.Source{
		adrSource("adr1", "repo-1", "Use Redis Cache", "Accepted", true),
	}

	decisions, err := extractor.Extract(context.Background(), sources)
	if err != nil {
		t.Fatalf("Extract failed: %v", err)
	}

	if len(decisions) != 1 {
		t.Fatalf("expected 1 decision, got %d", len(decisions))
	}

	dec := decisions[0]
	if dec.Title != "Use Redis Cache" {
		t.Errorf("expected title 'Use Redis Cache', got %q", dec.Title)
	}
	if dec.Status != "Accepted" {
		t.Errorf("expected status 'Accepted', got %q", dec.Status)
	}
	if dec.SourceID != "adr1" {
		t.Errorf("expected source ID 'adr1', got %q", dec.SourceID)
	}
}

func TestExtract_ProposedADR(t *testing.T) {
	extractor := NewExtractor()
	sources := []*models.Source{
		adrSource("adr2", "repo-1", "Adopt gRPC", "Proposed", true),
	}

	decisions, err := extractor.Extract(context.Background(), sources)
	if err != nil {
		t.Fatalf("Extract failed: %v", err)
	}

	if len(decisions) != 1 {
		t.Fatalf("expected 1 decision, got %d", len(decisions))
	}

	dec := decisions[0]
	if dec.Title != "Adopt gRPC" {
		t.Errorf("expected title 'Adopt gRPC', got %q", dec.Title)
	}
	if dec.Status != "Proposed" {
		t.Errorf("expected status 'Proposed', got %q", dec.Status)
	}
}

func TestExtract_RejectedADR(t *testing.T) {
	extractor := NewExtractor()
	sources := []*models.Source{
		adrSource("adr3", "repo-1", "Use XML", "Rejected", true),
	}

	decisions, err := extractor.Extract(context.Background(), sources)
	if err != nil {
		t.Fatalf("Extract failed: %v", err)
	}

	if len(decisions) != 1 {
		t.Fatalf("expected 1 decision, got %d", len(decisions))
	}

	dec := decisions[0]
	if dec.Status != "Rejected" {
		t.Errorf("expected status 'Rejected', got %q", dec.Status)
	}
}

func TestExtract_MissingStatus(t *testing.T) {
	extractor := NewExtractor()
	// Create an ADR source with no status field in MetadataJSON
	src := adrSource("adr4", "repo-1", "Build CLI", "", true)
	src.MetadataJSON = `{"path": "docs/adr/adr4.md"}`

	decisions, err := extractor.Extract(context.Background(), []*models.Source{src})
	if err != nil {
		t.Fatalf("Extract failed: %v", err)
	}

	if len(decisions) != 1 {
		t.Fatalf("expected 1 decision, got %d", len(decisions))
	}

	dec := decisions[0]
	if dec.Status != "Proposed" {
		t.Errorf("expected fallback status 'Proposed', got %q", dec.Status)
	}
}

func TestExtract_MalformedMetadata(t *testing.T) {
	extractor := NewExtractor()
	sources := []*models.Source{
		adrSource("adr5", "repo-1", "Clean Architecture", "Accepted", false),
	}

	decisions, err := extractor.Extract(context.Background(), sources)
	if err != nil {
		t.Fatalf("Extract failed: %v", err)
	}

	if len(decisions) != 1 {
		t.Fatalf("expected 1 decision, got %d", len(decisions))
	}

	dec := decisions[0]
	if dec.Status != "Proposed" {
		t.Errorf("expected fallback status 'Proposed' for malformed metadata, got %q", dec.Status)
	}
}

func TestExtract_NonADRSource(t *testing.T) {
	extractor := NewExtractor()
	sources := []*models.Source{
		{
			ID:           "commit1",
			RepositoryID: "repo-1",
			SourceType:   "commit",
			Reference:    "commit1",
			Title:        "feat: add feature",
			Timestamp:    time.Now(),
		},
	}

	decisions, err := extractor.Extract(context.Background(), sources)
	if err != nil {
		t.Fatalf("Extract failed: %v", err)
	}

	if len(decisions) != 0 {
		t.Errorf("expected 0 decisions extracted from commit source, got %d", len(decisions))
	}
}

func TestExtract_DeterministicIDs(t *testing.T) {
	extractor := NewExtractor()
	src1 := adrSource("adr_id", "repo-1", "Title", "Accepted", true)
	src2 := adrSource("adr_id", "repo-1", "Title", "Accepted", true)

	decisions1, _ := extractor.Extract(context.Background(), []*models.Source{src1})
	decisions2, _ := extractor.Extract(context.Background(), []*models.Source{src2})

	if decisions1[0].ID != decisions2[0].ID {
		t.Errorf("expected deterministic IDs, got %q and %q", decisions1[0].ID, decisions2[0].ID)
	}

	// Make sure prefix is decision_
	if len(decisions1[0].ID) < 9 || decisions1[0].ID[:9] != "decision_" {
		t.Errorf("expected ID prefix 'decision_', got %q", decisions1[0].ID)
	}
}
