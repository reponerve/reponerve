package discovery

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

// --- Unit Tests ---

func TestService_EmptyRepository(t *testing.T) {
	ctx := context.Background()

	relEngine := relationships.NewEngine(
		&mockDecisionReader{},
		&mockIntentReader{},
		&mockFactReader{},
		&mockEventReader{},
		&mockRelationshipReader{},
		&mockContributorReader{},
		&mockExpertiseReader{},
		&mockSourceReader{},
	)
	travEngine := traversal.NewEngine(relEngine)
	impactSvc := impact.NewService(travEngine)

	svc := NewService(
		&mockDecisionReader{},
		&mockFactReader{},
		&mockEventReader{},
		&mockContributorReader{},
		&mockExpertiseReader{},
		&mockRelationshipReader{},
		relEngine,
		travEngine,
		impactSvc,
	)

	report, err := svc.Discover(ctx, "repo_1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(report.Items) != 0 {
		t.Errorf("expected 0 discovery items, got %d", len(report.Items))
	}
}

func TestService_DecisionDiscovery(t *testing.T) {
	ctx := context.Background()
	repoID := "repo_1"

	decisions := []*memorymodels.Decision{
		{ID: "dec_1", RepositoryID: repoID, Title: "Use PostgreSQL", Status: "Accepted", CreatedAt: time.Now()},
	}

	relEngine := relationships.NewEngine(
		&mockDecisionReader{decisions: decisions},
		&mockIntentReader{},
		&mockFactReader{},
		&mockEventReader{},
		&mockRelationshipReader{},
		&mockContributorReader{},
		&mockExpertiseReader{},
		&mockSourceReader{},
	)
	travEngine := traversal.NewEngine(relEngine)
	impactSvc := impact.NewService(travEngine)

	svc := NewService(
		&mockDecisionReader{decisions: decisions},
		&mockFactReader{},
		&mockEventReader{},
		&mockContributorReader{},
		&mockExpertiseReader{},
		&mockRelationshipReader{},
		relEngine,
		travEngine,
		impactSvc,
	)

	report, err := svc.Discover(ctx, repoID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(report.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(report.Items))
	}

	item := report.Items[0]
	if item.EntityType != EntityTypeDecision || item.EntityID != "dec_1" {
		t.Errorf("unexpected item: %+v", item)
	}
	if item.Explanation != "This decision participates in 0 graph relationships and 0 impact paths." {
		t.Errorf("unexpected explanation: %q", item.Explanation)
	}

	var ev map[string]interface{}
	_ = json.Unmarshal([]byte(item.EvidenceJSON), &ev)
	if ev["graph_relationships"].(float64) != 0 || ev["impact_paths"].(float64) != 0 {
		t.Errorf("unexpected evidence: %s", item.EvidenceJSON)
	}
}

func TestService_FactDiscovery(t *testing.T) {
	ctx := context.Background()
	repoID := "repo_1"

	facts := []*memorymodels.Fact{
		{ID: "fact_1", RepositoryID: repoID, Subject: "auth", Predicate: "uses", Object: "JWT", CreatedAt: time.Now()},
	}

	relEngine := relationships.NewEngine(
		&mockDecisionReader{},
		&mockIntentReader{},
		&mockFactReader{facts: facts},
		&mockEventReader{},
		&mockRelationshipReader{},
		&mockContributorReader{},
		&mockExpertiseReader{},
		&mockSourceReader{},
	)
	travEngine := traversal.NewEngine(relEngine)
	impactSvc := impact.NewService(travEngine)

	svc := NewService(
		&mockDecisionReader{},
		&mockFactReader{facts: facts},
		&mockEventReader{},
		&mockContributorReader{},
		&mockExpertiseReader{},
		&mockRelationshipReader{},
		relEngine,
		travEngine,
		impactSvc,
	)

	report, err := svc.Discover(ctx, repoID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(report.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(report.Items))
	}

	item := report.Items[0]
	if item.EntityType != EntityTypeFact || item.EntityID != "fact_1" {
		t.Errorf("unexpected item: %+v", item)
	}
	if item.Explanation != "This fact participates in 0 repository knowledge chains." {
		t.Errorf("unexpected explanation: %q", item.Explanation)
	}
}

func TestService_ContributorDiscovery(t *testing.T) {
	ctx := context.Background()
	repoID := "repo_1"

	contribs := []*models.Contributor{
		{ID: "contrib_1", RepositoryID: repoID, Name: "Dev", Email: "dev@example.com", CommitCount: 5},
	}
	exps := []*models.Expertise{
		{ID: "exp_1", RepositoryID: repoID, ContributorID: "contrib_1", Domain: "auth", Score: 0.8},
		{ID: "exp_2", RepositoryID: repoID, ContributorID: "contrib_1", Domain: "storage", Score: 0.9},
	}

	relEngine := relationships.NewEngine(
		&mockDecisionReader{},
		&mockIntentReader{},
		&mockFactReader{},
		&mockEventReader{},
		&mockRelationshipReader{},
		&mockContributorReader{contribs: contribs},
		&mockExpertiseReader{expertise: exps},
		&mockSourceReader{},
	)
	travEngine := traversal.NewEngine(relEngine)
	impactSvc := impact.NewService(travEngine)

	svc := NewService(
		&mockDecisionReader{},
		&mockFactReader{},
		&mockEventReader{},
		&mockContributorReader{contribs: contribs},
		&mockExpertiseReader{expertise: exps},
		&mockRelationshipReader{},
		relEngine,
		travEngine,
		impactSvc,
	)

	report, err := svc.Discover(ctx, repoID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(report.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(report.Items))
	}

	item := report.Items[0]
	if item.EntityType != EntityTypeContributor || item.EntityID != "contrib_1" {
		t.Errorf("unexpected item: %+v", item)
	}
	if item.Explanation != "This contributor owns expertise across 2 repository domains." {
		t.Errorf("unexpected explanation: %q", item.Explanation)
	}

	var ev map[string]interface{}
	_ = json.Unmarshal([]byte(item.EvidenceJSON), &ev)
	if ev["expertise_count"].(float64) != 2 || ev["domains"].(float64) != 2 {
		t.Errorf("unexpected evidence: %s", item.EvidenceJSON)
	}
	if item.Score != 4.0 {
		t.Errorf("expected score 4.0, got %f", item.Score)
	}
}

func TestService_ScoringAndOrdering(t *testing.T) {
	ctx := context.Background()
	repoID := "repo_1"

	decisions := []*memorymodels.Decision{
		{ID: "dec_1", RepositoryID: repoID, Title: "Use PG", Status: "Accepted", CreatedAt: time.Now()},
		{ID: "dec_2", RepositoryID: repoID, Title: "Use WAL", Status: "Accepted", CreatedAt: time.Now()},
	}

	// We seed relationships to give dec_1 more graph connections
	rels := []*memorymodels.Relationship{
		{ID: "r1", RepositoryID: repoID, FromID: "dec_1", ToID: "dec_2", Type: "DECISION_SUPERCEDES_DECISION"},
	}

	relEngine := relationships.NewEngine(
		&mockDecisionReader{decisions: decisions},
		&mockIntentReader{},
		&mockFactReader{},
		&mockEventReader{},
		&mockRelationshipReader{rels: rels},
		&mockContributorReader{},
		&mockExpertiseReader{},
		&mockSourceReader{},
	)
	travEngine := traversal.NewEngine(relEngine)
	impactSvc := impact.NewService(travEngine)

	svc := NewService(
		&mockDecisionReader{decisions: decisions},
		&mockFactReader{},
		&mockEventReader{},
		&mockContributorReader{},
		&mockExpertiseReader{},
		&mockRelationshipReader{rels: rels},
		relEngine,
		travEngine,
		impactSvc,
	)

	report, err := svc.Discover(ctx, repoID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(report.Items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(report.Items))
	}

	// dec_1 has 1 graph rel (dec_1->dec_2), and impact paths = 1 (dec_1->dec_2)
	// dec_2 has 1 graph rel (dec_1->dec_2), and impact paths = 0
	// So dec_1 should rank higher than dec_2
	if report.Items[0].EntityID != "dec_1" || report.Items[1].EntityID != "dec_2" {
		t.Errorf("unexpected ordering of items: 0 = %s, 1 = %s", report.Items[0].EntityID, report.Items[1].EntityID)
	}
}

func TestValidateItem_Unit(t *testing.T) {
	// 1. Valid item
	item := &DiscoveryItem{
		EntityType:   "DECISION",
		EntityID:     "dec_1",
		Score:        1.0,
		EvidenceJSON: `{"graph_relationships":1}`,
		Explanation:  "explanation",
	}
	if err := ValidateItem(item); err != nil {
		t.Errorf("expected valid item, got error: %v", err)
	}

	// 2. Nil item
	if err := ValidateItem(nil); err == nil {
		t.Error("expected error for nil item")
	}

	// 3. Invalid EntityType
	item.EntityType = "UNKNOWN"
	if err := ValidateItem(item); err == nil {
		t.Error("expected error for invalid EntityType")
	}
	item.EntityType = "DECISION"

	// 4. Invalid Score
	item.Score = -1.0
	if err := ValidateItem(item); err == nil {
		t.Error("expected error for negative Score")
	}
	item.Score = 1.0

	// 5. Invalid JSON
	item.EvidenceJSON = "invalid_json"
	if err := ValidateItem(item); err == nil {
		t.Error("expected error for invalid JSON evidence")
	}
}

// --- Integration Tests ---

func TestService_Integration(t *testing.T) {
	ctx := context.Background()
	tempDir, err := os.MkdirTemp("", "reponerve-discovery-integration-*")
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

	svc := NewService(dr, fr, er, cr, expr, rr, relEngine, travEngine, impactSvc)

	report, err := svc.Discover(ctx, repoID)
	if err != nil {
		t.Fatalf("Discover returned error: %v", err)
	}

	// Expected items:
	// - dec_a (Decision): graph_relationships = 1 (fact_a->dec_a), impact_paths = 0. Score = 1.0
	// - fact_a (Fact): graph_relationships = 1 (fact_a->dec_a), impact_paths = 1 (fact_a->dec_a). Score = 2.0
	// - contrib_a (Contributor): expertise = 1, domain = 1. Score = 2.0
	// Expected sorting order:
	// 1. Score 2.0 (Fact: fact_a)
	// 2. Score 2.0 (Contributor: contrib_a) -- sorting by EntityType ascending: "CONTRIBUTOR" < "FACT"
	// 3. Score 1.0 (Decision: dec_a)
	if len(report.Items) != 3 {
		t.Fatalf("expected 3 items, got %d", len(report.Items))
	}

	item0 := report.Items[0]
	if item0.EntityType != EntityTypeContributor || item0.EntityID != "contrib_a" {
		t.Errorf("unexpected item at index 0: %+v", item0)
	}

	item1 := report.Items[1]
	if item1.EntityType != EntityTypeFact || item1.EntityID != "fact_a" {
		t.Errorf("unexpected item at index 1: %+v", item1)
	}

	item2 := report.Items[2]
	if item2.EntityType != EntityTypeDecision || item2.EntityID != "dec_a" {
		t.Errorf("unexpected item at index 2: %+v", item2)
	}
}
