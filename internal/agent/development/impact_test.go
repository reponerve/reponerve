package development_test

import (
	"context"
	"testing"

	"github.com/reponerve/reponerve/internal/agent/development"
	codemodels "github.com/reponerve/reponerve/internal/code/models"
	"github.com/reponerve/reponerve/internal/code"
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
		nil, nil, nil, nil, nil,
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
