package development_test

import (
	"context"
	"testing"

	"github.com/reponerve/reponerve/internal/agent/development"
	"github.com/reponerve/reponerve/internal/agent/workflow"
	codemodels "github.com/reponerve/reponerve/internal/code/models"
	"github.com/reponerve/reponerve/internal/code"
	memorymodels "github.com/reponerve/reponerve/internal/memory/models"
	models "github.com/reponerve/reponerve/pkg/models"
)

func TestPlan_TopicResolution(t *testing.T) {
	repoID := "repo-1"
	decision := &memorymodels.Decision{
		ID:    "decision-oauth",
		Title: "Use JWT for OAuth login",
	}
	file := &codemodels.CodeEntity{
		ID: "file-auth", EntityType: codemodels.EntityTypeFile,
		QualifiedName: "internal/auth/service.go", FilePath: "internal/auth/service.go",
	}
	pkg := &codemodels.CodeEntity{
		ID: "pkg-auth", EntityType: codemodels.EntityTypePackage,
		QualifiedName: "internal/auth", PackagePath: "internal/auth",
	}
	exp := &models.Expertise{
		ID: "exp-oauth", RepositoryID: repoID,
		ContributorID: "bob@example.com", Domain: "authentication", Score: 38,
		EvidenceJSON: `{"domain":"authentication"}`,
	}

	codeEntityReader := &mockCodeEntityReader{entities: []*codemodels.CodeEntity{file, pkg}}
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

	out, err := svc.Plan(context.Background(), development.DevelopmentRequest{
		RepositoryID: repoID,
		Topic:        "Add OAuth login",
	})
	if err != nil {
		t.Fatalf("Plan failed: %v", err)
	}
	if out.Task != "Add OAuth login" {
		t.Fatalf("unexpected task: %q", out.Task)
	}
	if out.SuggestedWorkflow != workflow.WorkflowTypeChangePreparation {
		t.Fatalf("expected change_preparation workflow, got %q", out.SuggestedWorkflow)
	}
	if len(out.RelevantDecisions) == 0 && len(out.ImpactedAreas) == 0 {
		t.Fatalf("expected plan content from topic resolution")
	}
	if len(out.SourceServices) == 0 {
		t.Fatalf("expected source services")
	}
}
