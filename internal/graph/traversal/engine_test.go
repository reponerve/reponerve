package traversal

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/reponerve/reponerve/internal/graph/model"
	"github.com/reponerve/reponerve/internal/graph/relationships"
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
func (m *mockFactReader) ListAll(ctx context.Context) ([]*memorymodels.Fact, error) { return nil, nil }

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

// --- Unit Tests ---

func TestEngine_EmptyGraph(t *testing.T) {
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
	engine := NewEngine(relEngine)

	res, err := engine.TraceGraph(ctx, "repo_1", "start", TraversalOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res.Paths) != 0 {
		t.Errorf("expected 0 paths, got %d", len(res.Paths))
	}
}

func TestEngine_DependencyAndDependent(t *testing.T) {
	ctx := context.Background()

	intents := []*memorymodels.Intent{{ID: "intent_1", RepositoryID: "repo_1"}}
	decisions := []*memorymodels.Decision{{ID: "dec_1", RepositoryID: "repo_1"}}
	events := []*models.Event{{ID: "event_1", RepositoryID: "repo_1"}}

	rels := []*memorymodels.Relationship{
		{ID: "r1", RepositoryID: "repo_1", FromID: "intent_1", ToID: "dec_1", Type: "INTENT_DRIVES_DECISION"},
		{ID: "r2", RepositoryID: "repo_1", FromID: "dec_1", ToID: "event_1", Type: "DECISION_RESULTS_IN_EVENT"},
	}

	relEngine := relationships.NewEngine(
		&mockDecisionReader{decisions: decisions},
		&mockIntentReader{intents: intents},
		&mockFactReader{},
		&mockEventReader{events: events},
		&mockRelationshipReader{rels: rels},
		&mockContributorReader{},
		&mockExpertiseReader{},
		&mockSourceReader{},
	)
	engine := NewEngine(relEngine)

	// Test Outbound: FindDependencies from intent_1
	startNodeID := model.NodeID("repo_1", model.NodeTypeIntent, "intent_1")
	res, err := engine.FindDependencies(ctx, "repo_1", startNodeID, TraversalOptions{IncludeStored: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Paths expected:
	// 1. intent_1 -> dec_1 (length 1)
	// 2. intent_1 -> dec_1 -> event_1 (length 2)
	if len(res.Paths) != 2 {
		t.Fatalf("expected 2 dependency paths, got %d", len(res.Paths))
	}

	p0 := res.Paths[0]
	if len(p0.Edges) != 1 || p0.Nodes[0].EntityID != "intent_1" || p0.Nodes[1].EntityID != "dec_1" {
		t.Errorf("unexpected path 0: %+v", p0)
	}

	p1 := res.Paths[1]
	if len(p1.Edges) != 2 || p1.Nodes[0].EntityID != "intent_1" || p1.Nodes[1].EntityID != "dec_1" || p1.Nodes[2].EntityID != "event_1" {
		t.Errorf("unexpected path 1: %+v", p1)
	}

	// Test Inbound: FindDependents for event_1
	targetNodeID := model.NodeID("repo_1", model.NodeTypeEvent, "event_1")
	resDep, err := engine.FindDependents(ctx, "repo_1", targetNodeID, TraversalOptions{IncludeStored: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Paths expected:
	// 1. dec_1 -> event_1 (length 1)
	// 2. intent_1 -> dec_1 -> event_1 (length 2)
	if len(resDep.Paths) != 2 {
		t.Fatalf("expected 2 dependent paths, got %d", len(resDep.Paths))
	}

	dp0 := resDep.Paths[0]
	if len(dp0.Edges) != 1 || dp0.Nodes[0].EntityID != "dec_1" || dp0.Nodes[1].EntityID != "event_1" {
		t.Errorf("unexpected dependent path 0: %+v", dp0)
	}

	dp1 := resDep.Paths[1]
	if len(dp1.Edges) != 2 || dp1.Nodes[0].EntityID != "intent_1" || dp1.Nodes[1].EntityID != "dec_1" || dp1.Nodes[2].EntityID != "event_1" {
		t.Errorf("unexpected dependent path 1: %+v", dp1)
	}
}

func TestEngine_EdgeFiltering(t *testing.T) {
	ctx := context.Background()

	intents := []*memorymodels.Intent{{ID: "intent_1", RepositoryID: "repo_1"}}
	decisions := []*memorymodels.Decision{
		{ID: "dec_1", RepositoryID: "repo_1", SourceID: "src_a"},
		{ID: "dec_2", RepositoryID: "repo_1", SourceID: "src_b"},
	}

	// Stored relationship
	rels := []*memorymodels.Relationship{
		{ID: "r1", RepositoryID: "repo_1", FromID: "intent_1", ToID: "dec_1", Type: "INTENT_DRIVES_DECISION"},
	}

	// Derived relationship setup: dec_1 depends on dec_2
	metaBytes, _ := json.Marshal(struct {
		Content string `json:"content"`
	}{
		Content: "Depends on decision_b", // Wait! The ID of dec_2 is dec_2, not decision_b. Let's make it dec_2.
	})
	// Change content to match dec_2 ID
	metaBytes, _ = json.Marshal(struct {
		Content string `json:"content"`
	}{
		Content: "Depends on: dec_2",
	})

	sources := []*models.Source{
		{ID: "src_a", RepositoryID: "repo_1", SourceType: "adr", Reference: "docs/adr/0001.md", MetadataJSON: string(metaBytes)},
	}

	relEngine := relationships.NewEngine(
		&mockDecisionReader{decisions: decisions},
		&mockIntentReader{intents: intents},
		&mockFactReader{},
		&mockEventReader{},
		&mockRelationshipReader{rels: rels},
		&mockContributorReader{},
		&mockExpertiseReader{},
		&mockSourceReader{sources: sources},
	)
	engine := NewEngine(relEngine)

	startNodeID := model.NodeID("repo_1", model.NodeTypeIntent, "intent_1")

	// 1. Both Stored & Derived
	resBoth, _ := engine.TraceGraph(ctx, "repo_1", startNodeID, TraversalOptions{IncludeStored: true, IncludeDerived: true})
	// intent_1 -> dec_1 (stored)
	// intent_1 -> dec_1 -> dec_2 (stored -> derived)
	if len(resBoth.Paths) != 2 {
		t.Errorf("expected 2 paths with both, got %d", len(resBoth.Paths))
	}

	// 2. Stored Only
	resStored, _ := engine.TraceGraph(ctx, "repo_1", startNodeID, TraversalOptions{IncludeStored: true, IncludeDerived: false})
	if len(resStored.Paths) != 1 {
		t.Errorf("expected 1 path with stored only, got %d", len(resStored.Paths))
	}
	if resStored.Paths[0].Edges[0].Category != model.CategoryStored {
		t.Errorf("expected stored edge, got %s", resStored.Paths[0].Edges[0].Category)
	}

	// 3. Derived Only
	// Note: TraceGraph starts from intent_1, which has no outbound derived edges.
	// So starting from dec_1
	dec1NodeID := model.NodeID("repo_1", model.NodeTypeDecision, "dec_1")
	resDerived, _ := engine.TraceGraph(ctx, "repo_1", dec1NodeID, TraversalOptions{IncludeStored: false, IncludeDerived: true})
	if len(resDerived.Paths) != 1 {
		t.Errorf("expected 1 path starting from dec_1, got %d", len(resDerived.Paths))
	}
	if resDerived.Paths[0].Edges[0].Category != model.CategoryDerived {
		t.Errorf("expected derived edge, got %s", resDerived.Paths[0].Edges[0].Category)
	}
}

func TestEngine_CycleHandling(t *testing.T) {
	ctx := context.Background()

	decisions := []*memorymodels.Decision{
		{ID: "dec_1", RepositoryID: "repo_1"},
		{ID: "dec_2", RepositoryID: "repo_1"},
		{ID: "dec_3", RepositoryID: "repo_1"},
	}

	// Cyclic stored relationships: dec_1 -> dec_2 -> dec_3 -> dec_1
	rels := []*memorymodels.Relationship{
		{ID: "r1", RepositoryID: "repo_1", FromID: "dec_1", ToID: "dec_2", Type: "DECISION_DEPENDS_ON_DECISION"},
		{ID: "r2", RepositoryID: "repo_1", FromID: "dec_2", ToID: "dec_3", Type: "DECISION_DEPENDS_ON_DECISION"},
		{ID: "r3", RepositoryID: "repo_1", FromID: "dec_3", ToID: "dec_1", Type: "DECISION_DEPENDS_ON_DECISION"},
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
	engine := NewEngine(relEngine)

	dec1NodeID := model.NodeID("repo_1", model.NodeTypeDecision, "dec_1")
	res, err := engine.TraceGraph(ctx, "repo_1", dec1NodeID, TraversalOptions{IncludeStored: true, MaxDepth: 10})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// We expect:
	// 1. dec_1 -> dec_2 (len 1)
	// 2. dec_1 -> dec_2 -> dec_3 (len 2)
	// (dec_3 -> dec_1 is ignored because it would repeat dec_1)
	if len(res.Paths) != 2 {
		t.Fatalf("expected 2 paths, got %d", len(res.Paths))
	}
}

func TestEngine_MaxDepth(t *testing.T) {
	ctx := context.Background()

	decisions := []*memorymodels.Decision{
		{ID: "dec_1", RepositoryID: "repo_1"},
		{ID: "dec_2", RepositoryID: "repo_1"},
		{ID: "dec_3", RepositoryID: "repo_1"},
		{ID: "dec_4", RepositoryID: "repo_1"},
	}

	rels := []*memorymodels.Relationship{
		{ID: "r1", RepositoryID: "repo_1", FromID: "dec_1", ToID: "dec_2", Type: "DECISION_DEPENDS_ON_DECISION"},
		{ID: "r2", RepositoryID: "repo_1", FromID: "dec_2", ToID: "dec_3", Type: "DECISION_DEPENDS_ON_DECISION"},
		{ID: "r3", RepositoryID: "repo_1", FromID: "dec_3", ToID: "dec_4", Type: "DECISION_DEPENDS_ON_DECISION"},
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
	engine := NewEngine(relEngine)

	dec1NodeID := model.NodeID("repo_1", model.NodeTypeDecision, "dec_1")

	// Limit to depth 2
	res, _ := engine.TraceGraph(ctx, "repo_1", dec1NodeID, TraversalOptions{IncludeStored: true, MaxDepth: 2})
	if len(res.Paths) != 2 {
		t.Errorf("expected 2 paths with depth limit 2, got %d", len(res.Paths))
	}
}

func TestEngine_Sorting(t *testing.T) {
	ctx := context.Background()

	// Setup two decisions dec_a and dec_b, and intent_1 pointing to both
	intents := []*memorymodels.Intent{{ID: "intent_1", RepositoryID: "repo_1"}}
	decisions := []*memorymodels.Decision{
		{ID: "dec_b", RepositoryID: "repo_1"},
		{ID: "dec_a", RepositoryID: "repo_1"},
	}

	rels := []*memorymodels.Relationship{
		{ID: "r1", RepositoryID: "repo_1", FromID: "intent_1", ToID: "dec_b", Type: "INTENT_DRIVES_DECISION"},
		{ID: "r2", RepositoryID: "repo_1", FromID: "intent_1", ToID: "dec_a", Type: "INTENT_DRIVES_DECISION"},
	}

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
	engine := NewEngine(relEngine)

	startNodeID := model.NodeID("repo_1", model.NodeTypeIntent, "intent_1")
	res, _ := engine.TraceGraph(ctx, "repo_1", startNodeID, TraversalOptions{IncludeStored: true})

	if len(res.Paths) != 2 {
		t.Fatalf("expected 2 paths, got %d", len(res.Paths))
	}

	// Nodes should sort by Ending node ID (since Starting Node ID is same)
	node0 := res.Paths[0].Nodes[1].ID
	node1 := res.Paths[1].Nodes[1].ID
	if node0 >= node1 {
		t.Errorf("paths not sorted by ending node ID: path 0 ID = %s, path 1 ID = %s", node0, node1)
	}
}

// --- Integration Tests ---

func TestEngine_Integration(t *testing.T) {
	ctx := context.Background()
	tempDir, err := os.MkdirTemp("", "reponerve-traversal-integration-*")
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

	// Seed sources
	// Source A references dec_b via markdown link [dec_b]
	metaA := struct {
		Content string `json:"content"`
		Status  string `json:"status"`
		Path    string `json:"path"`
	}{
		Content: "Depends on: dec_b",
		Status:  "Accepted",
		Path:    "docs/adr/0001.md",
	}
	metaBytesA, _ := json.Marshal(metaA)

	_, err = db.Exec("INSERT INTO sources (id, repository_id, source_type, reference, title, author, timestamp, created_at) VALUES (?, ?, ?, ?, ?, ?, datetime(), datetime())", "src_a", repoID, "adr", "docs/adr/0001.md", "Title A", "Author", "2024-01-01T00:00:00Z")
	if err != nil {
		t.Fatalf("failed to insert source a: %v", err)
	}
	_, err = db.Exec("UPDATE sources SET metadata_json = ? WHERE id = ?", string(metaBytesA), "src_a")
	if err != nil {
		t.Fatalf("failed to update source metadata a: %v", err)
	}

	_, err = db.Exec("INSERT INTO sources (id, repository_id, source_type, reference, title, author, timestamp, created_at) VALUES (?, ?, ?, ?, ?, ?, datetime(), datetime())", "src_b", repoID, "adr", "docs/adr/0002.md", "Title B", "Author", "2024-01-01T00:00:00Z")
	if err != nil {
		t.Fatalf("failed to insert source b: %v", err)
	}

	// Seed decisions
	_, err = db.Exec("INSERT INTO memory_decisions (id, repository_id, source_id, title, status, created_at) VALUES (?, ?, ?, ?, ?, datetime())", "dec_a", repoID, "src_a", "Use SQLite", "Accepted")
	if err != nil {
		t.Fatalf("failed to insert decision a: %v", err)
	}
	_, err = db.Exec("INSERT INTO memory_decisions (id, repository_id, source_id, title, status, created_at) VALUES (?, ?, ?, ?, ?, datetime())", "dec_b", repoID, "src_b", "Use WAL mode", "Accepted")
	if err != nil {
		t.Fatalf("failed to insert decision b: %v", err)
	}

	// Seed intents
	_, err = db.Exec("INSERT INTO memory_intents (id, repository_id, source_id, description, created_at) VALUES (?, ?, ?, ?, datetime())", "intent_x", repoID, "src_a", "Aggregate query logs")
	if err != nil {
		t.Fatalf("failed to insert intent_x: %v", err)
	}

	// Seed events
	_, err = db.Exec("INSERT INTO memory_events (id, repository_id, event_type, title, description, source_id, timestamp, created_at) VALUES (?, ?, ?, ?, ?, ?, datetime(), datetime())", "event_y", repoID, "commit", "Commit Title", "Commit Body", "src_b")
	if err != nil {
		t.Fatalf("failed to insert event_y: %v", err)
	}

	// Seed stored relationships
	// 1. intent_x drives dec_a
	// 2. dec_b results in event_y
	rel1 := &memorymodels.Relationship{
		ID:           "rel_1",
		RepositoryID: repoID,
		FromID:       "intent_x",
		ToID:         "dec_a",
		Type:         "INTENT_DRIVES_DECISION",
		CreatedAt:    time.Now(),
	}
	rel2 := &memorymodels.Relationship{
		ID:           "rel_2",
		RepositoryID: repoID,
		FromID:       "dec_b",
		ToID:         "event_y",
		Type:         "DECISION_RESULTS_IN_EVENT",
		CreatedAt:    time.Now(),
	}

	_, err = db.Exec("INSERT INTO memory_relationships (id, repository_id, from_id, to_id, relationship_type, created_at) VALUES (?, ?, ?, ?, ?, datetime())", rel1.ID, repoID, rel1.FromID, rel1.ToID, rel1.Type)
	if err != nil {
		t.Fatalf("failed to insert stored relationship 1: %v", err)
	}
	_, err = db.Exec("INSERT INTO memory_relationships (id, repository_id, from_id, to_id, relationship_type, created_at) VALUES (?, ?, ?, ?, ?, datetime())", rel2.ID, repoID, rel2.FromID, rel2.ToID, rel2.Type)
	if err != nil {
		t.Fatalf("failed to insert stored relationship 2: %v", err)
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
	engine := NewEngine(relEngine)

	// Run Outbound TraceGraph starting at intent_x
	startNodeID := model.NodeID(repoID, model.NodeTypeIntent, "intent_x")
	res, err := engine.TraceGraph(ctx, repoID, startNodeID, TraversalOptions{IncludeStored: true, IncludeDerived: true})
	if err != nil {
		t.Fatalf("unexpected error tracing graph: %v", err)
	}

	// Reachable chain:
	// intent_x
	// ↓ (stored INTENT_DRIVES_DECISION)
	// dec_a
	// ↓ (derived DECISION_DEPENDS_ON_DECISION)
	// dec_b
	// ↓ (stored DECISION_RESULTS_IN_EVENT)
	// event_y
	//
	// Expected paths:
	// 1. intent_x -> dec_a (len 1)
	// 2. intent_x -> dec_a -> dec_b (len 2)
	// 3. intent_x -> dec_a -> dec_b -> event_y (len 3)
	if len(res.Paths) != 3 {
		t.Fatalf("expected 3 paths in trace graph, got %d", len(res.Paths))
	}

	p0 := res.Paths[0]
	if len(p0.Edges) != 1 || p0.Nodes[0].EntityID != "intent_x" || p0.Nodes[1].EntityID != "dec_a" {
		t.Errorf("unexpected path 0: %+v", p0)
	}

	p1 := res.Paths[1]
	if len(p1.Edges) != 2 || p1.Nodes[2].EntityID != "dec_b" {
		t.Errorf("unexpected path 1: %+v", p1)
	}

	p2 := res.Paths[2]
	if len(p2.Edges) != 3 || p2.Nodes[3].EntityID != "event_y" {
		t.Errorf("unexpected path 2: %+v", p2)
	}
}
