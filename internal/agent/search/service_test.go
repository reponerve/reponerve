package agentsearch

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"reponerve/internal/graph/impact"
	"reponerve/internal/graph/relationships"
	"reponerve/internal/graph/traversal"
	"reponerve/internal/intelligence/discovery"
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
func (m *mockDecisionReader) ListAll(_ context.Context) ([]*memorymodels.Decision, error) { return nil, nil }

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

func (m *mockContributorReader) GetByID(_ context.Context, _, id string) (*models.Contributor, error) {
	for _, c := range m.contribs {
		if c.ID == id {
			return c, nil
		}
	}
	return nil, fmt.Errorf("not found")
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

type mockIntentReader struct{}

func (m *mockIntentReader) GetByID(_ context.Context, _ string) (*memorymodels.Intent, error) {
	return nil, nil
}
func (m *mockIntentReader) ListByRepository(_ context.Context, _ string) ([]*memorymodels.Intent, error) {
	return nil, nil
}
func (m *mockIntentReader) ListAll(_ context.Context) ([]*memorymodels.Intent, error) { return nil, nil }

type mockSourceReader struct{}

func (m *mockSourceReader) ListByRepository(_ context.Context, _ string) ([]*models.Source, error) {
	return nil, nil
}

// ─── Test helpers ─────────────────────────────────────────────────────────────

func buildSearchService(
	dr storage.DecisionReader,
	fr storage.FactReader,
	er storage.EventReader,
	rr storage.RelationshipReader,
	cr storage.ContributorReader,
	expr storage.ExpertiseReader,
) *Service {
	relEngine := relationships.NewEngine(
		dr,
		&mockIntentReader{},
		fr,
		er,
		rr,
		cr,
		expr,
		&mockSourceReader{},
	)
	travEngine := traversal.NewEngine(relEngine)
	impactSvc := impact.NewService(travEngine)
	discoverySvc := discovery.NewService(dr, fr, er, cr, expr, rr, relEngine, travEngine, impactSvc)

	return NewService(dr, fr, er, rr, cr, expr, discoverySvc)
}

func emptyService() *Service {
	return buildSearchService(
		&mockDecisionReader{},
		&mockFactReader{},
		&mockEventReader{},
		&mockRelationshipReader{},
		&mockContributorReader{},
		&mockExpertiseReader{},
	)
}

func parseEvidence(t *testing.T, hit *SearchHit) matchEvidence {
	t.Helper()
	var ev matchEvidence
	if err := json.Unmarshal([]byte(hit.EvidenceJSON), &ev); err != nil {
		t.Fatalf("failed to parse evidence: %v", err)
	}
	return ev
}

// ─── ValidateResult Unit Tests ────────────────────────────────────────────────

func TestValidateResult_Nil(t *testing.T) {
	if err := ValidateResult(nil); err == nil {
		t.Error("expected error for nil result")
	}
}

func TestValidateResult_EmptyRepositoryID(t *testing.T) {
	r := &SearchResult{RepositoryID: "", Query: "redis", Hits: []*SearchHit{}}
	if err := ValidateResult(r); err == nil {
		t.Error("expected error for empty repository ID")
	}
}

func TestValidateResult_EmptyQuery(t *testing.T) {
	r := &SearchResult{RepositoryID: "repo_1", Query: "", Hits: []*SearchHit{}}
	if err := ValidateResult(r); err == nil {
		t.Error("expected error for empty query")
	}
}

func TestValidateResult_NilHits(t *testing.T) {
	r := &SearchResult{RepositoryID: "repo_1", Query: "redis", Hits: nil}
	if err := ValidateResult(r); err == nil {
		t.Error("expected error for nil hits")
	}
}

func TestValidateResult_NilHit(t *testing.T) {
	r := &SearchResult{RepositoryID: "repo_1", Query: "redis", Hits: []*SearchHit{nil}}
	if err := ValidateResult(r); err == nil {
		t.Error("expected error for nil hit")
	}
}

func TestValidateResult_InvalidSource(t *testing.T) {
	r := &SearchResult{
		RepositoryID: "repo_1",
		Query:        "redis",
		Hits: []*SearchHit{{
			EntityType: EntityTypeDecision, EntityID: "dec_1",
			Source: "unknown", MatchScore: 100,
			EvidenceJSON: `{"match_type":"exact","field":"title"}`,
		}},
	}
	if err := ValidateResult(r); err == nil {
		t.Error("expected error for unsupported source")
	}
}

func TestValidateResult_NegativeScore(t *testing.T) {
	r := &SearchResult{
		RepositoryID: "repo_1",
		Query:        "redis",
		Hits: []*SearchHit{{
			EntityType: EntityTypeDecision, EntityID: "dec_1",
			Source: SourceMemory, MatchScore: -1,
			EvidenceJSON: `{"match_type":"exact","field":"title"}`,
		}},
	}
	if err := ValidateResult(r); err == nil {
		t.Error("expected error for negative match score")
	}
}

func TestValidateResult_ValidEmptyHits(t *testing.T) {
	r := &SearchResult{RepositoryID: "repo_1", Query: "redis", Hits: []*SearchHit{}}
	if err := ValidateResult(r); err != nil {
		t.Errorf("expected valid result, got: %v", err)
	}
}

// ─── Query Parsing Tests ──────────────────────────────────────────────────────

func TestSearch_UnknownPrefix(t *testing.T) {
	svc := emptyService()
	_, err := svc.Search(context.Background(), "repo_1", "owner:alice")
	if err == nil {
		t.Error("expected error for unknown prefix owner")
	}
}

func TestSearch_UnknownPrefixSeverity(t *testing.T) {
	svc := emptyService()
	_, err := svc.Search(context.Background(), "repo_1", "severity:high")
	if err == nil {
		t.Error("expected error for unknown prefix severity")
	}
}

func TestSearch_EmptyRepositoryID(t *testing.T) {
	svc := emptyService()
	_, err := svc.Search(context.Background(), "", "redis")
	if err == nil {
		t.Error("expected error for empty repository ID")
	}
}

func TestSearch_EmptyQuery(t *testing.T) {
	svc := emptyService()
	_, err := svc.Search(context.Background(), "repo_1", "")
	if err == nil {
		t.Error("expected error for empty query")
	}
}

// ─── Empty Repository ─────────────────────────────────────────────────────────

func TestSearch_EmptyRepository(t *testing.T) {
	svc := emptyService()
	result, err := svc.Search(context.Background(), "repo_1", "redis")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Hits) != 0 {
		t.Errorf("expected 0 hits, got %d", len(result.Hits))
	}
	if err := ValidateResult(result); err != nil {
		t.Errorf("ValidateResult failed: %v", err)
	}
}

// ─── Exact / Prefix / Partial Matches ─────────────────────────────────────────

func TestSearch_ExactMatch(t *testing.T) {
	repoID := "repo_1"
	svc := buildSearchService(
		&mockDecisionReader{decisions: []*memorymodels.Decision{
			{ID: "dec_1", RepositoryID: repoID, Title: "Use Redis for caching", Status: "Accepted", CreatedAt: time.Now()},
		}},
		&mockFactReader{}, &mockEventReader{}, &mockRelationshipReader{},
		&mockContributorReader{}, &mockExpertiseReader{},
	)

	result, err := svc.Search(context.Background(), repoID, "dec_1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Hits) == 0 {
		t.Fatal("expected at least one hit")
	}

	var hit *SearchHit
	for _, h := range result.Hits {
		if h.EntityID == "dec_1" && h.EntityType == EntityTypeDecision {
			hit = h
			break
		}
	}
	if hit == nil {
		t.Fatal("expected decision hit for dec_1")
	}
	if hit.MatchScore != ScoreExact {
		t.Errorf("expected exact match score %d, got %d", ScoreExact, hit.MatchScore)
	}
	if hit.Source != SourceMemory {
		t.Errorf("expected source %q, got %q", SourceMemory, hit.Source)
	}
	ev := parseEvidence(t, hit)
	if ev.MatchType != "exact" || ev.Field != "id" {
		t.Errorf("unexpected evidence: %+v", ev)
	}
}

func TestSearch_PrefixMatch(t *testing.T) {
	repoID := "repo_1"
	svc := buildSearchService(
		&mockDecisionReader{decisions: []*memorymodels.Decision{
			{ID: "dec_1", RepositoryID: repoID, Title: "Authentication via OIDC", Status: "Accepted", CreatedAt: time.Now()},
		}},
		&mockFactReader{}, &mockEventReader{}, &mockRelationshipReader{},
		&mockContributorReader{}, &mockExpertiseReader{},
	)

	result, err := svc.Search(context.Background(), repoID, "auth*")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Hits) == 0 {
		t.Fatal("expected at least one hit")
	}
	if result.Hits[0].MatchScore != ScorePrefix {
		t.Errorf("expected prefix match score %d, got %d", ScorePrefix, result.Hits[0].MatchScore)
	}
	ev := parseEvidence(t, result.Hits[0])
	if ev.MatchType != "prefix" {
		t.Errorf("expected prefix match type, got %q", ev.MatchType)
	}
}

func TestSearch_PartialMatch(t *testing.T) {
	repoID := "repo_1"
	svc := buildSearchService(
		&mockDecisionReader{},
		&mockFactReader{facts: []*memorymodels.Fact{
			{ID: "fact_1", RepositoryID: repoID, Subject: "cache", Predicate: "uses", Object: "redis", CreatedAt: time.Now()},
		}},
		&mockEventReader{}, &mockRelationshipReader{},
		&mockContributorReader{}, &mockExpertiseReader{},
	)

	result, err := svc.Search(context.Background(), repoID, "edis")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Hits) == 0 {
		t.Fatal("expected at least one hit")
	}
	if result.Hits[0].MatchScore != ScorePartial {
		t.Errorf("expected partial match score %d, got %d", ScorePartial, result.Hits[0].MatchScore)
	}
	ev := parseEvidence(t, result.Hits[0])
	if ev.MatchType != "partial" {
		t.Errorf("expected partial match type, got %q", ev.MatchType)
	}
}

// ─── Structured Search ────────────────────────────────────────────────────────

func TestSearch_StructuredTypeDecision(t *testing.T) {
	repoID := "repo_1"
	svc := buildSearchService(
		&mockDecisionReader{decisions: []*memorymodels.Decision{
			{ID: "dec_1", RepositoryID: repoID, Title: "Use Redis", Status: "Accepted", CreatedAt: time.Now()},
			{ID: "dec_2", RepositoryID: repoID, Title: "Use SQLite", Status: "Accepted", CreatedAt: time.Now()},
		}},
		&mockFactReader{facts: []*memorymodels.Fact{
			{ID: "fact_1", RepositoryID: repoID, Subject: "redis", Predicate: "is", Object: "cache", CreatedAt: time.Now()},
		}},
		&mockEventReader{}, &mockRelationshipReader{},
		&mockContributorReader{}, &mockExpertiseReader{},
	)

	result, err := svc.Search(context.Background(), repoID, "type:decision redis")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, hit := range result.Hits {
		if hit.EntityType != EntityTypeDecision {
			t.Errorf("expected only DECISION hits, got %q", hit.EntityType)
		}
	}
	if len(result.Hits) == 0 {
		t.Error("expected at least one decision hit")
	}
}

func TestSearch_StructuredTypeFact(t *testing.T) {
	repoID := "repo_1"
	svc := buildSearchService(
		&mockDecisionReader{},
		&mockFactReader{facts: []*memorymodels.Fact{
			{ID: "fact_1", RepositoryID: repoID, Subject: "auth", Predicate: "uses", Object: "oidc", CreatedAt: time.Now()},
		}},
		&mockEventReader{}, &mockRelationshipReader{},
		&mockContributorReader{}, &mockExpertiseReader{},
	)

	result, err := svc.Search(context.Background(), repoID, "type:fact authentication")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, hit := range result.Hits {
		if hit.EntityType != EntityTypeFact {
			t.Errorf("expected only FACT hits, got %q", hit.EntityType)
		}
	}
}

// ─── Domain Search ────────────────────────────────────────────────────────────

func TestSearch_DomainSearch(t *testing.T) {
	repoID := "repo_1"
	svc := buildSearchService(
		&mockDecisionReader{}, &mockFactReader{}, &mockEventReader{},
		&mockRelationshipReader{},
		&mockContributorReader{contribs: []*models.Contributor{
			{ID: "ctr_1", RepositoryID: repoID, Name: "Alice", Email: "alice@example.com"},
		}},
		&mockExpertiseReader{expertise: []*models.Expertise{
			{ID: "exp_1", RepositoryID: repoID, ContributorID: "ctr_1", Domain: "Security", Score: 0.9, EvidenceJSON: `{}`},
		}},
	)

	result, err := svc.Search(context.Background(), repoID, "domain:security")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Hits) == 0 {
		t.Fatal("expected at least one expertise hit")
	}
	found := false
	for _, hit := range result.Hits {
		if hit.EntityType == EntityTypeExpertise && hit.Source == SourceOwnership {
			found = true
			ev := parseEvidence(t, hit)
			if ev.Field != "domain" {
				t.Errorf("expected domain field in evidence, got %q", ev.Field)
			}
		}
	}
	if !found {
		t.Error("expected expertise hit with ownership source")
	}
}

// ─── Contributor / Graph Sources ──────────────────────────────────────────────

func TestSearch_ContributorSearch(t *testing.T) {
	repoID := "repo_1"
	svc := buildSearchService(
		&mockDecisionReader{}, &mockFactReader{}, &mockEventReader{},
		&mockRelationshipReader{},
		&mockContributorReader{contribs: []*models.Contributor{
			{ID: "ctr_alice", RepositoryID: repoID, Name: "Alice", Email: "alice@example.com"},
		}},
		&mockExpertiseReader{},
	)

	result, err := svc.Search(context.Background(), repoID, "alice@example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Hits) == 0 {
		t.Fatal("expected contributor hit")
	}
	if result.Hits[0].EntityType != EntityTypeContributor {
		t.Errorf("expected CONTRIBUTOR, got %q", result.Hits[0].EntityType)
	}
	if result.Hits[0].Source != SourceOwnership {
		t.Errorf("expected ownership source, got %q", result.Hits[0].Source)
	}
}

func TestSearch_GraphRelationshipSearch(t *testing.T) {
	repoID := "repo_1"
	svc := buildSearchService(
		&mockDecisionReader{}, &mockFactReader{}, &mockEventReader{},
		&mockRelationshipReader{rels: []*memorymodels.Relationship{
			{ID: "rel_1", RepositoryID: repoID, FromID: "dec_1", ToID: "dec_2", Type: "DECISION_DEPENDS_ON", CreatedAt: time.Now()},
		}},
		&mockContributorReader{}, &mockExpertiseReader{},
	)

	result, err := svc.Search(context.Background(), repoID, "DECISION_DEPENDS_ON")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Hits) == 0 {
		t.Fatal("expected relationship hit")
	}
	if result.Hits[0].EntityType != EntityTypeRelationship {
		t.Errorf("expected RELATIONSHIP, got %q", result.Hits[0].EntityType)
	}
	if result.Hits[0].Source != SourceGraph {
		t.Errorf("expected graph source, got %q", result.Hits[0].Source)
	}
}

// ─── Discovery Source ─────────────────────────────────────────────────────────

func TestSearch_DiscoverySource(t *testing.T) {
	repoID := "repo_1"
	svc := buildSearchService(
		&mockDecisionReader{decisions: []*memorymodels.Decision{
			{ID: "dec_1", RepositoryID: repoID, Title: "Use Redis for caching", Status: "Accepted", CreatedAt: time.Now()},
		}},
		&mockFactReader{}, &mockEventReader{}, &mockRelationshipReader{},
		&mockContributorReader{}, &mockExpertiseReader{},
	)

	result, err := svc.Search(context.Background(), repoID, "Redis")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	hasMemory := false
	hasDiscovery := false
	for _, hit := range result.Hits {
		if hit.EntityID == "dec_1" {
			switch hit.Source {
			case SourceMemory:
				hasMemory = true
			case SourceDiscovery:
				hasDiscovery = true
			}
		}
	}
	// After dedup, only one hit per entity — memory should win over discovery at same score
	if !hasMemory && !hasDiscovery {
		t.Error("expected at least one hit for dec_1")
	}
	// Verify discovery path produces valid hits when searched with type filter only via memory
	result2, err := svc.Search(context.Background(), repoID, "type:decision Redis")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result2.Hits) == 0 {
		t.Error("expected structured discovery-compatible hit")
	}
}

// ─── Ordering & Determinism ───────────────────────────────────────────────────

func TestSearch_DeterministicOrdering(t *testing.T) {
	repoID := "repo_1"
	svc := buildSearchService(
		&mockDecisionReader{decisions: []*memorymodels.Decision{
			{ID: "dec_b", RepositoryID: repoID, Title: "cache layer", Status: "Accepted", CreatedAt: time.Now()},
			{ID: "dec_a", RepositoryID: repoID, Title: "cache store", Status: "Accepted", CreatedAt: time.Now()},
		}},
		&mockFactReader{facts: []*memorymodels.Fact{
			{ID: "fact_1", RepositoryID: repoID, Subject: "cache", Predicate: "is", Object: "redis", CreatedAt: time.Now()},
		}},
		&mockEventReader{}, &mockRelationshipReader{},
		&mockContributorReader{}, &mockExpertiseReader{},
	)

	result1, err := svc.Search(context.Background(), repoID, "cache")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	result2, err := svc.Search(context.Background(), repoID, "cache")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result1.Hits) != len(result2.Hits) {
		t.Fatalf("hit count mismatch: %d vs %d", len(result1.Hits), len(result2.Hits))
	}
	for i := range result1.Hits {
		if result1.Hits[i].EntityID != result2.Hits[i].EntityID {
			t.Errorf("hit %d entity ID mismatch", i)
		}
		if result1.Hits[i].MatchScore != result2.Hits[i].MatchScore {
			t.Errorf("hit %d score mismatch", i)
		}
	}

	// Verify descending score order
	for i := 1; i < len(result1.Hits); i++ {
		if result1.Hits[i].MatchScore > result1.Hits[i-1].MatchScore {
			t.Errorf("hits not sorted by score DESC at index %d", i)
		}
	}
}

func TestSearch_EvidenceValidJSON(t *testing.T) {
	repoID := "repo_1"
	svc := buildSearchService(
		&mockDecisionReader{decisions: []*memorymodels.Decision{
			{ID: "dec_1", RepositoryID: repoID, Title: "Use Redis", Status: "Accepted", CreatedAt: time.Now()},
		}},
		&mockFactReader{}, &mockEventReader{}, &mockRelationshipReader{},
		&mockContributorReader{}, &mockExpertiseReader{},
	)

	result, err := svc.Search(context.Background(), repoID, "Redis")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, hit := range result.Hits {
		if !json.Valid([]byte(hit.EvidenceJSON)) {
			t.Errorf("invalid evidence JSON on hit %s/%s", hit.EntityType, hit.EntityID)
		}
	}
}

// ─── Integration Test ─────────────────────────────────────────────────────────

func TestService_Integration(t *testing.T) {
	ctx := context.Background()

	tempDir, err := os.MkdirTemp("", "reponerve-search-integration-*")
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

	repoID := "repo_search"

	_, err = db.Exec(
		"INSERT INTO repositories (id, name, path, default_branch, created_at, updated_at) VALUES (?, ?, ?, ?, datetime(), datetime())",
		repoID, "Search Repo", tempDir, "main",
	)
	if err != nil {
		t.Fatalf("failed to insert repository: %v", err)
	}

	_, err = db.Exec(
		"INSERT INTO sources (id, repository_id, source_type, reference, title, author, timestamp, created_at) VALUES (?, ?, ?, ?, ?, ?, datetime(), datetime())",
		"src_1", repoID, "adr", "docs/adr/0001.md", "Use SQLite", "Alice <alice@example.com>",
	)
	if err != nil {
		t.Fatalf("failed to insert source: %v", err)
	}

	_, err = db.Exec(
		"INSERT INTO memory_decisions (id, repository_id, source_id, title, status, created_at) VALUES (?, ?, ?, ?, ?, datetime())",
		"dec_1", repoID, "src_1", "Use SQLite as primary store", "Accepted",
	)
	if err != nil {
		t.Fatalf("failed to insert decision: %v", err)
	}

	_, err = db.Exec(
		"INSERT INTO memory_facts (id, repository_id, source_id, subject, predicate, object, created_at) VALUES (?, ?, ?, ?, ?, ?, datetime())",
		"fact_1", repoID, "src_1", "database", "engine", "sqlite", time.Now().Format("2006-01-02 15:04:05"),
	)
	if err != nil {
		t.Fatalf("failed to insert fact: %v", err)
	}

	_, err = db.Exec(
		"INSERT INTO memory_relationships (id, repository_id, from_id, to_id, relationship_type, created_at) VALUES (?, ?, ?, ?, ?, datetime())",
		"rel_1", repoID, "dec_1", "fact_1", "DECISION_RELATES_TO",
	)
	if err != nil {
		t.Fatalf("failed to insert relationship: %v", err)
	}

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
	svc := NewService(dr, fr, er, rr, cr, expr, discoverySvc)

	t.Run("Search_PlainText", func(t *testing.T) {
		result, err := svc.Search(ctx, repoID, "SQLite")
		if err != nil {
			t.Fatalf("Search failed: %v", err)
		}
		if len(result.Hits) == 0 {
			t.Fatal("expected at least one hit")
		}
		if err := ValidateResult(result); err != nil {
			t.Errorf("ValidateResult failed: %v", err)
		}
	})

	t.Run("Search_StructuredType", func(t *testing.T) {
		result, err := svc.Search(ctx, repoID, "type:decision SQLite")
		if err != nil {
			t.Fatalf("Search failed: %v", err)
		}
		for _, hit := range result.Hits {
			if hit.EntityType != EntityTypeDecision {
				t.Errorf("expected DECISION, got %q", hit.EntityType)
			}
		}
	})

	t.Run("Search_Domain", func(t *testing.T) {
		result, err := svc.Search(ctx, repoID, "domain:Storage")
		if err != nil {
			t.Fatalf("Search failed: %v", err)
		}
		if len(result.Hits) == 0 {
			t.Fatal("expected expertise hit for domain:Storage")
		}
		if result.Hits[0].Source != SourceOwnership {
			t.Errorf("expected ownership source, got %q", result.Hits[0].Source)
		}
	})

	t.Run("Search_GraphRelationship", func(t *testing.T) {
		result, err := svc.Search(ctx, repoID, "DECISION_RELATES_TO")
		if err != nil {
			t.Fatalf("Search failed: %v", err)
		}
		found := false
		for _, hit := range result.Hits {
			if hit.EntityType == EntityTypeRelationship && hit.Source == SourceGraph {
				found = true
			}
		}
		if !found {
			t.Error("expected graph relationship hit")
		}
	})

	t.Run("Search_UnknownPrefix", func(t *testing.T) {
		_, err := svc.Search(ctx, repoID, "owner:alice")
		if err == nil {
			t.Error("expected error for unknown prefix")
		}
	})

	t.Run("Search_Determinism", func(t *testing.T) {
		r1, err := svc.Search(ctx, repoID, "sqlite")
		if err != nil {
			t.Fatalf("first search failed: %v", err)
		}
		r2, err := svc.Search(ctx, repoID, "sqlite")
		if err != nil {
			t.Fatalf("second search failed: %v", err)
		}
		if len(r1.Hits) != len(r2.Hits) {
			t.Fatalf("hit count mismatch")
		}
		for i := range r1.Hits {
			if r1.Hits[i].EntityID != r2.Hits[i].EntityID {
				t.Errorf("hit %d entity ID mismatch", i)
			}
		}
	})

	t.Run("Search_Evidence", func(t *testing.T) {
		result, err := svc.Search(ctx, repoID, "sqlite")
		if err != nil {
			t.Fatalf("Search failed: %v", err)
		}
		for _, hit := range result.Hits {
			if !json.Valid([]byte(hit.EvidenceJSON)) {
				t.Errorf("invalid evidence on hit %s/%s", hit.EntityType, hit.EntityID)
			}
			ev := parseEvidence(t, hit)
			if ev.MatchType == "" || ev.Field == "" {
				t.Errorf("incomplete evidence on hit %s/%s", hit.EntityType, hit.EntityID)
			}
		}
	})
}
