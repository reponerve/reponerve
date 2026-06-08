package agentcontext

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	appcontext "reponerve/internal/context"
	"reponerve/internal/graph/impact"
	"reponerve/internal/graph/relationships"
	"reponerve/internal/graph/traversal"
	"reponerve/internal/intelligence/changeplan"
	"reponerve/internal/intelligence/discovery"
	"reponerve/internal/intelligence/learning"
	"reponerve/internal/intelligence/reviewers"
	memorymodels "reponerve/internal/memory/models"
	"reponerve/internal/query/storage"
	"reponerve/internal/storage/migrations"
	"reponerve/internal/storage/sqlite"
	models "reponerve/pkg/models"
)

// ─── Mock Readers ────────────────────────────────────────────────────────────

type mockDecisionReader struct {
	decisions []*memorymodels.Decision
}

func (m *mockDecisionReader) GetByID(_ context.Context, id string) (*memorymodels.Decision, error) {
	for _, d := range m.decisions {
		if d.ID == id {
			return d, nil
		}
	}
	return nil, fmt.Errorf("not found")
}
func (m *mockDecisionReader) ListByRepository(_ context.Context, _ string) ([]*memorymodels.Decision, error) {
	return m.decisions, nil
}
func (m *mockDecisionReader) ListAll(_ context.Context) ([]*memorymodels.Decision, error) {
	return nil, nil
}

type mockIntentReader struct {
	intents []*memorymodels.Intent
}

func (m *mockIntentReader) GetByID(_ context.Context, _ string) (*memorymodels.Intent, error) {
	return nil, nil
}
func (m *mockIntentReader) ListByRepository(_ context.Context, _ string) ([]*memorymodels.Intent, error) {
	return m.intents, nil
}
func (m *mockIntentReader) ListAll(_ context.Context) ([]*memorymodels.Intent, error) {
	return nil, nil
}

type mockFactReader struct {
	facts []*memorymodels.Fact
}

func (m *mockFactReader) GetByID(_ context.Context, id string) (*memorymodels.Fact, error) {
	for _, f := range m.facts {
		if f.ID == id {
			return f, nil
		}
	}
	return nil, fmt.Errorf("not found")
}
func (m *mockFactReader) ListByRepository(_ context.Context, _ string) ([]*memorymodels.Fact, error) {
	return m.facts, nil
}
func (m *mockFactReader) ListAll(_ context.Context) ([]*memorymodels.Fact, error) { return nil, nil }

type mockEventReader struct {
	events []*models.Event
}

func (m *mockEventReader) GetByID(_ context.Context, id string) (*models.Event, error) {
	for _, ev := range m.events {
		if ev.ID == id {
			return ev, nil
		}
	}
	return nil, fmt.Errorf("not found")
}
func (m *mockEventReader) ListByRepository(_ context.Context, _ string) ([]*models.Event, error) {
	return m.events, nil
}
func (m *mockEventReader) ListAll(_ context.Context) ([]*models.Event, error) { return nil, nil }

type mockRelationshipReader struct {
	rels []*memorymodels.Relationship
}

func (m *mockRelationshipReader) GetByID(_ context.Context, _ string) (*memorymodels.Relationship, error) {
	return nil, nil
}
func (m *mockRelationshipReader) ListByRepository(_ context.Context, _ string) ([]*memorymodels.Relationship, error) {
	return m.rels, nil
}
func (m *mockRelationshipReader) ListAll(_ context.Context) ([]*memorymodels.Relationship, error) {
	return nil, nil
}

type mockContributorReader struct {
	contribs []*models.Contributor
}

func (m *mockContributorReader) GetByID(_ context.Context, _, _ string) (*models.Contributor, error) {
	return nil, nil
}
func (m *mockContributorReader) ListByRepository(_ context.Context, _ string) ([]*models.Contributor, error) {
	return m.contribs, nil
}

type mockExpertiseReader struct {
	expertise []*models.Expertise
}

func (m *mockExpertiseReader) ListByRepository(_ context.Context, _ string) ([]*models.Expertise, error) {
	return m.expertise, nil
}
func (m *mockExpertiseReader) ListByContributor(_ context.Context, _, _ string) ([]*models.Expertise, error) {
	return nil, nil
}

type mockSourceReader struct {
	sources []*models.Source
}

func (m *mockSourceReader) ListByRepository(_ context.Context, _ string) ([]*models.Source, error) {
	return m.sources, nil
}

// ─── Mock context reader ──────────────────────────────────────────────────────

type mockContextReader struct {
	decisions []*memorymodels.Decision
	intents   []*memorymodels.Intent
	facts     []*memorymodels.Fact
	events    []*models.Event
}

func (m *mockContextReader) ReadContext(_ context.Context, repoID string) (*appcontext.ContextData, error) {
	return &appcontext.ContextData{
		RepositoryID: repoID,
		Decisions:    m.decisions,
		Intents:      m.intents,
		Facts:        m.facts,
		Events:       m.events,
	}, nil
}

// ─── Test helpers ─────────────────────────────────────────────────────────────

// buildFullServiceStack builds the complete intelligence stack from mock readers.
func buildFullServiceStack(
	dr storage.DecisionReader,
	ir storage.IntentReader,
	fr storage.FactReader,
	er storage.EventReader,
	rr storage.RelationshipReader,
	cr storage.ContributorReader,
	expr storage.ExpertiseReader,
	sr storage.SourceReader,
	decisions []*memorymodels.Decision,
	intents []*memorymodels.Intent,
	facts []*memorymodels.Fact,
	events []*models.Event,
) *Service {
	relEngine := relationships.NewEngine(dr, ir, fr, er, rr, cr, expr, sr)
	travEngine := traversal.NewEngine(relEngine)
	impactSvc := impact.NewService(travEngine)

	discoverySvc := discovery.NewService(dr, fr, er, cr, expr, rr, relEngine, travEngine, impactSvc)
	learningSvc := learning.NewService(discoverySvc, dr, fr, er, cr, expr, sr, relEngine)
	reviewerSvc := reviewers.NewService(discoverySvc, dr, fr, er, cr, expr, sr, impactSvc)
	changePlanSvc := changeplan.NewService(impactSvc)

	ctxReader := &mockContextReader{
		decisions: decisions,
		intents:   intents,
		facts:     facts,
		events:    events,
	}
	ctxGenerator := appcontext.NewGenerator(ctxReader)

	return NewService(discoverySvc, learningSvc, reviewerSvc, changePlanSvc, ctxGenerator)
}

// emptyService returns a Service wired with no repository data.
func emptyService() *Service {
	return buildFullServiceStack(
		&mockDecisionReader{},
		&mockIntentReader{},
		&mockFactReader{},
		&mockEventReader{},
		&mockRelationshipReader{},
		&mockContributorReader{},
		&mockExpertiseReader{},
		&mockSourceReader{},
		nil, nil, nil, nil,
	)
}

// ─── ValidatePackage Unit Tests ───────────────────────────────────────────────

func TestValidatePackage_Nil(t *testing.T) {
	if err := ValidatePackage(nil); err == nil {
		t.Error("expected error for nil package")
	}
}

func TestValidatePackage_EmptyRepositoryID(t *testing.T) {
	pkg := &AgentContextPackage{
		RepositoryID: "",
		Sections:     []*ContextSection{{Name: "x", Source: SourceContext, Data: json.RawMessage(`{}`)}},
	}
	if err := ValidatePackage(pkg); err == nil {
		t.Error("expected error for empty RepositoryID")
	}
}

func TestValidatePackage_NoSections(t *testing.T) {
	pkg := &AgentContextPackage{
		RepositoryID: "repo_1",
		Sections:     []*ContextSection{},
	}
	if err := ValidatePackage(pkg); err == nil {
		t.Error("expected error for no sections")
	}
}

func TestValidatePackage_NilSection(t *testing.T) {
	pkg := &AgentContextPackage{
		RepositoryID: "repo_1",
		Sections:     []*ContextSection{nil},
	}
	if err := ValidatePackage(pkg); err == nil {
		t.Error("expected error for nil section")
	}
}

func TestValidatePackage_EmptySectionName(t *testing.T) {
	pkg := &AgentContextPackage{
		RepositoryID: "repo_1",
		Sections: []*ContextSection{
			{Name: "", Source: SourceContext, Data: json.RawMessage(`{}`)},
		},
	}
	if err := ValidatePackage(pkg); err == nil {
		t.Error("expected error for empty section name")
	}
}

func TestValidatePackage_EmptySource(t *testing.T) {
	pkg := &AgentContextPackage{
		RepositoryID: "repo_1",
		Sections: []*ContextSection{
			{Name: "Overview", Source: "", Data: json.RawMessage(`{}`)},
		},
	}
	if err := ValidatePackage(pkg); err == nil {
		t.Error("expected error for empty source")
	}
}

func TestValidatePackage_InvalidSource(t *testing.T) {
	pkg := &AgentContextPackage{
		RepositoryID: "repo_1",
		Sections: []*ContextSection{
			{Name: "Overview", Source: "unknown_source", Data: json.RawMessage(`{}`)},
		},
	}
	if err := ValidatePackage(pkg); err == nil {
		t.Error("expected error for unsupported source value")
	}
}

func TestValidatePackage_EmptyData(t *testing.T) {
	pkg := &AgentContextPackage{
		RepositoryID: "repo_1",
		Sections: []*ContextSection{
			{Name: "Overview", Source: SourceContext, Data: json.RawMessage{}},
		},
	}
	if err := ValidatePackage(pkg); err == nil {
		t.Error("expected error for empty section data")
	}
}

func TestValidatePackage_Valid(t *testing.T) {
	pkg := &AgentContextPackage{
		RepositoryID: "repo_1",
		Sections: []*ContextSection{
			{Name: "Overview", Source: SourceContext, Data: json.RawMessage(`{"repository_id":"repo_1"}`)},
		},
	}
	if err := ValidatePackage(pkg); err != nil {
		t.Errorf("expected valid package, got: %v", err)
	}
}

// ─── BuildRepositoryContext Unit Tests ────────────────────────────────────────

func TestBuildRepositoryContext_EmptyRepositoryID(t *testing.T) {
	svc := emptyService()
	_, err := svc.BuildRepositoryContext(context.Background(), "")
	if err == nil {
		t.Error("expected error for empty repository ID")
	}
}

func TestBuildRepositoryContext_Empty(t *testing.T) {
	ctx := context.Background()
	svc := emptyService()

	pkg, err := svc.BuildRepositoryContext(ctx, "repo_1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pkg.RepositoryID != "repo_1" {
		t.Errorf("expected RepositoryID %q, got %q", "repo_1", pkg.RepositoryID)
	}
	if len(pkg.Sections) != 4 {
		t.Fatalf("expected 4 sections, got %d", len(pkg.Sections))
	}
}

func TestBuildRepositoryContext_SectionOrdering(t *testing.T) {
	ctx := context.Background()
	svc := emptyService()

	pkg, err := svc.BuildRepositoryContext(ctx, "repo_1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedNames := []string{"Repository Overview", "Discovery", "Learning Path", "Reviewer Recommendations"}
	expectedSources := []string{SourceContext, SourceDiscovery, SourceLearning, SourceReviewers}

	for i, section := range pkg.Sections {
		if section.Name != expectedNames[i] {
			t.Errorf("section %d: expected name %q, got %q", i, expectedNames[i], section.Name)
		}
		if section.Source != expectedSources[i] {
			t.Errorf("section %d: expected source %q, got %q", i, expectedSources[i], section.Source)
		}
		if len(section.Data) == 0 {
			t.Errorf("section %d (%q): data is empty", i, section.Name)
		}
	}
}

func TestBuildRepositoryContext_SectionDataIsValidJSON(t *testing.T) {
	ctx := context.Background()
	svc := emptyService()

	pkg, err := svc.BuildRepositoryContext(ctx, "repo_1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for i, section := range pkg.Sections {
		if !json.Valid(section.Data) {
			t.Errorf("section %d (%q): Data is not valid JSON", i, section.Name)
		}
	}
}

func TestBuildRepositoryContext_Deterministic(t *testing.T) {
	ctx := context.Background()
	svc := emptyService()

	pkg1, err := svc.BuildRepositoryContext(ctx, "repo_det")
	if err != nil {
		t.Fatalf("first call error: %v", err)
	}
	pkg2, err := svc.BuildRepositoryContext(ctx, "repo_det")
	if err != nil {
		t.Fatalf("second call error: %v", err)
	}

	if len(pkg1.Sections) != len(pkg2.Sections) {
		t.Fatalf("section count mismatch: %d vs %d", len(pkg1.Sections), len(pkg2.Sections))
	}
	for i := range pkg1.Sections {
		if pkg1.Sections[i].Name != pkg2.Sections[i].Name {
			t.Errorf("section %d name mismatch", i)
		}
		if pkg1.Sections[i].Source != pkg2.Sections[i].Source {
			t.Errorf("section %d source mismatch", i)
		}
	}
}

// ─── BuildDomainContext Unit Tests ────────────────────────────────────────────

func TestBuildDomainContext_EmptyRepositoryID(t *testing.T) {
	svc := emptyService()
	_, err := svc.BuildDomainContext(context.Background(), "", "Storage")
	if err == nil {
		t.Error("expected error for empty repository ID")
	}
}

func TestBuildDomainContext_EmptyDomain(t *testing.T) {
	svc := emptyService()
	_, err := svc.BuildDomainContext(context.Background(), "repo_1", "")
	if err == nil {
		t.Error("expected error for empty domain")
	}
}

func TestBuildDomainContext_SectionOrdering(t *testing.T) {
	ctx := context.Background()
	svc := emptyService()

	pkg, err := svc.BuildDomainContext(ctx, "repo_1", "Storage")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pkg.Sections) != 3 {
		t.Fatalf("expected 3 sections, got %d", len(pkg.Sections))
	}

	expectedNames := []string{"Domain Overview", "Learning Path", "Reviewer Recommendations"}
	expectedSources := []string{SourceContext, SourceLearning, SourceReviewers}

	for i, section := range pkg.Sections {
		if section.Name != expectedNames[i] {
			t.Errorf("section %d: expected name %q, got %q", i, expectedNames[i], section.Name)
		}
		if section.Source != expectedSources[i] {
			t.Errorf("section %d: expected source %q, got %q", i, expectedSources[i], section.Source)
		}
		if len(section.Data) == 0 {
			t.Errorf("section %d (%q): data is empty", i, section.Name)
		}
	}
}

func TestBuildDomainContext_SectionDataIsValidJSON(t *testing.T) {
	ctx := context.Background()
	svc := emptyService()

	pkg, err := svc.BuildDomainContext(ctx, "repo_1", "Storage")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for i, section := range pkg.Sections {
		if !json.Valid(section.Data) {
			t.Errorf("section %d (%q): Data is not valid JSON", i, section.Name)
		}
	}
}

// ─── BuildContributorContext Unit Tests ───────────────────────────────────────

func TestBuildContributorContext_EmptyRepositoryID(t *testing.T) {
	svc := emptyService()
	_, err := svc.BuildContributorContext(context.Background(), "", "ctr_abc")
	if err == nil {
		t.Error("expected error for empty repository ID")
	}
}

func TestBuildContributorContext_EmptyContributorID(t *testing.T) {
	svc := emptyService()
	_, err := svc.BuildContributorContext(context.Background(), "repo_1", "")
	if err == nil {
		t.Error("expected error for empty contributor ID")
	}
}

func TestBuildContributorContext_SectionOrdering(t *testing.T) {
	ctx := context.Background()
	svc := emptyService()

	pkg, err := svc.BuildContributorContext(ctx, "repo_1", "ctr_abc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pkg.Sections) != 3 {
		t.Fatalf("expected 3 sections, got %d", len(pkg.Sections))
	}

	expectedNames := []string{"Contributor Overview", "Learning Path", "Change Plan"}
	expectedSources := []string{SourceContext, SourceLearning, SourceChangePlan}

	for i, section := range pkg.Sections {
		if section.Name != expectedNames[i] {
			t.Errorf("section %d: expected name %q, got %q", i, expectedNames[i], section.Name)
		}
		if section.Source != expectedSources[i] {
			t.Errorf("section %d: expected source %q, got %q", i, expectedSources[i], section.Source)
		}
		if len(section.Data) == 0 {
			t.Errorf("section %d (%q): data is empty", i, section.Name)
		}
	}
}

func TestBuildContributorContext_SectionDataIsValidJSON(t *testing.T) {
	ctx := context.Background()
	svc := emptyService()

	pkg, err := svc.BuildContributorContext(ctx, "repo_1", "ctr_abc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for i, section := range pkg.Sections {
		if !json.Valid(section.Data) {
			t.Errorf("section %d (%q): Data is not valid JSON", i, section.Name)
		}
	}
}

// ─── Evidence Preservation Tests ─────────────────────────────────────────────

func TestBuildRepositoryContext_EvidencePreserved(t *testing.T) {
	ctx := context.Background()
	repoID := "repo_ev"

	decisions := []*memorymodels.Decision{
		{ID: "dec_1", RepositoryID: repoID, Title: "Use SQLite", Status: "Accepted", CreatedAt: time.Now()},
	}
	facts := []*memorymodels.Fact{
		{ID: "fact_1", RepositoryID: repoID, Subject: "db", Predicate: "is", Object: "WAL", CreatedAt: time.Now()},
	}

	svc := buildFullServiceStack(
		&mockDecisionReader{decisions: decisions},
		&mockIntentReader{},
		&mockFactReader{facts: facts},
		&mockEventReader{},
		&mockRelationshipReader{},
		&mockContributorReader{},
		&mockExpertiseReader{},
		&mockSourceReader{},
		decisions, nil, facts, nil,
	)

	pkg, err := svc.BuildRepositoryContext(ctx, repoID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Find the discovery section and unmarshal
	var discoverySection *ContextSection
	for _, s := range pkg.Sections {
		if s.Source == SourceDiscovery {
			discoverySection = s
			break
		}
	}
	if discoverySection == nil {
		t.Fatal("discovery section not found")
	}

	var report discovery.KnowledgeDiscoveryReport
	if err := json.Unmarshal(discoverySection.Data, &report); err != nil {
		t.Fatalf("failed to unmarshal discovery section: %v", err)
	}

	// Verify evidence fields are present on each item
	for _, item := range report.Items {
		if item.EvidenceJSON == "" {
			t.Errorf("item %q/%q: evidence_json is empty", item.EntityType, item.EntityID)
		}
		if item.Explanation == "" {
			t.Errorf("item %q/%q: explanation is empty", item.EntityType, item.EntityID)
		}
	}
}

func TestBuildRepositoryContext_LearningEvidencePreserved(t *testing.T) {
	ctx := context.Background()
	repoID := "repo_lev"

	decisions := []*memorymodels.Decision{
		{ID: "dec_1", RepositoryID: repoID, Title: "Use SQLite", Status: "Accepted", CreatedAt: time.Now()},
	}

	svc := buildFullServiceStack(
		&mockDecisionReader{decisions: decisions},
		&mockIntentReader{},
		&mockFactReader{},
		&mockEventReader{},
		&mockRelationshipReader{},
		&mockContributorReader{},
		&mockExpertiseReader{},
		&mockSourceReader{},
		decisions, nil, nil, nil,
	)

	pkg, err := svc.BuildRepositoryContext(ctx, repoID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Find learning section and check evidence preservation
	for _, s := range pkg.Sections {
		if s.Source != SourceLearning {
			continue
		}
		var path learning.LearningPath
		if err := json.Unmarshal(s.Data, &path); err != nil {
			t.Fatalf("failed to unmarshal learning section: %v", err)
		}
		for _, step := range path.Steps {
			if step.EvidenceJSON == "" {
				t.Errorf("step %d: evidence_json is empty", step.Position)
			}
			if step.Explanation == "" {
				t.Errorf("step %d: explanation is empty", step.Position)
			}
			if step.Position <= 0 {
				t.Errorf("step: position must be > 0, got %d", step.Position)
			}
		}
	}
}

// ─── Integration Test ─────────────────────────────────────────────────────────

func TestService_Integration(t *testing.T) {
	ctx := context.Background()

	tempDir, err := os.MkdirTemp("", "reponerve-agentctx-integration-*")
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

	repoID := "repo_agentctx"

	// Seed repository
	_, err = db.Exec(
		"INSERT INTO repositories (id, name, path, default_branch, created_at, updated_at) VALUES (?, ?, ?, ?, datetime(), datetime())",
		repoID, "Agent Context Repo", tempDir, "main",
	)
	if err != nil {
		t.Fatalf("failed to insert repository: %v", err)
	}

	// Seed source
	_, err = db.Exec(
		"INSERT INTO sources (id, repository_id, source_type, reference, title, author, timestamp, created_at) VALUES (?, ?, ?, ?, ?, ?, datetime(), datetime())",
		"src_1", repoID, "adr", "docs/adr/0001.md", "Use SQLite", "Alice <alice@example.com>",
	)
	if err != nil {
		t.Fatalf("failed to insert source: %v", err)
	}

	// Seed decisions
	_, err = db.Exec(
		"INSERT INTO memory_decisions (id, repository_id, source_id, title, status, created_at) VALUES (?, ?, ?, ?, ?, datetime())",
		"dec_1", repoID, "src_1", "Use SQLite as primary store", "Accepted",
	)
	if err != nil {
		t.Fatalf("failed to insert decision: %v", err)
	}

	_, err = db.Exec(
		"INSERT INTO memory_decisions (id, repository_id, source_id, title, status, created_at) VALUES (?, ?, ?, ?, ?, datetime())",
		"dec_2", repoID, "src_1", "Enable WAL mode", "Accepted",
	)
	if err != nil {
		t.Fatalf("failed to insert decision: %v", err)
	}

	// Seed relationship
	_, err = db.Exec(
		"INSERT INTO memory_relationships (id, repository_id, from_id, to_id, relationship_type, created_at) VALUES (?, ?, ?, ?, ?, datetime())",
		"rel_1", repoID, "dec_1", "dec_2", "DECISION_DEPENDS_ON",
	)
	if err != nil {
		t.Fatalf("failed to insert relationship: %v", err)
	}

	// Seed contributor + expertise
	_, err = db.Exec(
		"INSERT INTO contributors (id, repository_id, name, email, first_seen, last_seen, commit_count) VALUES (?, ?, ?, ?, datetime(), datetime(), ?)",
		"ctr_alice", repoID, "Alice", "alice@example.com", 10,
	)
	if err != nil {
		t.Fatalf("failed to insert contributor: %v", err)
	}

	_, err = db.Exec(
		"INSERT INTO expertise (id, repository_id, contributor_id, domain, score, evidence_json) VALUES (?, ?, ?, ?, ?, ?)",
		"exp_1", repoID, "ctr_alice", "Storage", 0.9, `{"explanation":"expert contributor"}`,
	)
	if err != nil {
		t.Fatalf("failed to insert expertise: %v", err)
	}

	// Build the full service stack from real SQLite readers
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
	learningSvc := learning.NewService(discoverySvc, dr, fr, er, cr, expr, sr, relEngine)
	reviewerSvc := reviewers.NewService(discoverySvc, dr, fr, er, cr, expr, sr, impactSvc)
	changePlanSvc := changeplan.NewService(impactSvc)

	ctxReader := appcontext.NewMemoryContextReader(er, dr, ir, fr)
	ctxGenerator := appcontext.NewGenerator(ctxReader)

	svc := NewService(discoverySvc, learningSvc, reviewerSvc, changePlanSvc, ctxGenerator)

	t.Run("BuildRepositoryContext", func(t *testing.T) {
		pkg, err := svc.BuildRepositoryContext(ctx, repoID)
		if err != nil {
			t.Fatalf("BuildRepositoryContext failed: %v", err)
		}
		if pkg.RepositoryID != repoID {
			t.Errorf("expected RepositoryID %q, got %q", repoID, pkg.RepositoryID)
		}
		if len(pkg.Sections) != 4 {
			t.Fatalf("expected 4 sections, got %d", len(pkg.Sections))
		}

		expectedNames := []string{"Repository Overview", "Discovery", "Learning Path", "Reviewer Recommendations"}
		for i, s := range pkg.Sections {
			if s.Name != expectedNames[i] {
				t.Errorf("section %d: expected %q, got %q", i, expectedNames[i], s.Name)
			}
			if !json.Valid(s.Data) {
				t.Errorf("section %d (%q): invalid JSON in Data", i, s.Name)
			}
		}

		// Validate package structure
		if err := ValidatePackage(pkg); err != nil {
			t.Errorf("ValidatePackage failed: %v", err)
		}

		// Verify discovery evidence is preserved
		var report discovery.KnowledgeDiscoveryReport
		if err := json.Unmarshal(pkg.Sections[1].Data, &report); err != nil {
			t.Fatalf("failed to unmarshal discovery: %v", err)
		}
		for _, item := range report.Items {
			if item.EvidenceJSON == "" {
				t.Errorf("discovery item missing evidence_json")
			}
			if item.Explanation == "" {
				t.Errorf("discovery item missing explanation")
			}
		}
	})

	t.Run("BuildRepositoryContext_Determinism", func(t *testing.T) {
		pkg1, err := svc.BuildRepositoryContext(ctx, repoID)
		if err != nil {
			t.Fatalf("first call failed: %v", err)
		}
		pkg2, err := svc.BuildRepositoryContext(ctx, repoID)
		if err != nil {
			t.Fatalf("second call failed: %v", err)
		}
		for i := range pkg1.Sections {
			if pkg1.Sections[i].Name != pkg2.Sections[i].Name {
				t.Errorf("section %d name mismatch on second call", i)
			}
			if pkg1.Sections[i].Source != pkg2.Sections[i].Source {
				t.Errorf("section %d source mismatch on second call", i)
			}
		}
	})

	t.Run("BuildDomainContext", func(t *testing.T) {
		pkg, err := svc.BuildDomainContext(ctx, repoID, "Storage")
		if err != nil {
			t.Fatalf("BuildDomainContext failed: %v", err)
		}
		if len(pkg.Sections) != 3 {
			t.Fatalf("expected 3 sections, got %d", len(pkg.Sections))
		}

		expectedNames := []string{"Domain Overview", "Learning Path", "Reviewer Recommendations"}
		for i, s := range pkg.Sections {
			if s.Name != expectedNames[i] {
				t.Errorf("section %d: expected %q, got %q", i, expectedNames[i], s.Name)
			}
		}

		if err := ValidatePackage(pkg); err != nil {
			t.Errorf("ValidatePackage failed: %v", err)
		}
	})

	t.Run("BuildContributorContext", func(t *testing.T) {
		pkg, err := svc.BuildContributorContext(ctx, repoID, "ctr_alice")
		if err != nil {
			t.Fatalf("BuildContributorContext failed: %v", err)
		}
		if len(pkg.Sections) != 3 {
			t.Fatalf("expected 3 sections, got %d", len(pkg.Sections))
		}

		expectedNames := []string{"Contributor Overview", "Learning Path", "Change Plan"}
		expectedSources := []string{SourceContext, SourceLearning, SourceChangePlan}
		for i, s := range pkg.Sections {
			if s.Name != expectedNames[i] {
				t.Errorf("section %d: expected name %q, got %q", i, expectedNames[i], s.Name)
			}
			if s.Source != expectedSources[i] {
				t.Errorf("section %d: expected source %q, got %q", i, expectedSources[i], s.Source)
			}
		}

		if err := ValidatePackage(pkg); err != nil {
			t.Errorf("ValidatePackage failed: %v", err)
		}

		// Verify change plan evidence preserved
		var plan changeplan.ChangePlan
		if err := json.Unmarshal(pkg.Sections[2].Data, &plan); err != nil {
			t.Fatalf("failed to unmarshal change plan: %v", err)
		}
		for _, item := range plan.Items {
			if item.EvidenceJSON == "" {
				t.Errorf("change plan item missing evidence_json")
			}
			if item.Explanation == "" {
				t.Errorf("change plan item missing explanation")
			}
		}
	})
}
