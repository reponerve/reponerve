package learning

import (
	"context"
	"encoding/json"
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
func (m *mockDecisionReader) GetByID(ctx context.Context, id string) (*memorymodels.Decision, error) { return nil, nil }
func (m *mockDecisionReader) ListByRepository(ctx context.Context, repoID string) ([]*memorymodels.Decision, error) { return m.decisions, nil }
func (m *mockDecisionReader) ListAll(ctx context.Context) ([]*memorymodels.Decision, error) { return nil, nil }

type mockIntentReader struct {
	intents []*memorymodels.Intent
}
func (m *mockIntentReader) GetByID(ctx context.Context, id string) (*memorymodels.Intent, error) { return nil, nil }
func (m *mockIntentReader) ListByRepository(ctx context.Context, repoID string) ([]*memorymodels.Intent, error) { return m.intents, nil }
func (m *mockIntentReader) ListAll(ctx context.Context) ([]*memorymodels.Intent, error) { return nil, nil }

type mockFactReader struct {
	facts []*memorymodels.Fact
}
func (m *mockFactReader) GetByID(ctx context.Context, id string) (*memorymodels.Fact, error) { return nil, nil }
func (m *mockFactReader) ListByRepository(ctx context.Context, repoID string) ([]*memorymodels.Fact, error) { return m.facts, nil }
func (m *mockFactReader) ListAll(ctx context.Context) ([]*memorymodels.Fact, error) { return nil, nil }

type mockEventReader struct {
	events []*models.Event
}
func (m *mockEventReader) GetByID(ctx context.Context, id string) (*models.Event, error) { return nil, nil }
func (m *mockEventReader) ListByRepository(ctx context.Context, repoID string) ([]*models.Event, error) { return m.events, nil }
func (m *mockEventReader) ListAll(ctx context.Context) ([]*models.Event, error) { return nil, nil }

type mockRelationshipReader struct {
	rels []*memorymodels.Relationship
}
func (m *mockRelationshipReader) GetByID(ctx context.Context, id string) (*memorymodels.Relationship, error) { return nil, nil }
func (m *mockRelationshipReader) ListByRepository(ctx context.Context, repoID string) ([]*memorymodels.Relationship, error) { return m.rels, nil }
func (m *mockRelationshipReader) ListAll(ctx context.Context) ([]*memorymodels.Relationship, error) { return nil, nil }

type mockContributorReader struct {
	contribs []*models.Contributor
}
func (m *mockContributorReader) GetByID(ctx context.Context, repoID string, id string) (*models.Contributor, error) { return nil, nil }
func (m *mockContributorReader) ListByRepository(ctx context.Context, repoID string) ([]*models.Contributor, error) { return m.contribs, nil }

type mockExpertiseReader struct {
	expertise []*models.Expertise
}
func (m *mockExpertiseReader) ListByRepository(ctx context.Context, repoID string) ([]*models.Expertise, error) { return m.expertise, nil }
func (m *mockExpertiseReader) ListByContributor(ctx context.Context, repoID string, cID string) ([]*models.Expertise, error) { return nil, nil }

type mockSourceReader struct {
	sources []*models.Source
}
func (m *mockSourceReader) ListByRepository(ctx context.Context, repoID string) ([]*models.Source, error) { return m.sources, nil }

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
	return NewService(discoverySvc, dr, fr, er, cr, expr, sr, relEngine)
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

	path, err := svc.GenerateRepositoryPath(ctx, "repo_1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(path.Steps) != 0 {
		t.Errorf("expected 0 steps, got %d", len(path.Steps))
	}
}

func TestService_RepositoryPathOrdering(t *testing.T) {
	ctx := context.Background()
	repoID := "repo_1"

	decisions := []*memorymodels.Decision{
		{ID: "dec_1", RepositoryID: repoID, Title: "Use SQLite", Status: "Accepted", CreatedAt: time.Now()},
	}
	facts := []*memorymodels.Fact{
		{ID: "fact_1", RepositoryID: repoID, Subject: "db", Predicate: "is", Object: "WAL", CreatedAt: time.Now()},
	}
	events := []*models.Event{
		{ID: "evt_1", RepositoryID: repoID, Title: "Commit", Timestamp: time.Now()},
	}
	contribs := []*models.Contributor{
		{ID: "contrib_1", RepositoryID: repoID, Name: "Dev", Email: "dev@example.com"},
	}

	svc := buildTestService(
		&mockDecisionReader{decisions: decisions},
		&mockFactReader{facts: facts},
		&mockEventReader{events: events},
		&mockContributorReader{contribs: contribs},
		&mockExpertiseReader{},
		&mockRelationshipReader{},
		&mockSourceReader{},
	)

	path, err := svc.GenerateRepositoryPath(ctx, repoID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Expected order: Decisions -> Facts -> Events -> Contributors
	if len(path.Steps) != 4 {
		t.Fatalf("expected 4 steps, got %d", len(path.Steps))
	}

	if path.Steps[0].EntityType != EntityTypeDecision || path.Steps[0].EntityID != "dec_1" || path.Steps[0].Position != 1 {
		t.Errorf("step 0 invalid: %+v", path.Steps[0])
	}
	if path.Steps[1].EntityType != EntityTypeFact || path.Steps[1].EntityID != "fact_1" || path.Steps[1].Position != 2 {
		t.Errorf("step 1 invalid: %+v", path.Steps[1])
	}
	if path.Steps[2].EntityType != EntityTypeEvent || path.Steps[2].EntityID != "evt_1" || path.Steps[2].Position != 3 {
		t.Errorf("step 2 invalid: %+v", path.Steps[2])
	}
	if path.Steps[3].EntityType != EntityTypeContributor || path.Steps[3].EntityID != "contrib_1" || path.Steps[3].Position != 4 {
		t.Errorf("step 3 invalid: %+v", path.Steps[3])
	}

	// Verify evidence preservation
	var ev map[string]interface{}
	_ = json.Unmarshal([]byte(path.Steps[0].EvidenceJSON), &ev)
	if ev["discovery_score"] == nil || ev["position_reason"] != "repository_foundation" {
		t.Errorf("evidence missing score or correct reason: %s", path.Steps[0].EvidenceJSON)
	}

	// Verify explanations
	if path.Steps[0].Explanation != "This decision appears early because it is highly ranked repository knowledge and provides foundational context." {
		t.Errorf("unexpected explanation: %q", path.Steps[0].Explanation)
	}
}

func TestService_DomainPath(t *testing.T) {
	ctx := context.Background()
	repoID := "repo_1"

	// Match Storage domain (keywords: storage, db, database, sqlite, postgres, sql, persistence, query)
	decisions := []*memorymodels.Decision{
		{ID: "dec_storage", RepositoryID: repoID, Title: "Storage DB selection", CreatedAt: time.Now()},
		{ID: "dec_auth", RepositoryID: repoID, Title: "Authentication layer", CreatedAt: time.Now()},
	}

	contribs := []*models.Contributor{
		{ID: "contrib_storage", RepositoryID: repoID, Name: "DB Expert"},
	}

	exps := []*models.Expertise{
		{ID: "exp_1", RepositoryID: repoID, ContributorID: "contrib_storage", Domain: "Storage", Score: 1.0},
	}

	svc := buildTestService(
		&mockDecisionReader{decisions: decisions},
		&mockFactReader{},
		&mockEventReader{},
		&mockContributorReader{contribs: contribs},
		&mockExpertiseReader{expertise: exps},
		&mockRelationshipReader{},
		&mockSourceReader{},
	)

	path, err := svc.GenerateDomainPath(ctx, repoID, "Storage")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Expected items matching storage: contrib_storage and dec_storage.
	// Order: Contributors -> Decisions
	if len(path.Steps) != 2 {
		t.Fatalf("expected 2 steps, got %d", len(path.Steps))
	}

	if path.Steps[0].EntityType != EntityTypeContributor || path.Steps[0].EntityID != "contrib_storage" {
		t.Errorf("expected contributor first, got: %+v", path.Steps[0])
	}
	if path.Steps[1].EntityType != EntityTypeDecision || path.Steps[1].EntityID != "dec_storage" {
		t.Errorf("expected decision second, got: %+v", path.Steps[1])
	}

	// Explanation verification
	if path.Steps[0].Explanation != "This contributor appears early because it is relevant to the selected repository domain." {
		t.Errorf("unexpected explanation: %q", path.Steps[0].Explanation)
	}
}

func TestService_ContributorPath(t *testing.T) {
	ctx := context.Background()
	repoID := "repo_1"

	cID := contributorID(repoID, "Dev A", "deva@example.com")

	contribs := []*models.Contributor{
		{ID: cID, RepositoryID: repoID, Name: "Dev A", Email: "deva@example.com"},
	}

	sources := []*models.Source{
		{ID: "src_1", RepositoryID: repoID, Author: "Dev A <deva@example.com>"},
	}

	// dec_1 is authored by Dev A, dec_2 is not.
	decisions := []*memorymodels.Decision{
		{ID: "dec_1", RepositoryID: repoID, SourceID: "src_1", Title: "PG layer", CreatedAt: time.Now()},
		{ID: "dec_2", RepositoryID: repoID, SourceID: "src_unknown", Title: "auth layer", CreatedAt: time.Now()},
	}

	svc := buildTestService(
		&mockDecisionReader{decisions: decisions},
		&mockFactReader{},
		&mockEventReader{},
		&mockContributorReader{contribs: contribs},
		&mockExpertiseReader{},
		&mockRelationshipReader{},
		&mockSourceReader{sources: sources},
	)

	path, err := svc.GenerateContributorPath(ctx, repoID, cID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Expected: contributor profile, and dec_1
	if len(path.Steps) != 2 {
		t.Fatalf("expected 2 steps, got %d", len(path.Steps))
	}

	if path.Steps[0].EntityType != EntityTypeContributor || path.Steps[0].EntityID != cID {
		t.Errorf("step 0 invalid: %+v", path.Steps[0])
	}
	if path.Steps[1].EntityType != EntityTypeDecision || path.Steps[1].EntityID != "dec_1" {
		t.Errorf("step 1 invalid: %+v", path.Steps[1])
	}

	if path.Steps[0].Explanation != "This expertise area appears early because it is associated with the selected contributor." {
		t.Errorf("unexpected explanation: %q", path.Steps[0].Explanation)
	}
	if path.Steps[1].Explanation != "This decision appears early because it is associated with the selected contributor." {
		t.Errorf("unexpected explanation: %q", path.Steps[1].Explanation)
	}
}

func TestValidateStep_Unit(t *testing.T) {
	step := &LearningStep{
		EntityType:   "DECISION",
		EntityID:     "dec_1",
		Position:     1,
		EvidenceJSON: `{"discovery_score":2}`,
		Explanation:  "explanation",
	}

	if err := ValidateStep(step); err != nil {
		t.Errorf("expected valid step, got: %v", err)
	}

	// Nil step
	if err := ValidateStep(nil); err == nil {
		t.Error("expected error for nil step")
	}

	// Invalid position
	step.Position = 0
	if err := ValidateStep(step); err == nil {
		t.Error("expected error for position <= 0")
	}
	step.Position = 1

	// Invalid type
	step.EntityType = "UNKNOWN"
	if err := ValidateStep(step); err == nil {
		t.Error("expected error for invalid EntityType")
	}
}

// --- Integration Tests ---

func TestService_Integration(t *testing.T) {
	ctx := context.Background()
	tempDir, err := os.MkdirTemp("", "reponerve-learning-integration-*")
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
	_, err = db.Exec("INSERT INTO sources (id, repository_id, source_type, reference, title, author, timestamp, created_at) VALUES (?, ?, ?, ?, ?, ?, datetime(), datetime())", "src_1", repoID, "adr", "docs/adr/0001.md", "Title 1", "Author", "2024-01-01T00:00:00Z")
	if err != nil {
		t.Fatalf("failed to insert source: %v", err)
	}

	// Seed decisions
	_, err = db.Exec("INSERT INTO memory_decisions (id, repository_id, source_id, title, status, created_at) VALUES (?, ?, ?, ?, ?, datetime())", "dec_a", repoID, "src_1", "Use SQLite", "Accepted")
	if err != nil {
		t.Fatalf("failed to insert decision: %v", err)
	}

	// Seed facts
	_, err = db.Exec("INSERT INTO memory_facts (id, repository_id, subject, predicate, object, source_id, created_at) VALUES (?, ?, ?, ?, ?, ?, datetime())", "fact_a", repoID, "sqlite", "is", "active", "src_1")
	if err != nil {
		t.Fatalf("failed to insert fact: %v", err)
	}

	// Seed relationship
	_, err = db.Exec("INSERT INTO memory_relationships (id, repository_id, from_id, to_id, relationship_type, created_at) VALUES (?, ?, ?, ?, ?, datetime())", "rel_a", repoID, "fact_a", "dec_a", "FACT_SUPPORTS_DECISION")
	if err != nil {
		t.Fatalf("failed to insert relationship: %v", err)
	}

	// Seed contributor & expertise
	_, err = db.Exec("INSERT INTO contributors (id, repository_id, email, name, first_seen, last_seen, commit_count) VALUES (?, ?, ?, ?, datetime(), datetime(), ?)", "contrib_a", repoID, "dev@example.com", "Dev User", 5)
	if err != nil {
		t.Fatalf("failed to insert contributor: %v", err)
	}
	_, err = db.Exec("INSERT INTO expertise (id, repository_id, contributor_id, domain, score, evidence_json) VALUES (?, ?, ?, ?, ?, ?)", "exp_a", repoID, "contrib_a", "storage", 0.9, `{"commits":5}`)
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
	svc := NewService(discoverySvc, dr, fr, er, cr, expr, sr, relEngine)

	path, err := svc.GenerateRepositoryPath(ctx, repoID)
	if err != nil {
		t.Fatalf("GenerateRepositoryPath failed: %v", err)
	}

	// Expected Buckets order: Decisions -> Facts -> Events -> Contributors
	// Matches: dec_a -> fact_a -> contrib_a
	if len(path.Steps) != 3 {
		t.Fatalf("expected 3 steps, got %d", len(path.Steps))
	}

	if path.Steps[0].EntityType != EntityTypeDecision || path.Steps[0].EntityID != "dec_a" {
		t.Errorf("unexpected step 0: %+v", path.Steps[0])
	}
	if path.Steps[1].EntityType != EntityTypeFact || path.Steps[1].EntityID != "fact_a" {
		t.Errorf("unexpected step 1: %+v", path.Steps[1])
	}
	if path.Steps[2].EntityType != EntityTypeContributor || path.Steps[2].EntityID != "contrib_a" {
		t.Errorf("unexpected step 2: %+v", path.Steps[2])
	}
}
