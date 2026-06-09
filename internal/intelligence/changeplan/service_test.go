package changeplan

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
func (m *mockDecisionReader) ListAll(ctx context.Context) ([]*memorymodels.Decision, error) {
	return nil, nil
}

type mockIntentReader struct {
	intents []*memorymodels.Intent
}

func (m *mockIntentReader) GetByID(ctx context.Context, id string) (*memorymodels.Intent, error) {
	return nil, nil
}
func (m *mockIntentReader) ListByRepository(ctx context.Context, repoID string) ([]*memorymodels.Intent, error) {
	return m.intents, nil
}
func (m *mockIntentReader) ListAll(ctx context.Context) ([]*memorymodels.Intent, error) {
	return nil, nil
}

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
func (m *mockRelationshipReader) ListAll(ctx context.Context) ([]*memorymodels.Relationship, error) {
	return nil, nil
}

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

	return NewService(impactSvc)
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

	plan, err := svc.GenerateDecisionPlan(ctx, "repo_1", "dec_1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(plan.Items) != 0 {
		t.Errorf("expected 0 items, got %d", len(plan.Items))
	}
}

func TestService_DecisionPlan(t *testing.T) {
	ctx := context.Background()
	repoID := "repo_1"

	decisions := []*memorymodels.Decision{
		{ID: "dec_1", RepositoryID: repoID, Title: "Use SQLite selection db", Status: "Accepted", CreatedAt: time.Now()},
		{ID: "dec_2", RepositoryID: repoID, Title: "Database replication selection db", Status: "Accepted", CreatedAt: time.Now()},
	}

	// dec_2 depends on dec_1
	rels := []*memorymodels.Relationship{
		{ID: "rel_1", RepositoryID: repoID, FromID: "dec_1", ToID: "dec_2", Type: "DECISION_DEPENDS_ON"},
	}

	svc := buildTestService(
		&mockDecisionReader{decisions: decisions},
		&mockFactReader{},
		&mockEventReader{},
		&mockContributorReader{},
		&mockExpertiseReader{},
		&mockRelationshipReader{rels: rels},
		&mockSourceReader{},
	)

	plan, err := svc.GenerateDecisionPlan(ctx, repoID, "dec_1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Expected: dec_2 is impacted by changing dec_1 (Priority 1)
	if len(plan.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(plan.Items))
	}

	item := plan.Items[0]
	if item.EntityType != "DECISION" || item.EntityID != "dec_2" || item.Priority != 1 {
		t.Errorf("unexpected item values: %+v", item)
	}

	expectedExplanation := "This decision should be reviewed because it directly depends on the changed decision."
	if item.Explanation != expectedExplanation {
		t.Errorf("expected explanation %q, got %q", expectedExplanation, item.Explanation)
	}

	var ev map[string]interface{}
	_ = json.Unmarshal([]byte(item.EvidenceJSON), &ev)
	if ev["impact_path_length"].(float64) != 1 {
		t.Errorf("unexpected evidence path length: %s", item.EvidenceJSON)
	}
}

func TestService_FactPlan(t *testing.T) {
	ctx := context.Background()
	repoID := "repo_1"

	facts := []*memorymodels.Fact{
		{ID: "fact_1", RepositoryID: repoID, Subject: "db", Predicate: "is", Object: "WAL", CreatedAt: time.Now()},
	}

	svc := buildTestService(
		&mockDecisionReader{},
		&mockFactReader{facts: facts},
		&mockEventReader{},
		&mockContributorReader{},
		&mockExpertiseReader{},
		&mockRelationshipReader{},
		&mockSourceReader{},
	)

	plan, err := svc.GenerateFactPlan(ctx, repoID, "fact_1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(plan.Items) != 0 {
		t.Errorf("expected 0 items, got %d", len(plan.Items))
	}
}

func TestValidateItem_Unit(t *testing.T) {
	item := &ChangePlanItem{
		EntityType:   "DECISION",
		EntityID:     "dec_1",
		Priority:     1,
		EvidenceJSON: `{"impact_path_length":1}`,
		Explanation:  "explanation",
	}

	if err := ValidateItem(item); err != nil {
		t.Errorf("expected valid item, got: %v", err)
	}

	if err := ValidateItem(nil); err == nil {
		t.Error("expected error for nil item")
	}

	item.Priority = 0
	if err := ValidateItem(item); err == nil {
		t.Error("expected error for priority <= 0")
	}
	item.Priority = 1

	item.EntityType = "INVALID"
	if err := ValidateItem(item); err == nil {
		t.Error("expected error for invalid entity type")
	}
}

// --- Integration Tests ---

func TestService_Integration(t *testing.T) {
	ctx := context.Background()
	tempDir, err := os.MkdirTemp("", "reponerve-changeplan-integration-*")
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
	_, err = db.Exec("INSERT INTO memory_decisions (id, repository_id, source_id, title, status, created_at) VALUES (?, ?, ?, ?, ?, datetime())", "dec_b", repoID, "src_1", "Use Backup Replication", "Accepted")
	if err != nil {
		t.Fatalf("failed to insert decision: %v", err)
	}

	// Seed relationship
	_, err = db.Exec("INSERT INTO memory_relationships (id, repository_id, from_id, to_id, relationship_type, created_at) VALUES (?, ?, ?, ?, ?, datetime())", "rel_ab", repoID, "dec_a", "dec_b", "DECISION_DEPENDS_ON")
	if err != nil {
		t.Fatalf("failed to insert relationship: %v", err)
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

	svc := NewService(impactSvc)

	plan, err := svc.GenerateDecisionPlan(ctx, repoID, "dec_a")
	if err != nil {
		t.Fatalf("GenerateDecisionPlan failed: %v", err)
	}

	// Expected: dec_b should be in the plan (Priority 1)
	if len(plan.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(plan.Items))
	}

	if plan.Items[0].EntityID != "dec_b" || plan.Items[0].Priority != 1 {
		t.Errorf("unexpected item in plan: %+v", plan.Items[0])
	}
}
