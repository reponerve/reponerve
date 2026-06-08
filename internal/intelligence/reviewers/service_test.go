package reviewers

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/reponerve/reponerve/internal/graph/impact"
	"github.com/reponerve/reponerve/internal/graph/relationships"
	"github.com/reponerve/reponerve/internal/graph/traversal"
	"github.com/reponerve/reponerve/internal/intelligence/discovery"
	memorymodels "github.com/reponerve/reponerve/internal/memory/models"
	"github.com/reponerve/reponerve/internal/query/storage"
	"github.com/reponerve/reponerve/internal/storage/migrations"
	"github.com/reponerve/reponerve/internal/storage/sqlite"
	models "github.com/reponerve/reponerve/pkg/models"
)

// --- Mock Readers ---

type mockDecisionReader struct {
	decisions []*memorymodels.Decision
}

func (m *mockDecisionReader) GetByID(ctx context.Context, id string) (*memorymodels.Decision, error) {
	for _, d := range m.decisions {
		if d.ID == id {
			return d, nil
		}
	}
	return nil, fmt.Errorf("not found")
}
func (m *mockDecisionReader) ListByRepository(ctx context.Context, repoID string) ([]*memorymodels.Decision, error) {
	return m.decisions, nil
}
func (m *mockDecisionReader) ListAll(ctx context.Context) ([]*memorymodels.Decision, error) { return nil, nil }

type mockIntentReader struct {
	intents []*memorymodels.Intent
}

func (m *mockIntentReader) GetByID(ctx context.Context, id string) (*memorymodels.Intent, error) { return nil, nil }
func (m *mockIntentReader) ListByRepository(ctx context.Context, repoID string) ([]*memorymodels.Intent, error) {
	return m.intents, nil
}
func (m *mockIntentReader) ListAll(ctx context.Context) ([]*memorymodels.Intent, error) { return nil, nil }

type mockFactReader struct {
	facts []*memorymodels.Fact
}

func (m *mockFactReader) GetByID(ctx context.Context, id string) (*memorymodels.Fact, error) {
	for _, f := range m.facts {
		if f.ID == id {
			return f, nil
		}
	}
	return nil, fmt.Errorf("not found")
}
func (m *mockFactReader) ListByRepository(ctx context.Context, repoID string) ([]*memorymodels.Fact, error) {
	return m.facts, nil
}
func (m *mockFactReader) ListAll(ctx context.Context) ([]*memorymodels.Fact, error) { return nil, nil }

type mockEventReader struct {
	events []*models.Event
}

func (m *mockEventReader) GetByID(ctx context.Context, id string) (*models.Event, error) {
	for _, ev := range m.events {
		if ev.ID == id {
			return ev, nil
		}
	}
	return nil, fmt.Errorf("not found")
}
func (m *mockEventReader) ListByRepository(ctx context.Context, repoID string) ([]*models.Event, error) {
	return m.events, nil
}
func (m *mockEventReader) ListAll(ctx context.Context) ([]*models.Event, error) { return nil, nil }

type mockRelationshipReader struct {
	rels []*memorymodels.Relationship
}

func (m *mockRelationshipReader) GetByID(ctx context.Context, id string) (*memorymodels.Relationship, error) {
	return nil, nil
}
func (m *mockRelationshipReader) ListByRepository(ctx context.Context, repoID string) ([]*memorymodels.Relationship, error) {
	return m.rels, nil
}
func (m *mockRelationshipReader) ListAll(ctx context.Context) ([]*memorymodels.Relationship, error) { return nil, nil }

type mockContributorReader struct {
	contribs []*models.Contributor
}

func (m *mockContributorReader) GetByID(ctx context.Context, repoID string, id string) (*models.Contributor, error) {
	return nil, nil
}
func (m *mockContributorReader) ListByRepository(ctx context.Context, repoID string) ([]*models.Contributor, error) {
	return m.contribs, nil
}

type mockExpertiseReader struct {
	expertise []*models.Expertise
}

func (m *mockExpertiseReader) ListByRepository(ctx context.Context, repoID string) ([]*models.Expertise, error) {
	return m.expertise, nil
}
func (m *mockExpertiseReader) ListByContributor(ctx context.Context, repoID string, cID string) ([]*models.Expertise, error) {
	return nil, nil
}

type mockSourceReader struct {
	sources []*models.Source
}

func (m *mockSourceReader) ListByRepository(ctx context.Context, repoID string) ([]*models.Source, error) {
	return m.sources, nil
}

// Helper to construct a Service under test
func buildTestService(
	dr storage.DecisionReader,
	fr storage.FactReader,
	er storage.EventReader,
	cr storage.ContributorReader,
	expr storage.ExpertiseReader,
	rr storage.RelationshipReader,
	sr storage.SourceReader,
) *Service {
	relEngine := relationships.NewEngine(dr, &mockIntentReader{}, fr, er, rr, cr, expr, sr)
	travEngine := traversal.NewEngine(relEngine)
	impactSvc := impact.NewService(travEngine)

	discoverySvc := discovery.NewService(dr, fr, er, cr, expr, rr, relEngine, travEngine, impactSvc)
	return NewService(discoverySvc, dr, fr, er, cr, expr, sr, impactSvc)
}

// --- Unit Tests ---

func TestService_EmptyRepository(t *testing.T) {
	ctx := context.Background()
	svc := buildTestService(
		&mockDecisionReader{},
		&mockFactReader{},
		&mockEventReader{},
		&mockContributorReader{},
		&mockExpertiseReader{},
		&mockRelationshipReader{},
		&mockSourceReader{},
	)

	report, err := svc.RecommendRepositoryReviewers(ctx, "repo_1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(report.Recommendations) != 0 {
		t.Errorf("expected 0 recommendations, got %d", len(report.Recommendations))
	}
}

func TestService_RepositoryReviewers(t *testing.T) {
	ctx := context.Background()
	repoID := "repo_1"

	cID1 := contributorID(repoID, "Dev A", "deva@example.com")
	cID2 := contributorID(repoID, "Dev B", "devb@example.com")

	contribs := []*models.Contributor{
		{ID: cID1, RepositoryID: repoID, Name: "Dev A", Email: "deva@example.com"},
		{ID: cID2, RepositoryID: repoID, Name: "Dev B", Email: "devb@example.com"},
	}

	exps := []*models.Expertise{
		{ID: "exp_1", RepositoryID: repoID, ContributorID: cID1, Domain: "Storage", Score: 1.0},
		{ID: "exp_2", RepositoryID: repoID, ContributorID: cID1, Domain: "Authentication", Score: 1.0},
		{ID: "exp_3", RepositoryID: repoID, ContributorID: cID2, Domain: "Storage", Score: 0.5},
	}

	svc := buildTestService(
		&mockDecisionReader{},
		&mockFactReader{},
		&mockEventReader{},
		&mockContributorReader{contribs: contribs},
		&mockExpertiseReader{expertise: exps},
		&mockRelationshipReader{},
		&mockSourceReader{},
	)

	report, err := svc.RecommendRepositoryReviewers(ctx, repoID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Dev A has 2 expertise domains, so Score = 2 + 2 + 2 = 6 (assuming discovery score is expCount+domainCount = 2)
	// Let's verify details
	if len(report.Recommendations) != 2 {
		t.Fatalf("expected 2 recommendations, got %d", len(report.Recommendations))
	}

	// Dev A should be first due to higher score
	recA := report.Recommendations[0]
	if recA.ContributorID != cID1 {
		t.Errorf("expected first contributor to be %s, got %s", cID1, recA.ContributorID)
	}

	var ev map[string]interface{}
	_ = json.Unmarshal([]byte(recA.EvidenceJSON), &ev)
	if ev["expertise_count"].(float64) != 2 || ev["domain_count"].(float64) != 2 {
		t.Errorf("unexpected evidence for Dev A: %s", recA.EvidenceJSON)
	}

	expectedExplanation := "Contributor is recommended because they participate in 2 repository domains and maintain 2 expertise areas."
	if recA.Explanation != expectedExplanation {
		t.Errorf("expected explanation %q, got %q", expectedExplanation, recA.Explanation)
	}
}

func TestService_DomainReviewers(t *testing.T) {
	ctx := context.Background()
	repoID := "repo_1"

	cID1 := contributorID(repoID, "Dev A", "deva@example.com")
	cID2 := contributorID(repoID, "Dev B", "devb@example.com")

	contribs := []*models.Contributor{
		{ID: cID1, RepositoryID: repoID, Name: "Dev A", Email: "deva@example.com"},
		{ID: cID2, RepositoryID: repoID, Name: "Dev B", Email: "devb@example.com"},
	}

	exps := []*models.Expertise{
		{ID: "exp_1", RepositoryID: repoID, ContributorID: cID1, Domain: "Storage", Score: 1.0},
		{ID: "exp_2", RepositoryID: repoID, ContributorID: cID2, Domain: "Authentication", Score: 1.0},
	}

	svc := buildTestService(
		&mockDecisionReader{},
		&mockFactReader{},
		&mockEventReader{},
		&mockContributorReader{contribs: contribs},
		&mockExpertiseReader{expertise: exps},
		&mockRelationshipReader{},
		&mockSourceReader{},
	)

	report, err := svc.RecommendDomainReviewers(ctx, repoID, "Storage")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(report.Recommendations) != 1 {
		t.Fatalf("expected 1 recommendation, got %d", len(report.Recommendations))
	}

	rec := report.Recommendations[0]
	if rec.ContributorID != cID1 {
		t.Errorf("expected recommended contributor to be %s, got %s", cID1, rec.ContributorID)
	}

	expectedExplanation := "Contributor is recommended because their expertise matches the selected repository domain."
	if rec.Explanation != expectedExplanation {
		t.Errorf("expected explanation %q, got %q", expectedExplanation, rec.Explanation)
	}
}

func TestService_ImpactReviewers(t *testing.T) {
	ctx := context.Background()
	repoID := "repo_1"

	cID1 := contributorID(repoID, "Dev A", "deva@example.com")

	contribs := []*models.Contributor{
		{ID: cID1, RepositoryID: repoID, Name: "Dev A", Email: "deva@example.com"},
	}

	sources := []*models.Source{
		{ID: "src_1", RepositoryID: repoID, Author: "Dev A <deva@example.com>"},
	}

	decisions := []*memorymodels.Decision{
		{ID: "dec_1", RepositoryID: repoID, SourceID: "src_1", Title: "Use SQL database Selection", CreatedAt: time.Now()},
	}

	exps := []*models.Expertise{
		{ID: "exp_1", RepositoryID: repoID, ContributorID: cID1, Domain: "Storage", Score: 1.0},
	}

	svc := buildTestService(
		&mockDecisionReader{decisions: decisions},
		&mockFactReader{},
		&mockEventReader{},
		&mockContributorReader{contribs: contribs},
		&mockExpertiseReader{expertise: exps},
		&mockRelationshipReader{},
		&mockSourceReader{sources: sources},
	)

	report, err := svc.RecommendImpactReviewers(ctx, repoID, "dec_1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(report.Recommendations) != 1 {
		t.Fatalf("expected 1 recommendation, got %d", len(report.Recommendations))
	}

	rec := report.Recommendations[0]
	if rec.ContributorID != cID1 {
		t.Errorf("expected recommended contributor to be %s, got %s", cID1, rec.ContributorID)
	}

	var ev map[string]interface{}
	_ = json.Unmarshal([]byte(rec.EvidenceJSON), &ev)
	if ev["impact_entities"].(float64) != 1 || ev["matching_expertise"].(float64) != 1 {
		t.Errorf("unexpected evidence values: %s", rec.EvidenceJSON)
	}

	expectedExplanation := "Contributor is recommended because their expertise overlaps with impacted repository knowledge."
	if rec.Explanation != expectedExplanation {
		t.Errorf("expected explanation %q, got %q", expectedExplanation, rec.Explanation)
	}
}

func TestValidateRecommendation_Unit(t *testing.T) {
	rec := &ReviewerRecommendation{
		ContributorID: "c_1",
		Score:         2.5,
		EvidenceJSON:  `{"expertise_count":2}`,
		Explanation:   "explanation",
	}

	if err := ValidateRecommendation(rec); err != nil {
		t.Errorf("expected valid recommendation, got: %v", err)
	}

	if err := ValidateRecommendation(nil); err == nil {
		t.Error("expected error for nil recommendation")
	}

	rec.Score = -0.1
	if err := ValidateRecommendation(rec); err == nil {
		t.Error("expected error for score < 0")
	}
	rec.Score = 0

	rec.ContributorID = ""
	if err := ValidateRecommendation(rec); err == nil {
		t.Error("expected error for empty contributor ID")
	}
}

// --- Integration Tests ---

func TestService_Integration(t *testing.T) {
	ctx := context.Background()
	tempDir, err := os.MkdirTemp("", "reponerve-reviewers-integration-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	dbPath := filepath.Join(tempDir, "test.db")
	db, err := sqlite.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	if err := migrations.RunUp(db); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	repoID := "repo_int"

	// Seed repository
	_, err = db.Exec("INSERT INTO repositories (id, name, path, default_branch, created_at, updated_at) VALUES (?, ?, ?, ?, datetime(), datetime())", repoID, "Repo Test", tempDir, "main")
	if err != nil {
		t.Fatalf("failed to insert repository: %v", err)
	}

	// Seed source
	_, err = db.Exec("INSERT INTO sources (id, repository_id, source_type, reference, title, author, timestamp, created_at) VALUES (?, ?, ?, ?, ?, ?, datetime(), datetime())", "src_1", repoID, "adr", "docs/adr/0001.md", "Title 1", "Dev User <dev@example.com>", "2024-01-01T00:00:00Z")
	if err != nil {
		t.Fatalf("failed to insert source: %v", err)
	}

	// Seed decisions
	_, err = db.Exec("INSERT INTO memory_decisions (id, repository_id, source_id, title, status, created_at) VALUES (?, ?, ?, ?, ?, datetime())", "dec_a", repoID, "src_1", "Use SQLite selection db", "Accepted")
	if err != nil {
		t.Fatalf("failed to insert decision: %v", err)
	}

	// Seed contributor & expertise
	cID := contributorID(repoID, "Dev User", "dev@example.com")
	_, err = db.Exec("INSERT INTO contributors (id, repository_id, email, name, first_seen, last_seen, commit_count) VALUES (?, ?, ?, ?, datetime(), datetime(), ?)", cID, repoID, "dev@example.com", "Dev User", 5)
	if err != nil {
		t.Fatalf("failed to insert contributor: %v", err)
	}
	_, err = db.Exec("INSERT INTO expertise (id, repository_id, contributor_id, domain, score, evidence_json) VALUES (?, ?, ?, ?, ?, ?)", "exp_a", repoID, cID, "Storage", 0.9, `{"commits":5}`)
	if err != nil {
		t.Fatalf("failed to insert expertise: %v", err)
	}

	// Instantiate actual storage readers
	dr := storage.NewSQLiteDecisionReader(db)
	ir := storage.NewSQLiteIntentReader(db)
	fr := storage.NewSQLiteFactReader(db)
	er := storage.NewSQLiteEventReader(db)
	rr := storage.NewSQLiteRelationshipReader(db)
	cr := storage.NewSQLiteContributorReader(db)
	expr := storage.NewSQLiteExpertiseReader(db)
	sr := storage.NewSQLiteSourceReader(db)

	relEngine := relationships.NewEngine(dr, ir, fr, er, rr, cr, expr, sr)
	travEngine := traversal.NewEngine(relEngine)
	impactSvc := impact.NewService(travEngine)

	discoverySvc := discovery.NewService(dr, fr, er, cr, expr, rr, relEngine, travEngine, impactSvc)
	svc := NewService(discoverySvc, dr, fr, er, cr, expr, sr, impactSvc)

	report, err := svc.RecommendRepositoryReviewers(ctx, repoID)
	if err != nil {
		t.Fatalf("RecommendRepositoryReviewers failed: %v", err)
	}

	if len(report.Recommendations) != 1 {
		t.Fatalf("expected 1 recommendation, got %d", len(report.Recommendations))
	}

	rec := report.Recommendations[0]
	if rec.ContributorID != cID {
		t.Errorf("unexpected recommended contributor ID: %s", rec.ContributorID)
	}
}
