package relationships

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/reponerve/reponerve/internal/graph/model"
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

type mockIntentReader struct{}

func (m *mockIntentReader) GetByID(ctx context.Context, id string) (*memorymodels.Intent, error) {
	return nil, nil
}
func (m *mockIntentReader) ListByRepository(ctx context.Context, repoID string) ([]*memorymodels.Intent, error) {
	return nil, nil
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

type mockEventReader struct{}

func (m *mockEventReader) GetByID(ctx context.Context, id string) (*models.Event, error) {
	return nil, nil
}
func (m *mockEventReader) ListByRepository(ctx context.Context, repoID string) ([]*models.Event, error) {
	return nil, nil
}
func (m *mockEventReader) ListAll(ctx context.Context) ([]*models.Event, error) { return nil, nil }

type mockRelationshipReader struct{}

func (m *mockRelationshipReader) GetByID(ctx context.Context, id string) (*memorymodels.Relationship, error) {
	return nil, nil
}
func (m *mockRelationshipReader) ListByRepository(ctx context.Context, repoID string) ([]*memorymodels.Relationship, error) {
	return nil, nil
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

func TestEngine_EmptyRepo(t *testing.T) {
	ctx := context.Background()
	engine := NewEngine(
		&mockDecisionReader{},
		&mockIntentReader{},
		&mockFactReader{},
		&mockEventReader{},
		&mockRelationshipReader{},
		&mockContributorReader{},
		&mockExpertiseReader{},
		&mockSourceReader{},
	)

	rels, err := engine.Generate(ctx, "repo_1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rels) != 0 {
		t.Errorf("expected 0 relationships, got %d", len(rels))
	}
}

func TestEngine_DecisionDependency(t *testing.T) {
	ctx := context.Background()

	// Content A has explicit prefix before the reference path docs/adr/0002.md
	metaA := adrMetadata{
		Content: "This depends on: docs/adr/0002.md",
		Status:  "Accepted",
		Path:    "docs/adr/0001.md",
	}
	metaBytesA, _ := json.Marshal(metaA)

	metaB := adrMetadata{
		Content: "Setup caching",
		Status:  "Accepted",
		Path:    "docs/adr/0002.md",
	}
	metaBytesB, _ := json.Marshal(metaB)

	// Content C contains path docs/adr/0002.md with no explicit prefix
	metaC := adrMetadata{
		Content: "Referencing docs/adr/0002.md directly.",
		Status:  "Accepted",
		Path:    "docs/adr/0003.md",
	}
	metaBytesC, _ := json.Marshal(metaC)

	// Content D contains decision_b with no explicit prefix
	metaD := adrMetadata{
		Content: "Referencing decision_b directly.",
		Status:  "Accepted",
		Path:    "docs/adr/0004.md",
	}
	metaBytesD, _ := json.Marshal(metaD)

	sources := []*models.Source{
		{ID: "src_a", RepositoryID: "repo_1", SourceType: "adr", Reference: "docs/adr/0001.md", MetadataJSON: string(metaBytesA)},
		{ID: "src_b", RepositoryID: "repo_1", SourceType: "adr", Reference: "docs/adr/0002.md", MetadataJSON: string(metaBytesB)},
		{ID: "src_c", RepositoryID: "repo_1", SourceType: "adr", Reference: "docs/adr/0003.md", MetadataJSON: string(metaBytesC)},
		{ID: "src_d", RepositoryID: "repo_1", SourceType: "adr", Reference: "docs/adr/0004.md", MetadataJSON: string(metaBytesD)},
	}

	decisions := []*memorymodels.Decision{
		{ID: "decision_a", RepositoryID: "repo_1", SourceID: "src_a", Title: "Decision A"},
		{ID: "decision_b", RepositoryID: "repo_1", SourceID: "src_b", Title: "Decision B"},
		{ID: "decision_c", RepositoryID: "repo_1", SourceID: "src_c", Title: "Decision C"},
		{ID: "decision_d", RepositoryID: "repo_1", SourceID: "src_d", Title: "Decision D"},
	}

	engine := NewEngine(
		&mockDecisionReader{decisions: decisions},
		&mockIntentReader{},
		&mockFactReader{},
		&mockEventReader{},
		&mockRelationshipReader{},
		&mockContributorReader{},
		&mockExpertiseReader{},
		&mockSourceReader{sources: sources},
	)

	rels, err := engine.Generate(ctx, "repo_1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// We expect 3 relationships:
	// 1. decision_a -> decision_b (explicit_reference matching docs/adr/0002.md)
	// 2. decision_c -> decision_b (source_reference matching docs/adr/0002.md)
	// 3. decision_d -> decision_b (decision_id matching decision_b)
	if len(rels) != 3 {
		t.Fatalf("expected 3 relationships, got %d", len(rels))
	}

	// Assertions for decision_a -> decision_b (explicit_reference)
	rel0 := rels[0]
	if rel0.Edge.FromNodeID != model.NodeID("repo_1", model.NodeTypeDecision, "decision_a") ||
		rel0.Edge.ToNodeID != model.NodeID("repo_1", model.NodeTypeDecision, "decision_b") {
		t.Errorf("incorrect edge 0 nodes: %s -> %s", rel0.Edge.FromNodeID, rel0.Edge.ToNodeID)
	}
	var ev0 decisionDependencyEvidence
	if err := json.Unmarshal(rel0.Evidence, &ev0); err != nil {
		t.Fatalf("failed to unmarshal evidence 0: %v", err)
	}
	if ev0.MatchType != "explicit_reference" || ev0.Value != "docs/adr/0002.md" {
		t.Errorf("unexpected evidence 0 values: %+v", ev0)
	}

	// Assertions for decision_c -> decision_b (source_reference)
	rel1 := rels[1]
	if rel1.Edge.FromNodeID != model.NodeID("repo_1", model.NodeTypeDecision, "decision_c") ||
		rel1.Edge.ToNodeID != model.NodeID("repo_1", model.NodeTypeDecision, "decision_b") {
		t.Errorf("incorrect edge 1 nodes: %s -> %s", rel1.Edge.FromNodeID, rel1.Edge.ToNodeID)
	}
	var ev1 decisionDependencyEvidence
	if err := json.Unmarshal(rel1.Evidence, &ev1); err != nil {
		t.Fatalf("failed to unmarshal evidence 1: %v", err)
	}
	if ev1.MatchType != "source_reference" || ev1.Value != "docs/adr/0002.md" {
		t.Errorf("unexpected evidence 1 values: %+v", ev1)
	}

	// Assertions for decision_d -> decision_b (decision_id)
	rel2 := rels[2]
	if rel2.Edge.FromNodeID != model.NodeID("repo_1", model.NodeTypeDecision, "decision_d") ||
		rel2.Edge.ToNodeID != model.NodeID("repo_1", model.NodeTypeDecision, "decision_b") {
		t.Errorf("incorrect edge 2 nodes: %s -> %s", rel2.Edge.FromNodeID, rel2.Edge.ToNodeID)
	}
	var ev2 decisionDependencyEvidence
	if err := json.Unmarshal(rel2.Evidence, &ev2); err != nil {
		t.Fatalf("failed to unmarshal evidence 2: %v", err)
	}
	if ev2.MatchType != "decision_id" || ev2.Value != "decision_b" {
		t.Errorf("unexpected evidence 2 values: %+v", ev2)
	}
}

func TestEngine_FactSupport(t *testing.T) {
	ctx := context.Background()

	facts := []*memorymodels.Fact{
		{ID: "fact_a", RepositoryID: "repo_1", Subject: "Auth", Predicate: "USES", Object: "Redis"},
		{ID: "fact_b", RepositoryID: "repo_1", Subject: "Redis", Predicate: "PROVIDES", Object: "Caching"},
	}

	engine := NewEngine(
		&mockDecisionReader{},
		&mockIntentReader{},
		&mockFactReader{facts: facts},
		&mockEventReader{},
		&mockRelationshipReader{},
		&mockContributorReader{},
		&mockExpertiseReader{},
		&mockSourceReader{},
	)

	rels, err := engine.Generate(ctx, "repo_1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Object of Fact A ("Redis") matches Subject of Fact B ("Redis"), so A supports B
	if len(rels) != 1 {
		t.Fatalf("expected 1 relationship, got %d", len(rels))
	}

	rel := rels[0]
	if rel.Edge.EdgeType != "FACT_SUPPORTS_FACT" {
		t.Errorf("expected type FACT_SUPPORTS_FACT, got %q", rel.Edge.EdgeType)
	}
	if rel.Edge.FromNodeID != model.NodeID("repo_1", model.NodeTypeFact, "fact_a") {
		t.Errorf("incorrect FromNodeID: %q", rel.Edge.FromNodeID)
	}
	if rel.Edge.ToNodeID != model.NodeID("repo_1", model.NodeTypeFact, "fact_b") {
		t.Errorf("incorrect ToNodeID: %q", rel.Edge.ToNodeID)
	}
	if !strings.Contains(rel.Explanation, "matches the subject") {
		t.Errorf("explanation does not match subject description: %q", rel.Explanation)
	}

	var ev factSupportEvidence
	if err := json.Unmarshal(rel.Evidence, &ev); err != nil {
		t.Fatalf("failed to unmarshal evidence: %v", err)
	}
	if ev.MatchingValue != "Redis" {
		t.Errorf("unexpected evidence values: %+v", ev)
	}
}

func TestEngine_DomainRelation(t *testing.T) {
	ctx := context.Background()

	contribs := []*models.Contributor{
		{ID: "c_1", RepositoryID: "repo_1", Name: "Jane Doe", Email: "jane@example.com"},
	}
	expertise := []*models.Expertise{
		{ID: "exp_1", RepositoryID: "repo_1", ContributorID: "c_1", Domain: "Storage", Score: 0.95},
		{ID: "exp_2", RepositoryID: "repo_1", ContributorID: "c_1", Domain: "Authentication", Score: 0.80},
	}

	engine := NewEngine(
		&mockDecisionReader{},
		&mockIntentReader{},
		&mockFactReader{},
		&mockEventReader{},
		&mockRelationshipReader{},
		&mockContributorReader{contribs: contribs},
		&mockExpertiseReader{expertise: expertise},
		&mockSourceReader{},
	)

	rels, err := engine.Generate(ctx, "repo_1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Jane Doe has active expertise in both domains "Storage" and "Authentication", so they relate.
	// We sort alphabetically, so exp_2 (Authentication) is FromNode and exp_1 (Storage) is ToNode.
	if len(rels) != 1 {
		t.Fatalf("expected 1 relationship, got %d", len(rels))
	}

	rel := rels[0]
	if rel.Edge.EdgeType != "DOMAIN_RELATES_TO_DOMAIN" {
		t.Errorf("expected type DOMAIN_RELATES_TO_DOMAIN, got %q", rel.Edge.EdgeType)
	}
	if rel.Edge.FromNodeID != model.NodeID("repo_1", model.NodeTypeExpertise, "exp_2") {
		t.Errorf("incorrect FromNodeID: %q", rel.Edge.FromNodeID)
	}
	if rel.Edge.ToNodeID != model.NodeID("repo_1", model.NodeTypeExpertise, "exp_1") {
		t.Errorf("incorrect ToNodeID: %q", rel.Edge.ToNodeID)
	}
	if !strings.Contains(rel.Explanation, "active expertise in both domains") {
		t.Errorf("explanation incorrect: %q", rel.Explanation)
	}

	var ev domainRelationEvidence
	if err := json.Unmarshal(rel.Evidence, &ev); err != nil {
		t.Fatalf("failed to unmarshal evidence: %v", err)
	}
	if ev.ContributorName != "Jane Doe" || ev.DomainA != "Authentication" || ev.DomainB != "Storage" {
		t.Errorf("unexpected evidence values: %+v", ev)
	}
}

func TestEngine_SortingAndDuplicatePrevention(t *testing.T) {
	ctx := context.Background()

	// Mix relationships
	facts := []*memorymodels.Fact{
		{ID: "fact_b", RepositoryID: "repo_1", Subject: "Redis", Predicate: "PROVIDES", Object: "Caching"},
		{ID: "fact_a", RepositoryID: "repo_1", Subject: "Auth", Predicate: "USES", Object: "Redis"},
	}

	metaA := adrMetadata{Content: "Depends on decision_b", Status: "Accepted", Path: "docs/adr/0001.md"}
	metaBytesA, _ := json.Marshal(metaA)
	sources := []*models.Source{
		{ID: "src_a", RepositoryID: "repo_1", SourceType: "adr", Reference: "docs/adr/0001.md", MetadataJSON: string(metaBytesA)},
	}
	decisions := []*memorymodels.Decision{
		{ID: "decision_a", RepositoryID: "repo_1", SourceID: "src_a", Title: "Decision A"},
		{ID: "decision_b", RepositoryID: "repo_1", SourceID: "src_b", Title: "Decision B"},
	}

	engine := NewEngine(
		&mockDecisionReader{decisions: decisions},
		&mockIntentReader{},
		&mockFactReader{facts: facts},
		&mockEventReader{},
		&mockRelationshipReader{},
		&mockContributorReader{},
		&mockExpertiseReader{},
		&mockSourceReader{sources: sources},
	)

	rels1, err := engine.Generate(ctx, "repo_1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify sorting order: EdgeType ascending (DECISION_DEPENDS_ON_DECISION before FACT_SUPPORTS_FACT)
	if len(rels1) != 2 {
		t.Fatalf("expected 2 relationships, got %d", len(rels1))
	}
	if rels1[0].Edge.EdgeType != "DECISION_DEPENDS_ON_DECISION" || rels1[1].Edge.EdgeType != "FACT_SUPPORTS_FACT" {
		t.Errorf("incorrect edge sorting order: %s, %s", rels1[0].Edge.EdgeType, rels1[1].Edge.EdgeType)
	}

	// Duplicate prevention / Idempotency check
	rels2, _ := engine.Generate(ctx, "repo_1")
	if len(rels1) != len(rels2) {
		t.Errorf("counts match failed: %d vs %d", len(rels1), len(rels2))
	}
	for i := range rels1 {
		if rels1[i].Edge.ID != rels2[i].Edge.ID {
			t.Errorf("unstable Edge ID: %q vs %q", rels1[i].Edge.ID, rels2[i].Edge.ID)
		}
	}
}

// --- Integration Tests ---

func TestEngine_Integration(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "reponerve-graph-integration-*")
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

	repoID := "repo_xxx"

	// Seed database with decision dependency, fact support, and domain expertise relations
	_, err = db.Exec("INSERT INTO repositories (id, name, path, default_branch, created_at, updated_at) VALUES (?, ?, ?, ?, datetime(), datetime())", repoID, "Repo Test", tempDir, "main")
	if err != nil {
		t.Fatalf("failed to insert repository: %v", err)
	}

	metaA := adrMetadata{
		Content: "This decision depends on decision_b.",
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

	_, err = db.Exec("INSERT INTO memory_decisions (id, repository_id, source_id, title, status, created_at) VALUES (?, ?, ?, ?, ?, datetime())", "decision_a", repoID, "src_a", "Use SQLite", "Accepted")
	if err != nil {
		t.Fatalf("failed to insert decision a: %v", err)
	}
	_, err = db.Exec("INSERT INTO memory_decisions (id, repository_id, source_id, title, status, created_at) VALUES (?, ?, ?, ?, ?, datetime())", "decision_b", repoID, "src_b", "Use WAL mode", "Accepted")
	if err != nil {
		t.Fatalf("failed to insert decision b: %v", err)
	}

	_, err = db.Exec("INSERT INTO memory_facts (id, repository_id, source_id, subject, predicate, object, created_at) VALUES (?, ?, ?, ?, ?, ?, datetime())", "fact_a", repoID, "src_a", "Database", "USES", "SQLite")
	if err != nil {
		t.Fatalf("failed to insert fact a: %v", err)
	}
	_, err = db.Exec("INSERT INTO memory_facts (id, repository_id, source_id, subject, predicate, object, created_at) VALUES (?, ?, ?, ?, ?, ?, datetime())", "fact_b", repoID, "src_b", "SQLite", "HAS", "WAL Mode")
	if err != nil {
		t.Fatalf("failed to insert fact b: %v", err)
	}

	_, err = db.Exec("INSERT INTO contributors (id, repository_id, name, email, first_seen, last_seen, commit_count) VALUES (?, ?, ?, ?, datetime(), datetime(), ?)", "c_1", repoID, "John", "john@example.com", 10)
	if err != nil {
		t.Fatalf("failed to insert contributor: %v", err)
	}
	_, err = db.Exec("INSERT INTO expertise (id, repository_id, contributor_id, domain, score, evidence_json) VALUES (?, ?, ?, ?, ?, ?)", "exp_1", repoID, "c_1", "Storage", 0.9, `{"a": 1}`)
	if err != nil {
		t.Fatalf("failed to insert expertise 1: %v", err)
	}
	_, err = db.Exec("INSERT INTO expertise (id, repository_id, contributor_id, domain, score, evidence_json) VALUES (?, ?, ?, ?, ?, ?)", "exp_2", repoID, "c_1", "Infrastructure", 0.7, `{"b": 2}`)
	if err != nil {
		t.Fatalf("failed to insert expertise 2: %v", err)
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

	engine := NewEngine(dr, ir, fr, er, rr, cr, expr, sr)

	ctx := context.Background()
	rels, err := engine.Generate(ctx, repoID)
	if err != nil {
		t.Fatalf("unexpected error generating relationships: %v", err)
	}

	// We expect 3 derived relationships:
	// 1. DECISION_DEPENDS_ON_DECISION (decision_a -> decision_b)
	// 2. FACT_SUPPORTS_FACT (fact_a -> fact_b)
	// 3. DOMAIN_RELATES_TO_DOMAIN (exp_2 -> exp_1) because Infrastructure < Storage
	if len(rels) != 3 {
		t.Fatalf("expected 3 derived relationships, got %d", len(rels))
	}

	// Check DECISION_DEPENDS_ON_DECISION
	rDec := rels[0]
	if rDec.Edge.EdgeType != "DECISION_DEPENDS_ON_DECISION" {
		t.Errorf("expected decision dependency relation, got: %s", rDec.Edge.EdgeType)
	}
	if rDec.Edge.FromNodeID != model.NodeID(repoID, model.NodeTypeDecision, "decision_a") {
		t.Errorf("incorrect FromNodeID: %q", rDec.Edge.FromNodeID)
	}

	// Check DOMAIN_RELATES_TO_DOMAIN
	rDom := rels[1]
	if rDom.Edge.EdgeType != "DOMAIN_RELATES_TO_DOMAIN" {
		t.Errorf("expected domain relation, got: %s", rDom.Edge.EdgeType)
	}
	if rDom.Edge.FromNodeID != model.NodeID(repoID, model.NodeTypeExpertise, "exp_2") { // Infrastructure
		t.Errorf("incorrect FromNodeID: %q", rDom.Edge.FromNodeID)
	}

	// Check FACT_SUPPORTS_FACT
	rFact := rels[2]
	if rFact.Edge.EdgeType != "FACT_SUPPORTS_FACT" {
		t.Errorf("expected fact support relation, got: %s", rFact.Edge.EdgeType)
	}
	if rFact.Edge.FromNodeID != model.NodeID(repoID, model.NodeTypeFact, "fact_a") {
		t.Errorf("incorrect FromNodeID: %q", rFact.Edge.FromNodeID)
	}
}
