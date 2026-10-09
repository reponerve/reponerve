package development_test

import (
	"context"
	"strings"
	"testing"

	"github.com/reponerve/reponerve/internal/agent/development"
	agentsearch "github.com/reponerve/reponerve/internal/agent/search"
	"github.com/reponerve/reponerve/internal/code"
	codemodels "github.com/reponerve/reponerve/internal/code/models"
	memorymodels "github.com/reponerve/reponerve/internal/memory/models"
	models "github.com/reponerve/reponerve/pkg/models"
)

type mockCodeEntityReader struct {
	entities []*codemodels.CodeEntity
}

func (m *mockCodeEntityReader) GetByID(_ context.Context, id string) (*codemodels.CodeEntity, error) {
	for _, e := range m.entities {
		if e.ID == id {
			return e, nil
		}
	}
	return nil, context.Canceled
}

func (m *mockCodeEntityReader) ListByRepository(_ context.Context, _ string) ([]*codemodels.CodeEntity, error) {
	return m.entities, nil
}

func (m *mockCodeEntityReader) ListByFilePath(_ context.Context, _, filePath string) ([]*codemodels.CodeEntity, error) {
	var out []*codemodels.CodeEntity
	for _, e := range m.entities {
		if e.FilePath == filePath {
			out = append(out, e)
		}
	}
	return out, nil
}

func (m *mockCodeEntityReader) ListByModulePath(_ context.Context, _, _ string) ([]*codemodels.CodeEntity, error) {
	return nil, nil
}

func (m *mockCodeEntityReader) ListByEntityType(_ context.Context, _, entityType string) ([]*codemodels.CodeEntity, error) {
	var out []*codemodels.CodeEntity
	for _, e := range m.entities {
		if e.EntityType == entityType {
			out = append(out, e)
		}
	}
	return out, nil
}

func (m *mockCodeEntityReader) FindByQualifiedName(_ context.Context, _, qualifiedName string) ([]*codemodels.CodeEntity, error) {
	var out []*codemodels.CodeEntity
	for _, e := range m.entities {
		if e.QualifiedName == qualifiedName {
			out = append(out, e)
		}
	}
	return out, nil
}

type mockRelReader struct {
	inbound  map[string][]*codemodels.CodeRelationship
	outbound map[string][]*codemodels.CodeRelationship
}

func (m *mockRelReader) ListByFromEntity(_ context.Context, id string) ([]*codemodels.CodeRelationship, error) {
	if m.outbound != nil {
		return m.outbound[id], nil
	}
	return nil, nil
}
func (m *mockRelReader) ListByToEntity(_ context.Context, id string) ([]*codemodels.CodeRelationship, error) {
	if m.inbound != nil {
		return m.inbound[id], nil
	}
	return nil, nil
}
func (m *mockRelReader) ListByRepository(context.Context, string) ([]*codemodels.CodeRelationship, error) {
	return nil, nil
}

type mockRepoCodeReader struct {
	links []*codemodels.RepositoryCodeRelationship
}

func (m *mockRepoCodeReader) ListByRepositoryEntity(context.Context, string, string) ([]*codemodels.RepositoryCodeRelationship, error) {
	return nil, nil
}
func (m *mockRepoCodeReader) ListByCodeEntity(context.Context, string, string) ([]*codemodels.RepositoryCodeRelationship, error) {
	return nil, nil
}
func (m *mockRepoCodeReader) ListByRepository(context.Context, string) ([]*codemodels.RepositoryCodeRelationship, error) {
	return m.links, nil
}

type mockDecisionReader struct {
	decisions []*memorymodels.Decision
}

func (m *mockDecisionReader) GetByID(_ context.Context, id string) (*memorymodels.Decision, error) {
	for _, d := range m.decisions {
		if d.ID == id {
			return d, nil
		}
	}
	return nil, context.Canceled
}
func (m *mockDecisionReader) ListByRepository(context.Context, string) ([]*memorymodels.Decision, error) {
	return m.decisions, nil
}
func (m *mockDecisionReader) ListAll(context.Context) ([]*memorymodels.Decision, error) {
	return m.decisions, nil
}

type mockFactReader struct{}

func (m *mockFactReader) GetByID(context.Context, string) (*memorymodels.Fact, error) {
	return nil, context.Canceled
}
func (m *mockFactReader) ListByRepository(context.Context, string) ([]*memorymodels.Fact, error) {
	return nil, nil
}
func (m *mockFactReader) ListAll(context.Context) ([]*memorymodels.Fact, error) {
	return nil, nil
}

type mockEventReader struct{}

func (m *mockEventReader) GetByID(context.Context, string) (*models.Event, error) {
	return nil, context.Canceled
}
func (m *mockEventReader) ListByRepository(context.Context, string) ([]*models.Event, error) {
	return nil, nil
}
func (m *mockEventReader) ListAll(context.Context) ([]*models.Event, error) {
	return nil, nil
}

type mockExpertiseReader struct {
	expertise []*models.Expertise
}

func (m *mockExpertiseReader) ListByRepository(context.Context, string) ([]*models.Expertise, error) {
	return m.expertise, nil
}
func (m *mockExpertiseReader) ListByContributor(context.Context, string, string) ([]*models.Expertise, error) {
	return nil, nil
}

type mockContributorReader struct {
	contributors []*models.Contributor
}

func (m *mockContributorReader) GetByID(_ context.Context, _ string, id string) (*models.Contributor, error) {
	for _, c := range m.contributors {
		if c.ID == id {
			return c, nil
		}
	}
	return nil, nil
}
func (m *mockContributorReader) ListByRepository(_ context.Context, _ string) ([]*models.Contributor, error) {
	return m.contributors, nil
}

type mockRelationshipReader struct{}

func (m *mockRelationshipReader) GetByID(context.Context, string) (*memorymodels.Relationship, error) {
	return nil, context.Canceled
}
func (m *mockRelationshipReader) ListByRepository(context.Context, string) ([]*memorymodels.Relationship, error) {
	return nil, nil
}
func (m *mockRelationshipReader) ListAll(context.Context) ([]*memorymodels.Relationship, error) {
	return nil, nil
}

func newTestSearchService(decisions []*memorymodels.Decision) *agentsearch.Service {
	return agentsearch.NewService(
		&mockDecisionReader{decisions: decisions},
		&mockFactReader{},
		&mockEventReader{},
		&mockRelationshipReader{},
		&mockContributorReader{},
		&mockExpertiseReader{},
		nil,
		nil,
	)
}

func TestExplain_ResolvesCodeAndRepository(t *testing.T) {
	repoID := "repo-1"
	authStruct := &codemodels.CodeEntity{
		ID:            "code-struct-auth",
		RepositoryID:  repoID,
		EntityType:    codemodels.EntityTypeStruct,
		Name:          "Service",
		QualifiedName: "internal/auth.Service",
		FilePath:      "internal/auth/service.go",
	}
	fileEntity := &codemodels.CodeEntity{
		ID:            "code-file-auth",
		RepositoryID:  repoID,
		EntityType:    codemodels.EntityTypeFile,
		Name:          "service.go",
		QualifiedName: "internal/auth/service.go",
		FilePath:      "internal/auth/service.go",
	}
	decision := &memorymodels.Decision{
		ID:           "decision-auth",
		RepositoryID: repoID,
		Title:        "Use JWT for authentication",
	}
	link := &codemodels.RepositoryCodeRelationship{
		ID:                   "link-1",
		RepositoryID:         repoID,
		RepositoryEntityID:   decision.ID,
		RepositoryEntityType: agentsearch.EntityTypeDecision,
		CodeEntityID:         fileEntity.ID,
		CodeEntityType:       codemodels.EntityTypeFile,
		RelationshipType:     "DECISION_REFERENCES_CODE",
		EvidenceJSON:         `{"match":"internal/auth/service.go"}`,
	}

	codeEntityReader := &mockCodeEntityReader{entities: []*codemodels.CodeEntity{authStruct, fileEntity}}
	codeSvc := code.NewService(codeEntityReader, &mockRelReader{}, &mockRepoCodeReader{links: []*codemodels.RepositoryCodeRelationship{link}})

	searchSvc := newTestSearchService([]*memorymodels.Decision{decision})

	svc := development.NewService(
		codeSvc,
		searchSvc,
		codeEntityReader,
		&mockRelReader{},
		&mockRepoCodeReader{links: []*codemodels.RepositoryCodeRelationship{link}},
		&mockDecisionReader{decisions: []*memorymodels.Decision{decision}},
		&mockFactReader{},
		&mockEventReader{},
		&mockExpertiseReader{},
		nil,
		&mockContributorReader{},
		nil,
		"",
		nil, nil, nil, nil, nil, nil,
	)

	out, err := svc.Explain(context.Background(), development.DevelopmentRequest{
		RepositoryID: repoID,
		Topic:        "authentication",
	})
	if err != nil {
		t.Fatalf("Explain failed: %v", err)
	}
	if out.CodeContext == nil || len(out.CodeContext.Structs) == 0 {
		t.Fatalf("expected code structs in explanation")
	}
	if out.RepositoryContext == nil || len(out.RepositoryContext.Decisions) == 0 {
		t.Fatalf("expected repository decisions in explanation")
	}
	if len(out.RepositoryCodeLinks) == 0 {
		t.Fatalf("expected repository-code links")
	}
	if len(out.SourceServices) == 0 {
		t.Fatalf("expected source services")
	}
}

func TestAsk_OwnershipQuestion(t *testing.T) {
	repoID := "repo-1"
	exp := &models.Expertise{
		ID:            "exp-auth",
		RepositoryID:  repoID,
		ContributorID: "alice@example.com",
		Domain:        "authentication",
		Score:         42,
		EvidenceJSON:  `{"domain":"authentication","score":42}`,
	}
	decision := &memorymodels.Decision{
		ID:    "decision-auth",
		Title: "Use JWT for authentication",
	}

	searchSvc := newTestSearchService([]*memorymodels.Decision{decision})
	codeEntityReader := &mockCodeEntityReader{}
	codeSvc := code.NewService(codeEntityReader, &mockRelReader{}, &mockRepoCodeReader{})

	svc := development.NewService(
		codeSvc,
		searchSvc,
		codeEntityReader,
		&mockRelReader{},
		&mockRepoCodeReader{},
		&mockDecisionReader{decisions: []*memorymodels.Decision{decision}},
		&mockFactReader{},
		&mockEventReader{},
		&mockExpertiseReader{expertise: []*models.Expertise{exp}},
		nil,
		&mockContributorReader{},
		nil,
		"",
		nil, nil, nil, nil, nil, nil,
	)

	out, err := svc.Ask(context.Background(), development.DevelopmentRequest{
		RepositoryID: repoID,
		Topic:        "Who owns authentication?",
	})
	if err != nil {
		t.Fatalf("Ask failed: %v", err)
	}
	if out.AnswerType != "ownership" {
		t.Fatalf("expected ownership answer type, got %q", out.AnswerType)
	}
	if len(out.Related) == 0 {
		t.Fatalf("expected related entities")
	}
}

type mockSourceReader struct {
	sources []*models.Source
}

func (m *mockSourceReader) ListByRepository(_ context.Context, _ string) ([]*models.Source, error) {
	return m.sources, nil
}

func TestAsk_RepositoryOverview(t *testing.T) {
	repoID := "repo-1"
	readmeSource := &models.Source{
		ID:           "src-readme",
		RepositoryID: repoID,
		SourceType:   "architecture_doc",
		Reference:    "README.md",
		Title:        "RepoNerve",
		MetadataJSON: `{"content":"# RepoNerve\n\nRepoNerve is the intelligence layer for software understanding.\n\nIts purpose is to preserve and transfer software knowledge.\n\n## Installation\n\nRun go install."}`,
	}

	searchSvc := newTestSearchService(nil)
	codeEntityReader := &mockCodeEntityReader{}
	codeSvc := code.NewService(codeEntityReader, &mockRelReader{}, &mockRepoCodeReader{})

	svc := development.NewService(
		codeSvc,
		searchSvc,
		codeEntityReader,
		&mockRelReader{},
		&mockRepoCodeReader{},
		&mockDecisionReader{},
		&mockFactReader{},
		&mockEventReader{},
		&mockExpertiseReader{},
		nil,
		&mockContributorReader{},
		&mockSourceReader{sources: []*models.Source{readmeSource}},
		"",
		nil, nil, nil, nil, nil, nil,
	)

	queries := []string{
		"what does this repo do?",
		"what does this project do",
		"project overview",
		"tell me about this repository",
	}

	for _, q := range queries {
		out, err := svc.Ask(context.Background(), development.DevelopmentRequest{
			RepositoryID: repoID,
			Topic:        q,
		})
		if err != nil {
			t.Fatalf("Ask(%q) failed: %v", q, err)
		}
		if out.AnswerType != "overview" {
			t.Fatalf("Ask(%q): expected answerType overview, got %q", q, out.AnswerType)
		}
		if !strings.Contains(out.Summary, "RepoNerve") || !strings.Contains(out.Summary, "intelligence layer") {
			t.Fatalf("Ask(%q): unexpected summary %q", q, out.Summary)
		}
		if len(out.Evidence) == 0 {
			t.Fatalf("Ask(%q): expected evidence pointing to README", q)
		}
	}
}

func TestAsk_DecisionRanking_PrioritizesPrimaryAdoption(t *testing.T) {
	repoID := "repo-1"
	d1 := &memorymodels.Decision{
		ID:           "ADR-001",
		RepositoryID: repoID,
		Title:        "Use SQLite as Embedded Storage Engine",
		SourceID:     "src-adr-1",
	}
	d2 := &memorymodels.Decision{
		ID:           "ADR-002",
		RepositoryID: repoID,
		Title:        "Use SQLite for storage",
		SourceID:     "src-adr-2",
	}
	d4 := &memorymodels.Decision{
		ID:           "ADR-004",
		RepositoryID: repoID,
		Title:        "Use SQLite FTS5 for search",
		SourceID:     "src-adr-4",
	}

	src1 := &models.Source{
		ID:           "src-adr-1",
		RepositoryID: repoID,
		Title:        d1.Title,
		MetadataJSON: `{"content":"Context: SQLite provides zero external daemon requirements."}`,
	}

	searchSvc := newTestSearchService([]*memorymodels.Decision{d1, d2, d4})
	codeEntityReader := &mockCodeEntityReader{}
	codeSvc := code.NewService(codeEntityReader, &mockRelReader{}, &mockRepoCodeReader{})

	svc := development.NewService(
		codeSvc,
		searchSvc,
		codeEntityReader,
		&mockRelReader{},
		&mockRepoCodeReader{},
		&mockDecisionReader{decisions: []*memorymodels.Decision{d4, d2, d1}}, // arbitrary input order
		&mockFactReader{},
		&mockEventReader{},
		&mockExpertiseReader{},
		nil,
		&mockContributorReader{},
		&mockSourceReader{sources: []*models.Source{src1}},
		"",
		nil, nil, nil, nil, nil, nil,
	)

	out, err := svc.Ask(context.Background(), development.DevelopmentRequest{
		RepositoryID: repoID,
		Topic:        "why do we use sqlite?",
	})
	if err != nil {
		t.Fatalf("Ask failed: %v", err)
	}
	if out.AnswerType != "decision_rationale" {
		t.Fatalf("expected decision_rationale, got %q", out.AnswerType)
	}
	if len(out.Related) == 0 {
		t.Fatalf("expected related decisions")
	}
	// ADR-001 must be ranked #1
	if out.Related[0].EntityID != "ADR-001" {
		t.Fatalf("expected ADR-001 as first related decision, got %s (%s)", out.Related[0].EntityID, out.Related[0].Label)
	}
	if !strings.HasPrefix(strings.Split(out.Summary, "\n  - ")[1], "Use SQLite as Embedded Storage Engine") {
		t.Fatalf("expected ADR-001 as top rationale line in summary: %s", out.Summary)
	}
}

func TestExplain_ConfidenceThreshold_NoMatch(t *testing.T) {
	repoID := "repo-1"
	// Repository only has authStruct and expressionStruct (unrelated word containing express substring)
	expressionStruct := &codemodels.CodeEntity{
		ID:            "code-expr",
		RepositoryID:  repoID,
		EntityType:    codemodels.EntityTypeStruct,
		Name:          "BinaryExpression",
		QualifiedName: "internal/ast.BinaryExpression",
		FilePath:      "internal/ast/expr.go",
	}

	searchSvc := newTestSearchService(nil)
	codeEntityReader := &mockCodeEntityReader{entities: []*codemodels.CodeEntity{expressionStruct}}
	codeSvc := code.NewService(codeEntityReader, &mockRelReader{}, &mockRepoCodeReader{})

	svc := development.NewService(
		codeSvc,
		searchSvc,
		codeEntityReader,
		&mockRelReader{},
		&mockRepoCodeReader{},
		&mockDecisionReader{},
		&mockFactReader{},
		&mockEventReader{},
		&mockExpertiseReader{},
		nil,
		&mockContributorReader{},
		nil,
		"",
		nil, nil, nil, nil, nil, nil,
	)

	_, err := svc.Explain(context.Background(), development.DevelopmentRequest{
		RepositoryID: repoID,
		Topic:        "express",
	})
	if err == nil {
		t.Fatalf("expected Explain('express') to fail with no match found")
	}
	if !strings.Contains(err.Error(), "no match found for \"express\" in this repository") {
		t.Fatalf("unexpected error message: %v", err)
	}

	// Ask about express should report clean negative answer
	askAns, err := svc.Ask(context.Background(), development.DevelopmentRequest{
		RepositoryID: repoID,
		Topic:        "what is express?",
	})
	if err != nil {
		t.Fatalf("Ask failed: %v", err)
	}
	if !strings.Contains(askAns.Summary, "No matches found") {
		t.Fatalf("expected 'No matches found', got: %s", askAns.Summary)
	}
}
