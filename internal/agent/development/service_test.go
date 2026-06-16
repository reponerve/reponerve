package development_test

import (
	"context"
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

type mockRelReader struct{}

func (m *mockRelReader) ListByFromEntity(context.Context, string) ([]*codemodels.CodeRelationship, error) {
	return nil, nil
}
func (m *mockRelReader) ListByToEntity(context.Context, string) ([]*codemodels.CodeRelationship, error) {
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

type mockContributorReader struct{}

func (m *mockContributorReader) GetByID(context.Context, string, string) (*models.Contributor, error) {
	return nil, context.Canceled
}
func (m *mockContributorReader) ListByRepository(context.Context, string) ([]*models.Contributor, error) {
	return nil, nil
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
		nil, nil, nil, nil, nil,
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
		nil, nil, nil, nil, nil,
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
