package agentsession

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	agentcontext "github.com/reponerve/reponerve/internal/agent/context"
	agentsearch "github.com/reponerve/reponerve/internal/agent/search"
	appcontext "github.com/reponerve/reponerve/internal/context"
	"github.com/reponerve/reponerve/internal/graph/impact"
	"github.com/reponerve/reponerve/internal/graph/relationships"
	"github.com/reponerve/reponerve/internal/graph/traversal"
	"github.com/reponerve/reponerve/internal/intelligence/changeplan"
	"github.com/reponerve/reponerve/internal/intelligence/discovery"
	"github.com/reponerve/reponerve/internal/intelligence/learning"
	"github.com/reponerve/reponerve/internal/intelligence/reviewers"
	memorymodels "github.com/reponerve/reponerve/internal/memory/models"
	"github.com/reponerve/reponerve/internal/query/storage"
	"github.com/reponerve/reponerve/internal/storage/migrations"
	"github.com/reponerve/reponerve/internal/storage/sqlite"
	models "github.com/reponerve/reponerve/pkg/models"
)

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

type mockSourceReader struct {
	sources []*models.Source
}

func (m *mockSourceReader) ListByRepository(_ context.Context, _ string) ([]*models.Source, error) {
	return m.sources, nil
}

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

func buildServices(
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
) (*Service, *agentcontext.Service, *agentsearch.Service) {
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

	contextSvc := agentcontext.NewService(discoverySvc, learningSvc, reviewerSvc, changePlanSvc, ctxGenerator)
	searchSvc := agentsearch.NewService(dr, fr, er, rr, cr, expr, discoverySvc, nil)

	return NewService(contextSvc, searchSvc), contextSvc, searchSvc
}

func emptyService() *Service {
	service, _, _ := buildServices(
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
	return service
}

func marshalJSON(t *testing.T, value any) json.RawMessage {
	t.Helper()
	payload, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("failed to marshal JSON: %v", err)
	}
	return json.RawMessage(payload)
}

func decodeContextPackage(t *testing.T, payload json.RawMessage) *agentcontext.AgentContextPackage {
	t.Helper()
	var pkg agentcontext.AgentContextPackage
	if err := json.Unmarshal(payload, &pkg); err != nil {
		t.Fatalf("failed to unmarshal context package: %v", err)
	}
	if err := agentcontext.ValidatePackage(&pkg); err != nil {
		t.Fatalf("invalid context package payload: %v", err)
	}
	return &pkg
}

func assertEquivalentContextPackage(t *testing.T, left, right json.RawMessage) {
	t.Helper()
	leftPkg := decodeContextPackage(t, left)
	rightPkg := decodeContextPackage(t, right)

	if leftPkg.RepositoryID != rightPkg.RepositoryID {
		t.Fatalf("repository ID mismatch: %q vs %q", leftPkg.RepositoryID, rightPkg.RepositoryID)
	}
	if len(leftPkg.Sections) != len(rightPkg.Sections) {
		t.Fatalf("section count mismatch: %d vs %d", len(leftPkg.Sections), len(rightPkg.Sections))
	}
	for i := range leftPkg.Sections {
		if leftPkg.Sections[i].Name != rightPkg.Sections[i].Name {
			t.Fatalf("section %d name mismatch: %q vs %q", i, leftPkg.Sections[i].Name, rightPkg.Sections[i].Name)
		}
		if leftPkg.Sections[i].Source != rightPkg.Sections[i].Source {
			t.Fatalf("section %d source mismatch: %q vs %q", i, leftPkg.Sections[i].Source, rightPkg.Sections[i].Source)
		}
		if !json.Valid(leftPkg.Sections[i].Data) || !json.Valid(rightPkg.Sections[i].Data) {
			t.Fatalf("section %d contains invalid JSON data", i)
		}
	}
}

func decodeSearchResult(t *testing.T, payload json.RawMessage) *agentsearch.SearchResult {
	t.Helper()
	var result agentsearch.SearchResult
	if err := json.Unmarshal(payload, &result); err != nil {
		t.Fatalf("failed to unmarshal search result: %v", err)
	}
	if err := agentsearch.ValidateResult(&result); err != nil {
		t.Fatalf("invalid search result payload: %v", err)
	}
	return &result
}

func TestNewArtifact_PreservesSerializedPayload(t *testing.T) {
	expected := &agentsearch.SearchResult{
		RepositoryID: "repo_1",
		Query:        "domain:Storage",
		Hits: []*agentsearch.SearchHit{
			{
				EntityType:   agentsearch.EntityTypeExpertise,
				EntityID:     "exp_1",
				Source:       agentsearch.SourceOwnership,
				MatchScore:   agentsearch.ScoreExact,
				EvidenceJSON: `{"field":"domain","match_type":"exact"}`,
			},
		},
	}

	artifact, err := newArtifact(ArtifactTypeSearchResult, SourceSearch, expected)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(artifact.Data) != string(marshalJSON(t, expected)) {
		t.Fatal("artifact payload was modified during serialization")
	}
}

func TestValidateSession_Nil(t *testing.T) {
	if err := ValidateSession(nil); err == nil {
		t.Fatal("expected error for nil session")
	}
}

func TestValidateSession_MissingSessionID(t *testing.T) {
	session := &AgentSession{
		SessionType:  SessionTypeRepository,
		RepositoryID: "repo_1",
		Artifacts: []*SessionArtifact{
			{ArtifactType: ArtifactTypeContextPackage, Source: SourceContext, Data: json.RawMessage(`{}`)},
		},
	}
	if err := ValidateSession(session); err == nil {
		t.Fatal("expected error for missing session ID")
	}
}

func TestValidateSession_InvalidSessionType(t *testing.T) {
	session := &AgentSession{
		SessionID:    "ses_1",
		SessionType:  "unknown",
		RepositoryID: "repo_1",
		Artifacts: []*SessionArtifact{
			{ArtifactType: ArtifactTypeContextPackage, Source: SourceContext, Data: json.RawMessage(`{}`)},
		},
	}
	if err := ValidateSession(session); err == nil {
		t.Fatal("expected error for invalid session type")
	}
}

func TestValidateSession_InvalidArtifactType(t *testing.T) {
	session := &AgentSession{
		SessionID:    "ses_1",
		SessionType:  SessionTypeRepository,
		RepositoryID: "repo_1",
		Artifacts: []*SessionArtifact{
			{ArtifactType: "unknown", Source: SourceContext, Data: json.RawMessage(`{}`)},
		},
	}
	if err := ValidateSession(session); err == nil {
		t.Fatal("expected error for invalid artifact type")
	}
}

func TestValidateSession_InvalidSource(t *testing.T) {
	session := &AgentSession{
		SessionID:    "ses_1",
		SessionType:  SessionTypeRepository,
		RepositoryID: "repo_1",
		Artifacts: []*SessionArtifact{
			{ArtifactType: ArtifactTypeContextPackage, Source: "unknown", Data: json.RawMessage(`{}`)},
		},
	}
	if err := ValidateSession(session); err == nil {
		t.Fatal("expected error for invalid source")
	}
}

func TestValidateSession_InvalidData(t *testing.T) {
	session := &AgentSession{
		SessionID:    "ses_1",
		SessionType:  SessionTypeRepository,
		RepositoryID: "repo_1",
		Artifacts: []*SessionArtifact{
			{ArtifactType: ArtifactTypeContextPackage, Source: SourceContext, Data: json.RawMessage(`{`)},
		},
	}
	if err := ValidateSession(session); err == nil {
		t.Fatal("expected error for invalid JSON data")
	}
}

func TestCreateRepositorySession_EmptyRepositoryID(t *testing.T) {
	_, err := emptyService().CreateRepositorySession(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for empty repository ID")
	}
}

func TestCreateDomainSession_EmptyDomain(t *testing.T) {
	_, err := emptyService().CreateDomainSession(context.Background(), "repo_1", "")
	if err == nil {
		t.Fatal("expected error for empty domain")
	}
}

func TestCreateContributorSession_EmptyContributorID(t *testing.T) {
	_, err := emptyService().CreateContributorSession(context.Background(), "repo_1", "")
	if err == nil {
		t.Fatal("expected error for empty contributor ID")
	}
}

func TestCreateRepositorySession(t *testing.T) {
	service := emptyService()

	session, err := service.CreateRepositorySession(context.Background(), "repo_1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if session.SessionType != SessionTypeRepository {
		t.Fatalf("expected session type %q, got %q", SessionTypeRepository, session.SessionType)
	}
	if session.SessionID != buildSessionID("repo_1", SessionTypeRepository, defaultRepositoryIdentifier) {
		t.Fatalf("unexpected session ID %q", session.SessionID)
	}
	if len(session.Artifacts) != 1 {
		t.Fatalf("expected 1 artifact, got %d", len(session.Artifacts))
	}
	if session.Artifacts[0].ArtifactType != ArtifactTypeContextPackage {
		t.Fatalf("expected context package artifact, got %q", session.Artifacts[0].ArtifactType)
	}
	if session.Artifacts[0].Source != SourceContext {
		t.Fatalf("expected context source, got %q", session.Artifacts[0].Source)
	}
	if err := ValidateSession(session); err != nil {
		t.Fatalf("expected valid session, got %v", err)
	}
}

func TestCreateDomainSession_OrderAndArtifactPreservation(t *testing.T) {
	repoID := "repo_1"
	domain := "Storage"

	decisions := []*memorymodels.Decision{
		{ID: "dec_1", RepositoryID: repoID, Title: "Use SQLite storage", Status: "Accepted", CreatedAt: time.Now()},
	}
	contributors := []*models.Contributor{
		{ID: "ctr_alice", RepositoryID: repoID, Name: "Alice", Email: "alice@example.com"},
	}
	expertise := []*models.Expertise{
		{ID: "exp_1", RepositoryID: repoID, ContributorID: "ctr_alice", Domain: domain, Score: 0.9, EvidenceJSON: `{"reason":"match"}`},
	}

	service, contextSvc, searchSvc := buildServices(
		&mockDecisionReader{decisions: decisions},
		&mockIntentReader{},
		&mockFactReader{},
		&mockEventReader{},
		&mockRelationshipReader{},
		&mockContributorReader{contribs: contributors},
		&mockExpertiseReader{expertise: expertise},
		&mockSourceReader{},
		decisions, nil, nil, nil,
	)

	session, err := service.CreateDomainSession(context.Background(), repoID, domain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(session.Artifacts) != 2 {
		t.Fatalf("expected 2 artifacts, got %d", len(session.Artifacts))
	}
	if session.Artifacts[0].ArtifactType != ArtifactTypeContextPackage || session.Artifacts[0].Source != SourceContext {
		t.Fatalf("unexpected first artifact: %#v", session.Artifacts[0])
	}
	if session.Artifacts[1].ArtifactType != ArtifactTypeSearchResult || session.Artifacts[1].Source != SourceSearch {
		t.Fatalf("unexpected second artifact: %#v", session.Artifacts[1])
	}

	expectedContext, err := contextSvc.BuildDomainContext(context.Background(), repoID, domain)
	if err != nil {
		t.Fatalf("failed to build expected context: %v", err)
	}
	expectedSearch, err := searchSvc.Search(context.Background(), repoID, "domain:"+domain)
	if err != nil {
		t.Fatalf("failed to build expected search: %v", err)
	}

	assertEquivalentContextPackage(t, session.Artifacts[0].Data, marshalJSON(t, expectedContext))
	if string(session.Artifacts[1].Data) != string(marshalJSON(t, expectedSearch)) {
		t.Fatal("domain search artifact was modified during packaging")
	}

	result := decodeSearchResult(t, session.Artifacts[1].Data)
	if result.Query != "domain:"+domain {
		t.Fatalf("expected query %q, got %q", "domain:"+domain, result.Query)
	}
}

func TestCreateContributorSession_OrderAndArtifactPreservation(t *testing.T) {
	repoID := "repo_1"
	contributorID := "ctr_alice"

	decisions := []*memorymodels.Decision{
		{ID: "dec_1", RepositoryID: repoID, Title: "Use SQLite", Status: "Accepted", CreatedAt: time.Now()},
	}
	contributors := []*models.Contributor{
		{ID: contributorID, RepositoryID: repoID, Name: "Alice", Email: "alice@example.com"},
	}
	expertise := []*models.Expertise{
		{ID: "exp_1", RepositoryID: repoID, ContributorID: contributorID, Domain: "Storage", Score: 0.9, EvidenceJSON: `{"reason":"owner"}`},
	}

	service, contextSvc, searchSvc := buildServices(
		&mockDecisionReader{decisions: decisions},
		&mockIntentReader{},
		&mockFactReader{},
		&mockEventReader{},
		&mockRelationshipReader{},
		&mockContributorReader{contribs: contributors},
		&mockExpertiseReader{expertise: expertise},
		&mockSourceReader{},
		decisions, nil, nil, nil,
	)

	session, err := service.CreateContributorSession(context.Background(), repoID, contributorID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(session.Artifacts) != 2 {
		t.Fatalf("expected 2 artifacts, got %d", len(session.Artifacts))
	}

	expectedContext, err := contextSvc.BuildContributorContext(context.Background(), repoID, contributorID)
	if err != nil {
		t.Fatalf("failed to build expected contributor context: %v", err)
	}
	expectedSearch, err := searchSvc.Search(context.Background(), repoID, contributorID)
	if err != nil {
		t.Fatalf("failed to build expected contributor search: %v", err)
	}

	assertEquivalentContextPackage(t, session.Artifacts[0].Data, marshalJSON(t, expectedContext))
	if string(session.Artifacts[1].Data) != string(marshalJSON(t, expectedSearch)) {
		t.Fatal("contributor search artifact was modified during packaging")
	}

	result := decodeSearchResult(t, session.Artifacts[1].Data)
	if result.Query != contributorID {
		t.Fatalf("expected query %q, got %q", contributorID, result.Query)
	}
}

func TestCreateSessions_DeterministicIDsAndReconstruction(t *testing.T) {
	repoID := "repo_1"
	domain := "Storage"
	contributorID := "ctr_alice"

	decisions := []*memorymodels.Decision{
		{ID: "dec_1", RepositoryID: repoID, Title: "Use SQLite", Status: "Accepted", CreatedAt: time.Now()},
	}
	contributors := []*models.Contributor{
		{ID: contributorID, RepositoryID: repoID, Name: "Alice", Email: "alice@example.com"},
	}
	expertise := []*models.Expertise{
		{ID: "exp_1", RepositoryID: repoID, ContributorID: contributorID, Domain: domain, Score: 0.9, EvidenceJSON: `{"reason":"owner"}`},
	}

	service, _, _ := buildServices(
		&mockDecisionReader{decisions: decisions},
		&mockIntentReader{},
		&mockFactReader{},
		&mockEventReader{},
		&mockRelationshipReader{},
		&mockContributorReader{contribs: contributors},
		&mockExpertiseReader{expertise: expertise},
		&mockSourceReader{},
		decisions, nil, nil, nil,
	)

	session1, err := service.CreateDomainSession(context.Background(), repoID, domain)
	if err != nil {
		t.Fatalf("first domain session failed: %v", err)
	}
	session2, err := service.CreateDomainSession(context.Background(), repoID, domain)
	if err != nil {
		t.Fatalf("second domain session failed: %v", err)
	}

	if session1.SessionID != session2.SessionID {
		t.Fatalf("expected deterministic domain session ID, got %q and %q", session1.SessionID, session2.SessionID)
	}
	if session1.SessionID != buildSessionID(repoID, SessionTypeDomain, domain) {
		t.Fatalf("unexpected deterministic domain session ID %q", session1.SessionID)
	}
	if len(session1.Artifacts) != len(session2.Artifacts) {
		t.Fatalf("domain artifact count mismatch: %d vs %d", len(session1.Artifacts), len(session2.Artifacts))
	}
	assertEquivalentContextPackage(t, session1.Artifacts[0].Data, session2.Artifacts[0].Data)
	if string(session1.Artifacts[1].Data) != string(session2.Artifacts[1].Data) {
		t.Fatal("domain search artifact differs between repeated sessions")
	}

	contributorSession1, err := service.CreateContributorSession(context.Background(), repoID, contributorID)
	if err != nil {
		t.Fatalf("first contributor session failed: %v", err)
	}
	contributorSession2, err := service.CreateContributorSession(context.Background(), repoID, contributorID)
	if err != nil {
		t.Fatalf("second contributor session failed: %v", err)
	}

	if contributorSession1.SessionID != contributorSession2.SessionID {
		t.Fatalf("expected deterministic contributor session ID, got %q and %q", contributorSession1.SessionID, contributorSession2.SessionID)
	}
	if len(contributorSession1.Artifacts) != len(contributorSession2.Artifacts) {
		t.Fatalf("contributor artifact count mismatch: %d vs %d", len(contributorSession1.Artifacts), len(contributorSession2.Artifacts))
	}
	assertEquivalentContextPackage(t, contributorSession1.Artifacts[0].Data, contributorSession2.Artifacts[0].Data)
	if string(contributorSession1.Artifacts[1].Data) != string(contributorSession2.Artifacts[1].Data) {
		t.Fatal("contributor search artifact differs between repeated sessions")
	}
}

func TestService_Integration(t *testing.T) {
	ctx := context.Background()

	tempDir, err := os.MkdirTemp("", "reponerve-agentsession-integration-*")
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

	repoID := "repo_agentsession"
	contributorID := "ctr_alice"
	domain := "Storage"

	_, err = db.Exec(
		"INSERT INTO repositories (id, name, path, default_branch, created_at, updated_at) VALUES (?, ?, ?, ?, datetime(), datetime())",
		repoID, "Agent Session Repo", tempDir, "main",
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
		"INSERT INTO memory_relationships (id, repository_id, from_id, to_id, relationship_type, created_at) VALUES (?, ?, ?, ?, ?, datetime())",
		"rel_1", repoID, "dec_1", "dec_1", "DECISION_RELATES_TO",
	)
	if err != nil {
		t.Fatalf("failed to insert relationship: %v", err)
	}

	_, err = db.Exec(
		"INSERT INTO contributors (id, repository_id, name, email, first_seen, last_seen, commit_count) VALUES (?, ?, ?, ?, datetime(), datetime(), ?)",
		contributorID, repoID, "Alice", "alice@example.com", 10,
	)
	if err != nil {
		t.Fatalf("failed to insert contributor: %v", err)
	}

	_, err = db.Exec(
		"INSERT INTO expertise (id, repository_id, contributor_id, domain, score, evidence_json) VALUES (?, ?, ?, ?, ?, ?)",
		"exp_1", repoID, contributorID, domain, 0.9, `{"explanation":"expert contributor"}`,
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
	learningSvc := learning.NewService(discoverySvc, dr, fr, er, cr, expr, sr, relEngine)
	reviewerSvc := reviewers.NewService(discoverySvc, dr, fr, er, cr, expr, sr, impactSvc)
	changePlanSvc := changeplan.NewService(impactSvc)
	ctxReader := appcontext.NewMemoryContextReader(er, dr, ir, fr)
	ctxGenerator := appcontext.NewGenerator(ctxReader)

	contextSvc := agentcontext.NewService(discoverySvc, learningSvc, reviewerSvc, changePlanSvc, ctxGenerator)
	searchSvc := agentsearch.NewService(dr, fr, er, rr, cr, expr, discoverySvc, nil)
	service := NewService(contextSvc, searchSvc)

	t.Run("CreateRepositorySession", func(t *testing.T) {
		session, err := service.CreateRepositorySession(ctx, repoID)
		if err != nil {
			t.Fatalf("CreateRepositorySession failed: %v", err)
		}
		if err := ValidateSession(session); err != nil {
			t.Fatalf("ValidateSession failed: %v", err)
		}
		if len(session.Artifacts) != 1 {
			t.Fatalf("expected 1 artifact, got %d", len(session.Artifacts))
		}

		expectedContext, err := contextSvc.BuildRepositoryContext(ctx, repoID)
		if err != nil {
			t.Fatalf("failed to build expected repository context: %v", err)
		}
		assertEquivalentContextPackage(t, session.Artifacts[0].Data, marshalJSON(t, expectedContext))
	})

	t.Run("CreateDomainSession", func(t *testing.T) {
		session, err := service.CreateDomainSession(ctx, repoID, domain)
		if err != nil {
			t.Fatalf("CreateDomainSession failed: %v", err)
		}
		if err := ValidateSession(session); err != nil {
			t.Fatalf("ValidateSession failed: %v", err)
		}
		if len(session.Artifacts) != 2 {
			t.Fatalf("expected 2 artifacts, got %d", len(session.Artifacts))
		}
		if session.Artifacts[0].ArtifactType != ArtifactTypeContextPackage || session.Artifacts[1].ArtifactType != ArtifactTypeSearchResult {
			t.Fatalf("unexpected artifact ordering")
		}

		expectedContext, err := contextSvc.BuildDomainContext(ctx, repoID, domain)
		if err != nil {
			t.Fatalf("failed to build expected domain context: %v", err)
		}
		expectedSearch, err := searchSvc.Search(ctx, repoID, "domain:"+domain)
		if err != nil {
			t.Fatalf("failed to build expected domain search: %v", err)
		}
		assertEquivalentContextPackage(t, session.Artifacts[0].Data, marshalJSON(t, expectedContext))
		if string(session.Artifacts[1].Data) != string(marshalJSON(t, expectedSearch)) {
			t.Fatal("domain session did not preserve search artifact")
		}
	})

	t.Run("CreateContributorSession", func(t *testing.T) {
		session1, err := service.CreateContributorSession(ctx, repoID, contributorID)
		if err != nil {
			t.Fatalf("CreateContributorSession failed: %v", err)
		}
		session2, err := service.CreateContributorSession(ctx, repoID, contributorID)
		if err != nil {
			t.Fatalf("CreateContributorSession second call failed: %v", err)
		}
		if err := ValidateSession(session1); err != nil {
			t.Fatalf("ValidateSession failed: %v", err)
		}
		if session1.SessionID != session2.SessionID {
			t.Fatalf("expected deterministic session IDs, got %q and %q", session1.SessionID, session2.SessionID)
		}
		assertEquivalentContextPackage(t, session1.Artifacts[0].Data, session2.Artifacts[0].Data)
		if string(session1.Artifacts[1].Data) != string(session2.Artifacts[1].Data) {
			t.Fatal("contributor search artifact differs between reconstructed sessions")
		}

		expectedContext, err := contextSvc.BuildContributorContext(ctx, repoID, contributorID)
		if err != nil {
			t.Fatalf("failed to build expected contributor context: %v", err)
		}
		expectedSearch, err := searchSvc.Search(ctx, repoID, contributorID)
		if err != nil {
			t.Fatalf("failed to build expected contributor search: %v", err)
		}
		assertEquivalentContextPackage(t, session1.Artifacts[0].Data, marshalJSON(t, expectedContext))
		if string(session1.Artifacts[1].Data) != string(marshalJSON(t, expectedSearch)) {
			t.Fatal("contributor session did not preserve search artifact")
		}
	})
}
