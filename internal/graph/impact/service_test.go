package impact

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	memorymodels "github.com/reponerve/reponerve/internal/memory/models"
	"github.com/reponerve/reponerve/internal/graph/model"
	"github.com/reponerve/reponerve/internal/graph/relationships"
	"github.com/reponerve/reponerve/internal/graph/traversal"
	"github.com/reponerve/reponerve/internal/query/storage"
	"github.com/reponerve/reponerve/internal/storage/migrations"
	"github.com/reponerve/reponerve/internal/storage/sqlite"
	models "github.com/reponerve/reponerve/pkg/models"
)

// --- Mock Readers (copied from traversal tests to avoid cross-package dependency) ---

type mockDecisionReader struct {
	decisions []*memorymodels.Decision
}

func (m *mockDecisionReader) GetByID(ctx context.Context, id string) (*memorymodels.Decision, error) {
	return nil, nil
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
	return nil, nil
}
func (m *mockFactReader) ListByRepository(ctx context.Context, repoID string) ([]*memorymodels.Fact, error) {
	return m.facts, nil
}
func (m *mockFactReader) ListAll(ctx context.Context) ([]*memorymodels.Fact, error) {
	return nil, nil
}

type mockEventReader struct {
	events []*models.Event
}

func (m *mockEventReader) GetByID(ctx context.Context, id string) (*models.Event, error) {
	return nil, nil
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

// newTestService creates a Service backed by mock readers.
func newTestService(
	decisions []*memorymodels.Decision,
	intents []*memorymodels.Intent,
	facts []*memorymodels.Fact,
	events []*models.Event,
	rels []*memorymodels.Relationship,
	contribs []*models.Contributor,
	expertise []*models.Expertise,
	sources []*models.Source,
) *Service {
	relEngine := relationships.NewEngine(
		&mockDecisionReader{decisions: decisions},
		&mockIntentReader{intents: intents},
		&mockFactReader{facts: facts},
		&mockEventReader{events: events},
		&mockRelationshipReader{rels: rels},
		&mockContributorReader{contribs: contribs},
		&mockExpertiseReader{expertise: expertise},
		&mockSourceReader{sources: sources},
	)
	travEngine := traversal.NewEngine(relEngine)
	return NewService(travEngine)
}

// --- Unit Tests ---

func TestService_EmptyRepository(t *testing.T) {
	ctx := context.Background()
	svc := newTestService(nil, nil, nil, nil, nil, nil, nil, nil)

	report, err := svc.AnalyzeDecisionImpact(ctx, "repo_1", "dec_x")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(report.ImpactPaths) != 0 {
		t.Errorf("expected 0 impact paths for empty repo, got %d", len(report.ImpactPaths))
	}
}

func TestService_DecisionImpact_DecisionToDecision(t *testing.T) {
	ctx := context.Background()

	decisions := []*memorymodels.Decision{
		{ID: "dec_1", RepositoryID: "repo_1"},
		{ID: "dec_2", RepositoryID: "repo_1"},
	}
	rels := []*memorymodels.Relationship{
		{ID: "r1", RepositoryID: "repo_1", FromID: "dec_1", ToID: "dec_2", Type: "DECISION_DEPENDS_ON_DECISION"},
	}

	svc := newTestService(decisions, nil, nil, nil, rels, nil, nil, nil)

	report, err := svc.AnalyzeDecisionImpact(ctx, "repo_1", "dec_1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(report.ImpactPaths) != 1 {
		t.Fatalf("expected 1 impact path, got %d", len(report.ImpactPaths))
	}

	ip := report.ImpactPaths[0]
	if ip.Reason == "" {
		t.Error("expected non-empty reason")
	}
	if ip.Path == nil {
		t.Fatal("expected non-nil path")
	}
	if ip.Path.Nodes[0].EntityID != "dec_1" || ip.Path.Nodes[1].EntityID != "dec_2" {
		t.Errorf("unexpected path nodes: %v -> %v", ip.Path.Nodes[0].EntityID, ip.Path.Nodes[1].EntityID)
	}

	expectedReason := "Decision dec_1 impacts Decision dec_2 because Decision dec_2 depends on Decision dec_1."
	if ip.Reason != expectedReason {
		t.Errorf("unexpected reason:\n  got:  %q\n  want: %q", ip.Reason, expectedReason)
	}
}

func TestService_DecisionImpact_DecisionToEvent(t *testing.T) {
	ctx := context.Background()

	decisions := []*memorymodels.Decision{
		{ID: "dec_1", RepositoryID: "repo_1"},
	}
	events := []*models.Event{
		{ID: "event_1", RepositoryID: "repo_1"},
	}
	rels := []*memorymodels.Relationship{
		{ID: "r1", RepositoryID: "repo_1", FromID: "dec_1", ToID: "event_1", Type: "DECISION_RESULTS_IN_EVENT"},
	}

	svc := newTestService(decisions, nil, nil, events, rels, nil, nil, nil)

	report, err := svc.AnalyzeDecisionImpact(ctx, "repo_1", "dec_1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(report.ImpactPaths) != 1 {
		t.Fatalf("expected 1 impact path, got %d", len(report.ImpactPaths))
	}

	ip := report.ImpactPaths[0]
	expectedReason := "Decision dec_1 impacts Event event_1 because Event event_1 results from Decision dec_1."
	if ip.Reason != expectedReason {
		t.Errorf("unexpected reason:\n  got:  %q\n  want: %q", ip.Reason, expectedReason)
	}
}

func TestService_FactImpact_FactToFact(t *testing.T) {
	ctx := context.Background()

	facts := []*memorymodels.Fact{
		{ID: "fact_1", RepositoryID: "repo_1"},
		{ID: "fact_2", RepositoryID: "repo_1"},
	}
	rels := []*memorymodels.Relationship{
		{ID: "r1", RepositoryID: "repo_1", FromID: "fact_1", ToID: "fact_2", Type: "FACT_SUPPORTS_FACT"},
	}

	svc := newTestService(nil, nil, facts, nil, rels, nil, nil, nil)

	report, err := svc.AnalyzeFactImpact(ctx, "repo_1", "fact_1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(report.ImpactPaths) != 1 {
		t.Fatalf("expected 1 impact path, got %d", len(report.ImpactPaths))
	}

	ip := report.ImpactPaths[0]
	expectedReason := "Fact fact_1 impacts Fact fact_2 because Fact fact_2 is supported by Fact fact_1."
	if ip.Reason != expectedReason {
		t.Errorf("unexpected reason:\n  got:  %q\n  want: %q", ip.Reason, expectedReason)
	}
}

func TestService_FactImpact_FactToDecision(t *testing.T) {
	ctx := context.Background()

	facts := []*memorymodels.Fact{
		{ID: "fact_1", RepositoryID: "repo_1"},
	}
	decisions := []*memorymodels.Decision{
		{ID: "dec_1", RepositoryID: "repo_1"},
	}
	rels := []*memorymodels.Relationship{
		{ID: "r1", RepositoryID: "repo_1", FromID: "fact_1", ToID: "dec_1", Type: "FACT_SUPPORTS_DECISION"},
	}

	svc := newTestService(decisions, nil, facts, nil, rels, nil, nil, nil)

	report, err := svc.AnalyzeFactImpact(ctx, "repo_1", "fact_1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(report.ImpactPaths) != 1 {
		t.Fatalf("expected 1 impact path, got %d", len(report.ImpactPaths))
	}

	ip := report.ImpactPaths[0]
	expectedReason := "Fact fact_1 impacts Decision dec_1 because Decision dec_1 is supported by Fact fact_1."
	if ip.Reason != expectedReason {
		t.Errorf("unexpected reason:\n  got:  %q\n  want: %q", ip.Reason, expectedReason)
	}
}

func TestService_IntentImpact_IntentToDecision(t *testing.T) {
	ctx := context.Background()

	intents := []*memorymodels.Intent{
		{ID: "intent_1", RepositoryID: "repo_1"},
	}
	decisions := []*memorymodels.Decision{
		{ID: "dec_1", RepositoryID: "repo_1"},
	}
	rels := []*memorymodels.Relationship{
		{ID: "r1", RepositoryID: "repo_1", FromID: "intent_1", ToID: "dec_1", Type: "INTENT_DRIVES_DECISION"},
	}

	// Build engine directly to test reason generation for the Intent -> Decision pattern
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
	travEngine := traversal.NewEngine(relEngine)

	// Manually invoke with intent node to exercise reason generation (Intent -> Decision)
	nodeID := model.NodeID("repo_1", model.NodeTypeIntent, "intent_1")
	result, err := travEngine.FindDependencies(ctx, "repo_1", nodeID, traversal.TraversalOptions{IncludeStored: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	report := buildReport(result, "intent_1")

	if len(report.ImpactPaths) != 1 {
		t.Fatalf("expected 1 impact path, got %d", len(report.ImpactPaths))
	}

	ip := report.ImpactPaths[0]
	expectedReason := "Intent intent_1 impacts Decision dec_1 because Decision dec_1 is driven by Intent intent_1."
	if ip.Reason != expectedReason {
		t.Errorf("unexpected reason:\n  got:  %q\n  want: %q", ip.Reason, expectedReason)
	}
}

func TestService_ContributorImpact(t *testing.T) {
	ctx := context.Background()

	contribs := []*models.Contributor{
		{ID: "contrib_1", RepositoryID: "repo_1", Email: "alice@example.com", Name: "Alice"},
	}
	expertise := []*models.Expertise{
		{ID: "exp_1", RepositoryID: "repo_1", ContributorID: "contrib_1", Domain: "storage"},
	}

	svc := newTestService(nil, nil, nil, nil, nil, contribs, expertise, nil)

	report, err := svc.AnalyzeContributorImpact(ctx, "repo_1", "contrib_1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Contributor -> Expertise via inbound CONTRIBUTOR_EXPERT_IN_DOMAIN edge
	// The edge is Expertise -> Contributor (STORED), so FindDependents from Contributor
	// should find inbound edges from expertise to contributor.
	if len(report.ImpactPaths) != 1 {
		t.Fatalf("expected 1 impact path for contributor, got %d", len(report.ImpactPaths))
	}

	ip := report.ImpactPaths[0]
	if ip.Reason == "" {
		t.Error("expected non-empty reason for contributor impact path")
	}
	if ip.Path == nil {
		t.Fatal("expected non-nil path")
	}
}

func TestService_MultiHopPropagation(t *testing.T) {
	ctx := context.Background()

	decisions := []*memorymodels.Decision{
		{ID: "dec_1", RepositoryID: "repo_1"},
		{ID: "dec_2", RepositoryID: "repo_1"},
		{ID: "dec_3", RepositoryID: "repo_1"},
	}
	rels := []*memorymodels.Relationship{
		{ID: "r1", RepositoryID: "repo_1", FromID: "dec_1", ToID: "dec_2", Type: "DECISION_DEPENDS_ON_DECISION"},
		{ID: "r2", RepositoryID: "repo_1", FromID: "dec_2", ToID: "dec_3", Type: "DECISION_DEPENDS_ON_DECISION"},
	}

	svc := newTestService(decisions, nil, nil, nil, rels, nil, nil, nil)

	report, err := svc.AnalyzeDecisionImpact(ctx, "repo_1", "dec_1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Expected paths:
	// 1. dec_1 -> dec_2 (length 1)
	// 2. dec_1 -> dec_2 -> dec_3 (length 2)
	if len(report.ImpactPaths) != 2 {
		t.Fatalf("expected 2 impact paths, got %d", len(report.ImpactPaths))
	}

	if len(report.ImpactPaths[0].Path.Edges) != 1 {
		t.Errorf("expected first path to have 1 edge, got %d", len(report.ImpactPaths[0].Path.Edges))
	}
	if len(report.ImpactPaths[1].Path.Edges) != 2 {
		t.Errorf("expected second path to have 2 edges, got %d", len(report.ImpactPaths[1].Path.Edges))
	}
}

func TestService_SortingOrder(t *testing.T) {
	ctx := context.Background()

	decisions := []*memorymodels.Decision{
		{ID: "dec_1", RepositoryID: "repo_1"},
		{ID: "dec_2", RepositoryID: "repo_1"},
		{ID: "dec_3", RepositoryID: "repo_1"},
	}
	rels := []*memorymodels.Relationship{
		{ID: "r1", RepositoryID: "repo_1", FromID: "dec_1", ToID: "dec_3", Type: "DECISION_DEPENDS_ON_DECISION"},
		{ID: "r2", RepositoryID: "repo_1", FromID: "dec_1", ToID: "dec_2", Type: "DECISION_DEPENDS_ON_DECISION"},
	}

	svc := newTestService(decisions, nil, nil, nil, rels, nil, nil, nil)

	report, err := svc.AnalyzeDecisionImpact(ctx, "repo_1", "dec_1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(report.ImpactPaths) != 2 {
		t.Fatalf("expected 2 paths, got %d", len(report.ImpactPaths))
	}

	// Both are length 1 — should be sorted by ending node ID
	endID0 := report.ImpactPaths[0].Path.Nodes[1].ID
	endID1 := report.ImpactPaths[1].Path.Nodes[1].ID
	if endID0 >= endID1 {
		t.Errorf("paths not sorted by ending node ID: %s >= %s", endID0, endID1)
	}
}

func TestService_EvidencePreservation(t *testing.T) {
	ctx := context.Background()

	decisions := []*memorymodels.Decision{
		{ID: "dec_1", RepositoryID: "repo_1"},
		{ID: "dec_2", RepositoryID: "repo_1"},
	}
	rels := []*memorymodels.Relationship{
		{ID: "r1", RepositoryID: "repo_1", FromID: "dec_1", ToID: "dec_2", Type: "DECISION_DEPENDS_ON_DECISION"},
	}

	svc := newTestService(decisions, nil, nil, nil, rels, nil, nil, nil)

	report, err := svc.AnalyzeDecisionImpact(ctx, "repo_1", "dec_1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(report.ImpactPaths) == 0 {
		t.Fatal("expected at least one impact path")
	}

	// Verify edge evidence is non-empty
	for _, ip := range report.ImpactPaths {
		for _, edge := range ip.Path.Edges {
			if edge.EvidenceJSON == "" {
				t.Errorf("expected non-empty EvidenceJSON for edge %s", edge.ID)
			}
		}
	}
}

func TestService_GenericFallbackReason(t *testing.T) {
	ctx := context.Background()

	// Event -> Decision (no specific pattern defined — will use fallback)
	decisions := []*memorymodels.Decision{
		{ID: "dec_1", RepositoryID: "repo_1"},
	}
	events := []*models.Event{
		{ID: "event_1", RepositoryID: "repo_1"},
	}
	rels := []*memorymodels.Relationship{
		{ID: "r1", RepositoryID: "repo_1", FromID: "event_1", ToID: "dec_1", Type: "RELATES_TO"},
	}

	svc := newTestService(decisions, nil, nil, events, rels, nil, nil, nil)

	report, err := svc.AnalyzeEventImpact(ctx, "repo_1", "event_1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(report.ImpactPaths) != 1 {
		t.Fatalf("expected 1 impact path, got %d", len(report.ImpactPaths))
	}

	ip := report.ImpactPaths[0]
	// Should use the generic fallback
	expectedReason := "EVENT event_1 impacts DECISION dec_1 through traversal path."
	if ip.Reason != expectedReason {
		t.Errorf("unexpected reason:\n  got:  %q\n  want: %q", ip.Reason, expectedReason)
	}
}

// --- Integration Test ---

func TestService_Integration(t *testing.T) {
	ctx := context.Background()

	tempDir, err := os.MkdirTemp("", "reponerve-impact-integration-*")
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

	repoID := "repo_impact_int"

	// Seed repository
	_, err = db.Exec("INSERT INTO repositories (id, name, path, default_branch, created_at, updated_at) VALUES (?, ?, ?, ?, datetime(), datetime())", repoID, "Impact Test Repo", tempDir, "main")
	if err != nil {
		t.Fatalf("failed to insert repository: %v", err)
	}

	// Seed sources
	_, err = db.Exec("INSERT INTO sources (id, repository_id, source_type, reference, title, author, timestamp, created_at) VALUES (?, ?, ?, ?, ?, ?, datetime(), datetime())", "src_a", repoID, "adr", "docs/adr/0001.md", "ADR-1", "Alice", "2024-01-01T00:00:00Z")
	if err != nil {
		t.Fatalf("failed to insert source: %v", err)
	}
	_, err = db.Exec("INSERT INTO sources (id, repository_id, source_type, reference, title, author, timestamp, created_at) VALUES (?, ?, ?, ?, ?, ?, datetime(), datetime())", "src_b", repoID, "adr", "docs/adr/0002.md", "ADR-2", "Alice", "2024-01-02T00:00:00Z")
	if err != nil {
		t.Fatalf("failed to insert source b: %v", err)
	}

	// Seed decisions
	_, err = db.Exec("INSERT INTO memory_decisions (id, repository_id, source_id, title, status, created_at) VALUES (?, ?, ?, ?, ?, datetime())", "dec_a", repoID, "src_a", "Use PostgreSQL", "Accepted")
	if err != nil {
		t.Fatalf("failed to insert dec_a: %v", err)
	}
	_, err = db.Exec("INSERT INTO memory_decisions (id, repository_id, source_id, title, status, created_at) VALUES (?, ?, ?, ?, ?, datetime())", "dec_b", repoID, "src_b", "Use Connection Pooling", "Accepted")
	if err != nil {
		t.Fatalf("failed to insert dec_b: %v", err)
	}

	// Seed events
	_, err = db.Exec("INSERT INTO memory_events (id, repository_id, event_type, title, description, source_id, timestamp, created_at) VALUES (?, ?, ?, ?, ?, ?, datetime(), datetime())", "event_1", repoID, "commit", "Initial DB setup", "Setup DB", "src_a")
	if err != nil {
		t.Fatalf("failed to insert event: %v", err)
	}

	// Seed facts
	_, err = db.Exec("INSERT INTO memory_facts (id, repository_id, source_id, subject, predicate, object, created_at) VALUES (?, ?, ?, ?, ?, ?, datetime())", "fact_1", repoID, "src_a", "PostgreSQL", "supports", "JSONB")
	if err != nil {
		t.Fatalf("failed to insert fact: %v", err)
	}

	// Seed contributors
	_, err = db.Exec("INSERT INTO contributors (id, repository_id, email, name, first_seen, last_seen, commit_count) VALUES (?, ?, ?, ?, datetime(), datetime(), ?)", "contrib_1", repoID, "alice@example.com", "Alice", 10)
	if err != nil {
		t.Fatalf("failed to insert contributor: %v", err)
	}

	// Seed expertise
	_, err = db.Exec("INSERT INTO expertise (id, repository_id, contributor_id, domain, score, evidence_json) VALUES (?, ?, ?, ?, ?, ?)", "exp_1", repoID, "contrib_1", "storage", 10.0, `{"commits":5}`)
	if err != nil {
		t.Fatalf("failed to insert expertise: %v", err)
	}

	// Seed stored relationships: dec_a -> dec_b, dec_a -> event_1, fact_1 -> dec_a
	createdAt := time.Now()
	_, err = db.Exec("INSERT INTO memory_relationships (id, repository_id, from_id, to_id, relationship_type, created_at) VALUES (?, ?, ?, ?, ?, datetime())", "rel_1", repoID, "dec_a", "dec_b", "DECISION_DEPENDS_ON_DECISION", createdAt)
	if err != nil {
		t.Fatalf("failed to insert rel_1: %v", err)
	}
	_, err = db.Exec("INSERT INTO memory_relationships (id, repository_id, from_id, to_id, relationship_type, created_at) VALUES (?, ?, ?, ?, ?, datetime())", "rel_2", repoID, "dec_a", "event_1", "DECISION_RESULTS_IN_EVENT", createdAt)
	if err != nil {
		t.Fatalf("failed to insert rel_2: %v", err)
	}
	_, err = db.Exec("INSERT INTO memory_relationships (id, repository_id, from_id, to_id, relationship_type, created_at) VALUES (?, ?, ?, ?, ?, datetime())", "rel_3", repoID, "fact_1", "dec_a", "FACT_SUPPORTS_DECISION", createdAt)
	if err != nil {
		t.Fatalf("failed to insert rel_3: %v", err)
	}

	// Instantiate real readers
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
	svc := NewService(travEngine)

	t.Run("DecisionImpact", func(t *testing.T) {
		report, err := svc.AnalyzeDecisionImpact(ctx, repoID, "dec_a")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// dec_a -> dec_b (len 1)
		// dec_a -> event_1 (len 1)
		// dec_a -> dec_b -> (no further edges)
		if len(report.ImpactPaths) < 2 {
			t.Errorf("expected at least 2 impact paths, got %d", len(report.ImpactPaths))
		}

		// All paths must have non-empty reasons and evidence
		for i, ip := range report.ImpactPaths {
			if ip.Reason == "" {
				t.Errorf("path %d: empty reason", i)
			}
			for _, edge := range ip.Path.Edges {
				if edge.EvidenceJSON == "" {
					t.Errorf("path %d: empty evidence JSON on edge %s", i, edge.ID)
				}
			}
		}
	})

	t.Run("FactImpact", func(t *testing.T) {
		report, err := svc.AnalyzeFactImpact(ctx, repoID, "fact_1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// fact_1 -> dec_a (len 1) with reason: Fact impacts Decision
		// fact_1 -> dec_a -> dec_b (len 2)
		// fact_1 -> dec_a -> event_1 (len 2)
		if len(report.ImpactPaths) < 1 {
			t.Errorf("expected at least 1 impact path for fact, got %d", len(report.ImpactPaths))
		}

		// First path should be fact -> decision
		ip0 := report.ImpactPaths[0]
		if ip0.Path.Nodes[0].EntityID != "fact_1" {
			t.Errorf("expected start node fact_1, got %s", ip0.Path.Nodes[0].EntityID)
		}
		expectedReason := "Fact fact_1 impacts Decision dec_a because Decision dec_a is supported by Fact fact_1."
		if ip0.Reason != expectedReason {
			t.Errorf("unexpected reason:\n  got:  %q\n  want: %q", ip0.Reason, expectedReason)
		}
	})

	t.Run("ContributorImpact", func(t *testing.T) {
		report, err := svc.AnalyzeContributorImpact(ctx, repoID, "contrib_1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Contributor contrib_1 <- exp_1 (CONTRIBUTOR_EXPERT_IN_DOMAIN, inbound edge)
		// FindDependents from contrib_1 should find expertise node exp_1
		if len(report.ImpactPaths) != 1 {
			t.Fatalf("expected 1 impact path for contributor, got %d", len(report.ImpactPaths))
		}

		ip := report.ImpactPaths[0]
		if ip.Reason == "" {
			t.Error("expected non-empty reason for contributor impact path")
		}
	})
}
