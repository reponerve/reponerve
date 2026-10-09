package development_test

import (
	"context"
	"testing"

	"github.com/reponerve/reponerve/internal/agent/development"
	"github.com/reponerve/reponerve/internal/code"
	codemodels "github.com/reponerve/reponerve/internal/code/models"
	memorymodels "github.com/reponerve/reponerve/internal/memory/models"
	models "github.com/reponerve/reponerve/pkg/models"
)

func TestAnalyzeImpact_TopicResolution(t *testing.T) {
	repoID := "repo-1"
	decision := &memorymodels.Decision{
		ID:    "decision-user-service",
		Title: "Adopt microservice boundaries",
	}
	pkg := &codemodels.CodeEntity{
		ID: "pkg-user", EntityType: codemodels.EntityTypePackage,
		QualifiedName: "internal/service/user", PackagePath: "internal/service/user",
	}
	exp := &models.Expertise{
		ID: "exp-user", RepositoryID: repoID,
		ContributorID: "alice@example.com", Domain: "user-service", Score: 40,
		EvidenceJSON: `{"domain":"user-service"}`,
	}

	codeEntityReader := &mockCodeEntityReader{entities: []*codemodels.CodeEntity{pkg}}
	searchSvc := newTestSearchService([]*memorymodels.Decision{decision})
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

	out, err := svc.AnalyzeImpact(context.Background(), development.DevelopmentRequest{
		RepositoryID: repoID,
		Topic:        "user-service",
	})
	if err != nil {
		t.Fatalf("AnalyzeImpact failed: %v", err)
	}
	if out.Subject != "user-service" {
		t.Fatalf("unexpected subject: %q", out.Subject)
	}
	if len(out.SourceServices) == 0 {
		t.Fatalf("expected source services")
	}
	if len(out.ImpactedDecisions) == 0 && len(out.DependentAreas) == 0 {
		t.Fatalf("expected impact content from topic resolution")
	}
}

func TestAnalyzeImpact_InboundCallersAndTransitiveDependencies(t *testing.T) {
	repoID := "repo-1"
	targetFunc := &codemodels.CodeEntity{
		ID: "fn-target", EntityType: codemodels.EntityTypeFunction,
		QualifiedName: "pkg/service.Target", Name: "Target",
	}
	caller1Func := &codemodels.CodeEntity{
		ID: "fn-caller1", EntityType: codemodels.EntityTypeFunction,
		QualifiedName: "pkg/handler.Handle", Name: "Handle",
	}
	caller2Func := &codemodels.CodeEntity{
		ID: "fn-caller2", EntityType: codemodels.EntityTypeFunction,
		QualifiedName: "pkg/main.Run", Name: "Run",
	}

	codeEntities := []*codemodels.CodeEntity{targetFunc, caller1Func, caller2Func}
	codeEntityReader := &mockCodeEntityReader{entities: codeEntities}
	searchSvc := newTestSearchService(nil)
	codeSvc := code.NewService(codeEntityReader, &mockRelReader{}, &mockRepoCodeReader{})

	relReader := &mockRelReader{
		inbound: map[string][]*codemodels.CodeRelationship{
			"fn-target": {
				{
					RelationshipType: "CALLS",
					FromEntityID:     "fn-caller1",
					ToEntityID:       "fn-target",
				},
			},
			"fn-caller1": {
				{
					RelationshipType: "CALLS",
					FromEntityID:     "fn-caller2",
					ToEntityID:       "fn-caller1",
				},
			},
		},
	}

	svc := development.NewService(
		codeSvc,
		searchSvc,
		codeEntityReader,
		relReader,
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

	out, err := svc.AnalyzeImpact(context.Background(), development.DevelopmentRequest{
		RepositoryID: repoID,
		Topic:        "Target",
	})
	if err != nil {
		t.Fatalf("AnalyzeImpact failed: %v", err)
	}

	if len(out.CodeDependencies) < 2 {
		t.Fatalf("expected at least 2 inbound dependencies, got %d", len(out.CodeDependencies))
	}

	foundCaller1 := false
	foundCaller2 := false
	for _, dep := range out.DependentAreas {
		if dep.EntityID == "fn-caller1" {
			foundCaller1 = true
		}
		if dep.EntityID == "fn-caller2" {
			foundCaller2 = true
		}
	}

	if !foundCaller1 {
		t.Errorf("expected direct caller fn-caller1 in DependentAreas")
	}
	if !foundCaller2 {
		t.Errorf("expected transitive caller fn-caller2 in DependentAreas")
	}
}

func TestAnalyzeImpact_ContributorHumanReadableName(t *testing.T) {
	repoID := "repo-1"
	exp := &models.Expertise{
		ID: "exp-1", RepositoryID: repoID,
		ContributorID: "contrib-hash-12345", Domain: "auth", Score: 50,
	}
	contrib := &models.Contributor{
		ID:           "contrib-hash-12345",
		RepositoryID: repoID,
		Name:         "Alice Smith",
		Email:        "alice@example.com",
	}

	codeEntityReader := &mockCodeEntityReader{}
	searchSvc := newTestSearchService(nil)
	codeSvc := code.NewService(codeEntityReader, &mockRelReader{}, &mockRepoCodeReader{})
	contribReader := &mockContributorReader{contributors: []*models.Contributor{contrib}}

	svc := development.NewService(
		codeSvc,
		searchSvc,
		codeEntityReader,
		&mockRelReader{},
		&mockRepoCodeReader{},
		&mockDecisionReader{},
		&mockFactReader{},
		&mockEventReader{},
		&mockExpertiseReader{expertise: []*models.Expertise{exp}},
		nil,
		contribReader,
		nil,
		"",
		nil, nil, nil, nil, nil, nil,
	)

	out, err := svc.AnalyzeImpact(context.Background(), development.DevelopmentRequest{
		RepositoryID: repoID,
		Topic:        "auth",
	})
	if err != nil {
		t.Fatalf("AnalyzeImpact failed: %v", err)
	}

	if len(out.Owners) == 0 {
		t.Fatalf("expected owners to be matched")
	}
	if out.Owners[0].Label != "Alice Smith" {
		t.Errorf("expected human name 'Alice Smith', got %q", out.Owners[0].Label)
	}
}
