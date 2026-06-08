package workflow

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
	agentsession "github.com/reponerve/reponerve/internal/agent/session"
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
) (*Service, *discovery.Service, *learning.Service, *reviewers.Service, *changeplan.Service, *agentcontext.Service, *agentsearch.Service, *agentsession.Service) {
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
	searchSvc := agentsearch.NewService(dr, fr, er, rr, cr, expr, discoverySvc)
	sessionSvc := agentsession.NewService(contextSvc, searchSvc)
	workflowSvc := NewService(discoverySvc, learningSvc, reviewerSvc, changePlanSvc, contextSvc, searchSvc, sessionSvc)

	return workflowSvc, discoverySvc, learningSvc, reviewerSvc, changePlanSvc, contextSvc, searchSvc, sessionSvc
}

func emptyService() *Service {
	service, _, _, _, _, _, _, _ := buildServices(
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

func decodeSession(t *testing.T, payload json.RawMessage) *agentsession.AgentSession {
	t.Helper()
	var session agentsession.AgentSession
	if err := json.Unmarshal(payload, &session); err != nil {
		t.Fatalf("failed to unmarshal session: %v", err)
	}
	if err := agentsession.ValidateSession(&session); err != nil {
		t.Fatalf("invalid session payload: %v", err)
	}
	return &session
}

func assertEquivalentSession(t *testing.T, left, right json.RawMessage) {
	t.Helper()
	leftSession := decodeSession(t, left)
	rightSession := decodeSession(t, right)

	if leftSession.SessionID != rightSession.SessionID {
		t.Fatalf("session ID mismatch: %q vs %q", leftSession.SessionID, rightSession.SessionID)
	}
	if leftSession.SessionType != rightSession.SessionType {
		t.Fatalf("session type mismatch: %q vs %q", leftSession.SessionType, rightSession.SessionType)
	}
	if leftSession.RepositoryID != rightSession.RepositoryID {
		t.Fatalf("repository ID mismatch: %q vs %q", leftSession.RepositoryID, rightSession.RepositoryID)
	}
	if len(leftSession.Artifacts) != len(rightSession.Artifacts) {
		t.Fatalf("artifact count mismatch: %d vs %d", len(leftSession.Artifacts), len(rightSession.Artifacts))
	}
	for i := range leftSession.Artifacts {
		if leftSession.Artifacts[i].ArtifactType != rightSession.Artifacts[i].ArtifactType {
			t.Fatalf("artifact %d type mismatch: %q vs %q", i, leftSession.Artifacts[i].ArtifactType, rightSession.Artifacts[i].ArtifactType)
		}
		if leftSession.Artifacts[i].Source != rightSession.Artifacts[i].Source {
			t.Fatalf("artifact %d source mismatch: %q vs %q", i, leftSession.Artifacts[i].Source, rightSession.Artifacts[i].Source)
		}
		if !json.Valid(leftSession.Artifacts[i].Data) || !json.Valid(rightSession.Artifacts[i].Data) {
			t.Fatalf("artifact %d contains invalid JSON payload", i)
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
			t.Fatalf("section %d contains invalid JSON payload", i)
		}
	}
}

func TestNewArtifact_PreservesSerializedPayload(t *testing.T) {
	expected := &agentsearch.SearchResult{
		RepositoryID: "repo_1",
		Query:        "SQLite",
		Hits: []*agentsearch.SearchHit{
			{
				EntityType:   agentsearch.EntityTypeDecision,
				EntityID:     "dec_1",
				Source:       agentsearch.SourceMemory,
				MatchScore:   agentsearch.ScoreExact,
				EvidenceJSON: `{"field":"id","match_type":"exact"}`,
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

func TestValidateWorkflow_Nil(t *testing.T) {
	if err := ValidateWorkflow(nil); err == nil {
		t.Fatal("expected error for nil workflow")
	}
}

func TestValidateWorkflow_InvalidWorkflowType(t *testing.T) {
	workflow := &WorkflowPackage{
		WorkflowID:   "wrk_1",
		WorkflowType: "unknown",
		RepositoryID: "repo_1",
		Version:      VersionV1,
		Artifacts: []*WorkflowArtifact{
			{ArtifactType: ArtifactTypeSession, Source: SourceSession, Data: json.RawMessage(`{}`)},
		},
	}
	if err := ValidateWorkflow(workflow); err == nil {
		t.Fatal("expected error for invalid workflow type")
	}
}

func TestValidateWorkflow_InvalidArtifactType(t *testing.T) {
	workflow := &WorkflowPackage{
		WorkflowID:   "wrk_1",
		WorkflowType: WorkflowTypeOnboarding,
		RepositoryID: "repo_1",
		Version:      VersionV1,
		Artifacts: []*WorkflowArtifact{
			{ArtifactType: "unknown", Source: SourceSession, Data: json.RawMessage(`{}`)},
		},
	}
	if err := ValidateWorkflow(workflow); err == nil {
		t.Fatal("expected error for invalid artifact type")
	}
}

func TestValidateWorkflow_InvalidSource(t *testing.T) {
	workflow := &WorkflowPackage{
		WorkflowID:   "wrk_1",
		WorkflowType: WorkflowTypeOnboarding,
		RepositoryID: "repo_1",
		Version:      VersionV1,
		Artifacts: []*WorkflowArtifact{
			{ArtifactType: ArtifactTypeSession, Source: "unknown", Data: json.RawMessage(`{}`)},
		},
	}
	if err := ValidateWorkflow(workflow); err == nil {
		t.Fatal("expected error for invalid source")
	}
}

func TestBuildOnboardingWorkflow(t *testing.T) {
	service := emptyService()

	workflow, err := service.BuildOnboardingWorkflow(context.Background(), "repo_1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if workflow.Version != VersionV1 {
		t.Fatalf("expected version %q, got %q", VersionV1, workflow.Version)
	}
	if workflow.WorkflowType != WorkflowTypeOnboarding {
		t.Fatalf("expected workflow type %q, got %q", WorkflowTypeOnboarding, workflow.WorkflowType)
	}
	if len(workflow.Artifacts) != 3 {
		t.Fatalf("expected 3 artifacts, got %d", len(workflow.Artifacts))
	}

	expectedTypes := []string{ArtifactTypeSession, ArtifactTypeDiscoveryReport, ArtifactTypeLearningPath}
	expectedSources := []string{SourceSession, SourceDiscovery, SourceLearning}
	for i := range workflow.Artifacts {
		if workflow.Artifacts[i].ArtifactType != expectedTypes[i] {
			t.Fatalf("artifact %d: expected type %q, got %q", i, expectedTypes[i], workflow.Artifacts[i].ArtifactType)
		}
		if workflow.Artifacts[i].Source != expectedSources[i] {
			t.Fatalf("artifact %d: expected source %q, got %q", i, expectedSources[i], workflow.Artifacts[i].Source)
		}
	}
}

func TestBuildReviewPreparationWorkflow(t *testing.T) {
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

	service, _, _, reviewerSvc, _, contextSvc, searchSvc, sessionSvc := buildServices(
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

	workflow, err := service.BuildReviewPreparationWorkflow(context.Background(), repoID, "domain:"+domain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(workflow.Artifacts) != 4 {
		t.Fatalf("expected 4 artifacts, got %d", len(workflow.Artifacts))
	}

	expectedSession, err := sessionSvc.CreateRepositorySession(context.Background(), repoID)
	if err != nil {
		t.Fatalf("failed to create expected session: %v", err)
	}
	expectedReviewers, err := reviewerSvc.RecommendDomainReviewers(context.Background(), repoID, domain)
	if err != nil {
		t.Fatalf("failed to create expected reviewer report: %v", err)
	}
	expectedSearch, err := searchSvc.Search(context.Background(), repoID, "domain:"+domain)
	if err != nil {
		t.Fatalf("failed to create expected search result: %v", err)
	}
	expectedContext, err := contextSvc.BuildRepositoryContext(context.Background(), repoID)
	if err != nil {
		t.Fatalf("failed to create expected context package: %v", err)
	}

	assertEquivalentSession(t, workflow.Artifacts[0].Data, marshalJSON(t, expectedSession))
	if string(workflow.Artifacts[1].Data) != string(marshalJSON(t, expectedReviewers)) {
		t.Fatal("reviewer report was modified during packaging")
	}
	if string(workflow.Artifacts[2].Data) != string(marshalJSON(t, expectedSearch)) {
		t.Fatal("search result was modified during packaging")
	}
	assertEquivalentContextPackage(t, workflow.Artifacts[3].Data, marshalJSON(t, expectedContext))
}

func TestBuildChangePreparationWorkflow(t *testing.T) {
	repoID := "repo_1"

	decisions := []*memorymodels.Decision{
		{ID: "dec_1", RepositoryID: repoID, Title: "Use SQLite", Status: "Accepted", CreatedAt: time.Now()},
		{ID: "dec_2", RepositoryID: repoID, Title: "Enable WAL", Status: "Accepted", CreatedAt: time.Now()},
	}
	rels := []*memorymodels.Relationship{
		{ID: "rel_1", RepositoryID: repoID, FromID: "dec_1", ToID: "dec_2", Type: "DECISION_DEPENDS_ON", CreatedAt: time.Now()},
	}

	service, _, _, _, changePlanSvc, contextSvc, searchSvc, sessionSvc := buildServices(
		&mockDecisionReader{decisions: decisions},
		&mockIntentReader{},
		&mockFactReader{},
		&mockEventReader{},
		&mockRelationshipReader{rels: rels},
		&mockContributorReader{},
		&mockExpertiseReader{},
		&mockSourceReader{},
		decisions, nil, nil, nil,
	)

	workflow, err := service.BuildChangePreparationWorkflow(context.Background(), repoID, "dec_1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(workflow.Artifacts) != 4 {
		t.Fatalf("expected 4 artifacts, got %d", len(workflow.Artifacts))
	}

	expectedSession, err := sessionSvc.CreateRepositorySession(context.Background(), repoID)
	if err != nil {
		t.Fatalf("failed to create expected session: %v", err)
	}
	expectedPlan, err := changePlanSvc.GenerateDecisionPlan(context.Background(), repoID, "dec_1")
	if err != nil {
		t.Fatalf("failed to create expected change plan: %v", err)
	}
	expectedSearch, err := searchSvc.Search(context.Background(), repoID, "dec_1")
	if err != nil {
		t.Fatalf("failed to create expected search result: %v", err)
	}
	expectedContext, err := contextSvc.BuildRepositoryContext(context.Background(), repoID)
	if err != nil {
		t.Fatalf("failed to create expected context package: %v", err)
	}

	assertEquivalentSession(t, workflow.Artifacts[0].Data, marshalJSON(t, expectedSession))
	if string(workflow.Artifacts[1].Data) != string(marshalJSON(t, expectedPlan)) {
		t.Fatal("change plan was modified during packaging")
	}
	if string(workflow.Artifacts[2].Data) != string(marshalJSON(t, expectedSearch)) {
		t.Fatal("search result was modified during packaging")
	}
	assertEquivalentContextPackage(t, workflow.Artifacts[3].Data, marshalJSON(t, expectedContext))
}

func TestBuildKnowledgeExplorationWorkflow(t *testing.T) {
	repoID := "repo_1"
	decisions := []*memorymodels.Decision{
		{ID: "dec_1", RepositoryID: repoID, Title: "Use SQLite", Status: "Accepted", CreatedAt: time.Now()},
	}

	service, _, _, _, _, _, searchSvc, sessionSvc := buildServices(
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

	workflow, err := service.BuildKnowledgeExplorationWorkflow(context.Background(), repoID, "SQLite")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(workflow.Artifacts) != 2 {
		t.Fatalf("expected 2 artifacts, got %d", len(workflow.Artifacts))
	}

	expectedSession, err := sessionSvc.CreateRepositorySession(context.Background(), repoID)
	if err != nil {
		t.Fatalf("failed to create expected session: %v", err)
	}
	expectedSearch, err := searchSvc.Search(context.Background(), repoID, "SQLite")
	if err != nil {
		t.Fatalf("failed to create expected search result: %v", err)
	}

	assertEquivalentSession(t, workflow.Artifacts[0].Data, marshalJSON(t, expectedSession))
	if string(workflow.Artifacts[1].Data) != string(marshalJSON(t, expectedSearch)) {
		t.Fatal("search result was modified during packaging")
	}
}

func TestBuildWorkflow_DeterministicIDsAndReconstruction(t *testing.T) {
	repoID := "repo_1"
	decisions := []*memorymodels.Decision{
		{ID: "dec_1", RepositoryID: repoID, Title: "Use SQLite", Status: "Accepted", CreatedAt: time.Now()},
	}

	service, _, _, _, _, _, _, _ := buildServices(
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

	workflow1, err := service.BuildKnowledgeExplorationWorkflow(context.Background(), repoID, "SQLite")
	if err != nil {
		t.Fatalf("first workflow failed: %v", err)
	}
	workflow2, err := service.BuildKnowledgeExplorationWorkflow(context.Background(), repoID, "SQLite")
	if err != nil {
		t.Fatalf("second workflow failed: %v", err)
	}

	if workflow1.WorkflowID != workflow2.WorkflowID {
		t.Fatalf("expected deterministic workflow ID, got %q and %q", workflow1.WorkflowID, workflow2.WorkflowID)
	}
	if workflow1.WorkflowID != buildWorkflowID(repoID, WorkflowTypeKnowledgeExploration, "SQLite") {
		t.Fatalf("unexpected workflow ID %q", workflow1.WorkflowID)
	}
	if len(workflow1.Artifacts) != len(workflow2.Artifacts) {
		t.Fatalf("artifact count mismatch: %d vs %d", len(workflow1.Artifacts), len(workflow2.Artifacts))
	}
	assertEquivalentSession(t, workflow1.Artifacts[0].Data, workflow2.Artifacts[0].Data)
	if string(workflow1.Artifacts[1].Data) != string(workflow2.Artifacts[1].Data) {
		t.Fatal("search result differs between repeated workflows")
	}
}

func TestBuildChangePreparationWorkflow_UnsupportedEntity(t *testing.T) {
	service := emptyService()
	_, err := service.BuildChangePreparationWorkflow(context.Background(), "repo_1", "missing")
	if err == nil {
		t.Fatal("expected error for unsupported change entity")
	}
}

func TestService_Integration(t *testing.T) {
	ctx := context.Background()

	tempDir, err := os.MkdirTemp("", "reponerve-workflow-integration-*")
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

	repoID := "repo_workflow"
	contributorID := "ctr_alice"
	domain := "Storage"

	_, err = db.Exec(
		"INSERT INTO repositories (id, name, path, default_branch, created_at, updated_at) VALUES (?, ?, ?, ?, datetime(), datetime())",
		repoID, "Workflow Repo", tempDir, "main",
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
		"INSERT INTO memory_decisions (id, repository_id, source_id, title, status, created_at) VALUES (?, ?, ?, ?, ?, datetime())",
		"dec_2", repoID, "src_1", "Enable WAL mode", "Accepted",
	)
	if err != nil {
		t.Fatalf("failed to insert decision: %v", err)
	}

	_, err = db.Exec(
		"INSERT INTO memory_relationships (id, repository_id, from_id, to_id, relationship_type, created_at) VALUES (?, ?, ?, ?, ?, datetime())",
		"rel_1", repoID, "dec_1", "dec_2", "DECISION_DEPENDS_ON",
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
	searchSvc := agentsearch.NewService(dr, fr, er, rr, cr, expr, discoverySvc)
	sessionSvc := agentsession.NewService(contextSvc, searchSvc)
	service := NewService(discoverySvc, learningSvc, reviewerSvc, changePlanSvc, contextSvc, searchSvc, sessionSvc)

	t.Run("Onboarding", func(t *testing.T) {
		workflow, err := service.BuildOnboardingWorkflow(ctx, repoID)
		if err != nil {
			t.Fatalf("BuildOnboardingWorkflow failed: %v", err)
		}
		if err := ValidateWorkflow(workflow); err != nil {
			t.Fatalf("ValidateWorkflow failed: %v", err)
		}
		if len(workflow.Artifacts) != 3 {
			t.Fatalf("expected 3 artifacts, got %d", len(workflow.Artifacts))
		}
	})

	t.Run("ReviewPreparation", func(t *testing.T) {
		workflow, err := service.BuildReviewPreparationWorkflow(ctx, repoID, "domain:"+domain)
		if err != nil {
			t.Fatalf("BuildReviewPreparationWorkflow failed: %v", err)
		}
		if err := ValidateWorkflow(workflow); err != nil {
			t.Fatalf("ValidateWorkflow failed: %v", err)
		}
		if len(workflow.Artifacts) != 4 {
			t.Fatalf("expected 4 artifacts, got %d", len(workflow.Artifacts))
		}

		result := decodeSearchResult(t, workflow.Artifacts[2].Data)
		if result.Query != "domain:"+domain {
			t.Fatalf("expected query %q, got %q", "domain:"+domain, result.Query)
		}
	})

	t.Run("ChangePreparation", func(t *testing.T) {
		workflow1, err := service.BuildChangePreparationWorkflow(ctx, repoID, "dec_1")
		if err != nil {
			t.Fatalf("BuildChangePreparationWorkflow failed: %v", err)
		}
		workflow2, err := service.BuildChangePreparationWorkflow(ctx, repoID, "dec_1")
		if err != nil {
			t.Fatalf("BuildChangePreparationWorkflow second call failed: %v", err)
		}
		if err := ValidateWorkflow(workflow1); err != nil {
			t.Fatalf("ValidateWorkflow failed: %v", err)
		}
		if workflow1.WorkflowID != workflow2.WorkflowID {
			t.Fatalf("expected deterministic workflow IDs, got %q and %q", workflow1.WorkflowID, workflow2.WorkflowID)
		}
		if len(workflow1.Artifacts) != 4 {
			t.Fatalf("expected 4 artifacts, got %d", len(workflow1.Artifacts))
		}
	})

	t.Run("KnowledgeExploration", func(t *testing.T) {
		workflow, err := service.BuildKnowledgeExplorationWorkflow(ctx, repoID, "SQLite")
		if err != nil {
			t.Fatalf("BuildKnowledgeExplorationWorkflow failed: %v", err)
		}
		if err := ValidateWorkflow(workflow); err != nil {
			t.Fatalf("ValidateWorkflow failed: %v", err)
		}
		if len(workflow.Artifacts) != 2 {
			t.Fatalf("expected 2 artifacts, got %d", len(workflow.Artifacts))
		}
	})
}
