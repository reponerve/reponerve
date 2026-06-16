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

func TestPrepareReview_TopicResolution(t *testing.T) {
	repoID := "repo-1"
	decision := &memorymodels.Decision{
		ID:    "decision-metadata-ui",
		Title: "Use component-based metadata UI",
	}
	pkg := &codemodels.CodeEntity{
		ID: "pkg-metadata", EntityType: codemodels.EntityTypePackage,
		QualifiedName: "internal/ui/metadata", PackagePath: "internal/ui/metadata",
	}
	exp := &models.Expertise{
		ID: "exp-metadata", RepositoryID: repoID,
		ContributorID: "bob@example.com", Domain: "metadata", Score: 35,
		EvidenceJSON: `{"domain":"metadata"}`,
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

	out, err := svc.PrepareReview(context.Background(), development.DevelopmentRequest{
		RepositoryID: repoID,
		Topic:        "metadata panel",
	})
	if err != nil {
		t.Fatalf("PrepareReview failed: %v", err)
	}
	if out.Topic != "metadata panel" {
		t.Fatalf("unexpected topic: %q", out.Topic)
	}
	if out.SuggestedWorkflow != workflow.WorkflowTypeReviewPreparation {
		t.Fatalf("expected review_preparation workflow, got %q", out.SuggestedWorkflow)
	}
	if len(out.SourceServices) == 0 {
		t.Fatalf("expected source services")
	}
	if len(out.RelatedKnowledge) == 0 && len(out.AffectedAreas) == 0 && len(out.RequiredExpertise) == 0 {
		t.Fatalf("expected review guide content from topic resolution")
	}
}
