package fact

import (
	"context"
	"testing"
	"time"

	models "github.com/reponerve/reponerve/pkg/models"
)

func adrSource(id, repoID, title, content string) *models.Source {
	var metadata string
	if content != "" {
		metadata = `{"content": "` + content + `"}`
	}
	return &models.Source{
		ID:           id,
		RepositoryID: repoID,
		SourceType:   "adr",
		Reference:    "docs/adr/" + id + ".md",
		Title:        title,
		MetadataJSON: metadata,
		Timestamp:    time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	}
}

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
		input         string
		wantSubject   string
		wantPredicate string
		wantObject    string
	}{
		{"Auth Service uses Redis", "Auth Service", "USES", "Redis"},
		{"API Gateway depends on Auth Service", "API Gateway", "DEPENDS_ON", "Auth Service"},
		{"Billing stores data in PostgreSQL", "Billing", "STORES_IN", "PostgreSQL"},
		{"User Service calls Notification Service", "User Service", "CALLS", "Notification Service"},
	}

	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			sources := []*models.Source{
				adrSource("adr1", "repo-1", "ADR Title", tc.input),
			}

			facts, err := extractor.Extract(context.Background(), sources)
			if err != nil {
				t.Fatalf("Extract failed: %v", err)
			}

			if len(facts) != 1 {
				t.Fatalf("expected 1 fact, got %d", len(facts))
			}

			f := facts[0]
			if f.Subject != tc.wantSubject {
				t.Errorf("expected Subject %q, got %q", tc.wantSubject, f.Subject)
			}
			if f.Predicate != tc.wantPredicate {
				t.Errorf("expected Predicate %q, got %q", tc.wantPredicate, f.Predicate)
			}
			if f.Object != tc.wantObject {
				t.Errorf("expected Object %q, got %q", tc.wantObject, f.Object)
			}
		})
	}
}

func TestExtract_NegativeCases(t *testing.T) {
	extractor := NewExtractor()
	cases := []string{
		"Use Redis",
		"Improve reliability",
		"Reduce latency",
		"uses Redis",
		"depends on auth",
		"calls notifications",
		"stores data in postgres",
	}

	for _, tc := range cases {
		t.Run(tc, func(t *testing.T) {
			sources := []*models.Source{
				adrSource("adr1", "repo-1", "ADR Title", tc),
			}

			facts, err := extractor.Extract(context.Background(), sources)
			if err != nil {
				t.Fatalf("Extract failed: %v", err)
			}

			if len(facts) != 0 {
				t.Errorf("expected 0 facts for %q, got %d", tc, len(facts))
			}
		})
	}
}

func TestExtract_CaseInsensitivity(t *testing.T) {
	extractor := NewExtractor()
	cases := []struct {
		input         string
		wantSubject   string
		wantPredicate string
		wantObject    string
	}{
		{"Auth Service USES Redis", "Auth Service", "USES", "Redis"},
		{"API Gateway DEPENDS ON Auth Service", "API Gateway", "DEPENDS_ON", "Auth Service"},
		{"Billing STORES DATA IN PostgreSQL", "Billing", "STORES_IN", "PostgreSQL"},
		{"User Service CALLS Notification Service", "User Service", "CALLS", "Notification Service"},
	}

	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			sources := []*models.Source{
				adrSource("adr1", "repo-1", "ADR Title", tc.input),
			}

			facts, err := extractor.Extract(context.Background(), sources)
			if err != nil {
				t.Fatalf("Extract failed: %v", err)
			}

			if len(facts) != 1 {
				t.Fatalf("expected 1 fact, got %d", len(facts))
			}

			f := facts[0]
			if f.Predicate != tc.wantPredicate {
				t.Errorf("expected Predicate %q, got %q", tc.wantPredicate, f.Predicate)
			}
		})
	}
}

func TestExtract_CleanMarkdown(t *testing.T) {
	extractor := NewExtractor()
	sources := []*models.Source{
		adrSource("adr1", "repo-1", "ADR Title", "**Auth Service** uses *Redis*"),
	}

	facts, err := extractor.Extract(context.Background(), sources)
	if err != nil {
		t.Fatalf("Extract failed: %v", err)
	}

	if len(facts) != 1 {
		t.Fatalf("expected 1 fact, got %d", len(facts))
	}

	f := facts[0]
	if f.Subject != "Auth Service" {
		t.Errorf("expected Subject 'Auth Service', got %q", f.Subject)
	}
	if f.Object != "Redis" {
		t.Errorf("expected Object 'Redis', got %q", f.Object)
	}
}

func TestExtract_MultipleFacts(t *testing.T) {
	extractor := NewExtractor()
	sources := []*models.Source{
		adrSource("adr1", "repo-1", "ADR Title", "Auth Service uses Redis and API Gateway depends on Auth Service."),
	}

	facts, err := extractor.Extract(context.Background(), sources)
	if err != nil {
		t.Fatalf("Extract failed: %v", err)
	}

	if len(facts) != 2 {
		t.Fatalf("expected 2 facts, got %d", len(facts))
	}

	if facts[0].Subject != "Auth Service" || facts[0].Predicate != "USES" || facts[0].Object != "Redis" {
		t.Errorf("unexpected first fact: %+v", facts[0])
	}
	if facts[1].Subject != "API Gateway" || facts[1].Predicate != "DEPENDS_ON" || facts[1].Object != "Auth Service" {
		t.Errorf("unexpected second fact: %+v", facts[1])
	}
}

func TestExtract_IgnoreNonADR(t *testing.T) {
	extractor := NewExtractor()
	sources := []*models.Source{
		commitSource("hash1", "repo-1", "Auth Service uses Redis"),
	}

	facts, err := extractor.Extract(context.Background(), sources)
	if err != nil {
		t.Fatalf("Extract failed: %v", err)
	}

	if len(facts) != 0 {
		t.Errorf("expected 0 facts from commit, got %d", len(facts))
	}
}

func TestExtract_DeterministicIDs(t *testing.T) {
	extractor := NewExtractor()
	src1 := adrSource("adr1", "repo-1", "ADR Title", "Auth Service uses Redis")
	src2 := adrSource("adr1", "repo-1", "ADR Title", "Auth Service uses Redis")

	facts1, _ := extractor.Extract(context.Background(), []*models.Source{src1})
	facts2, _ := extractor.Extract(context.Background(), []*models.Source{src2})

	if len(facts1) != 1 || len(facts2) != 1 {
		t.Fatalf("expected 1 fact extracted from each, got %d and %d", len(facts1), len(facts2))
	}

	if facts1[0].ID != facts2[0].ID {
		t.Errorf("expected deterministic IDs, got %q and %q", facts1[0].ID, facts2[0].ID)
	}

	if len(facts1[0].ID) < 5 || facts1[0].ID[:5] != "fact_" {
		t.Errorf("expected ID prefix 'fact_', got %q", facts1[0].ID)
	}
}
