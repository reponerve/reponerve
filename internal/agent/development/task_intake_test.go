package development_test

import (
	"context"
	"testing"

	"github.com/reponerve/reponerve/internal/agent/development"
	"github.com/reponerve/reponerve/internal/code"
	codemodels "github.com/reponerve/reponerve/internal/code/models"
	memorymodels "github.com/reponerve/reponerve/internal/memory/models"
)

func TestLooksLikeTaskDescription(t *testing.T) {
	if !development.LooksLikeTaskDescription("Add OAuth login") {
		t.Fatal("expected task description")
	}
	if !development.LooksLikeTaskDescription("Implement JWT refresh token rotation") {
		t.Fatal("expected implement prefix")
	}
	if development.LooksLikeTaskDescription("What is RepositoryContext?") {
		t.Fatal("question should not be task")
	}
}

func TestAsk_TaskDescriptionRoutesToPlan(t *testing.T) {
	repoID := "repo-1"
	file := &codemodels.CodeEntity{
		ID:            "file-auth",
		RepositoryID:  repoID,
		EntityType:    codemodels.EntityTypeFile,
		QualifiedName: "internal/auth/service.go",
		FilePath:      "internal/auth/service.go",
		Name:          "service.go",
	}
	authStruct := &codemodels.CodeEntity{
		ID:            "struct-auth",
		RepositoryID:  repoID,
		EntityType:    codemodels.EntityTypeStruct,
		Name:          "AuthService",
		QualifiedName: "internal/auth.AuthService",
		FilePath:      "internal/auth/service.go",
		PackagePath:   "internal/auth",
		StartLine:     20,
		EndLine:       40,
		Signature:     "Token string",
	}
	decision := &memorymodels.Decision{
		ID:    "decision-oauth",
		Title: "Use JWT for OAuth login",
	}

	codeEntityReader := &mockCodeEntityReader{entities: []*codemodels.CodeEntity{file, authStruct}}
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
		&mockExpertiseReader{},
		nil,
		&mockContributorReader{},
		nil,
		"",
		nil, nil, nil, nil, nil, nil,
	)

	out, err := svc.Ask(context.Background(), development.DevelopmentRequest{
		RepositoryID: repoID,
		Topic:        "Add OAuth login",
	})
	if err != nil {
		t.Fatalf("Ask failed: %v", err)
	}
	if out.AnswerType != "task_plan" {
		t.Fatalf("answer_type: got %q want task_plan", out.AnswerType)
	}
	if out.Plan == nil {
		t.Fatal("expected embedded plan")
	}
	if len(out.Plan.SuggestedSteps) == 0 {
		t.Fatal("expected suggested steps on plan")
	}
}

func TestPlan_IncludesEntityBriefings(t *testing.T) {
	repoID := "repo-1"
	authStruct := &codemodels.CodeEntity{
		ID:            "struct-auth",
		RepositoryID:  repoID,
		EntityType:    codemodels.EntityTypeStruct,
		Name:          "AuthService",
		QualifiedName: "internal/auth.AuthService",
		FilePath:      "internal/auth/service.go",
		PackagePath:   "internal/auth",
		StartLine:     20,
		EndLine:       40,
		Signature:     "Token string",
	}
	decision := &memorymodels.Decision{
		ID:    "decision-oauth",
		Title: "Use JWT for OAuth login",
	}

	codeEntityReader := &mockCodeEntityReader{entities: []*codemodels.CodeEntity{authStruct}}
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
		&mockExpertiseReader{},
		nil,
		&mockContributorReader{},
		nil,
		"",
		nil, nil, nil, nil, nil, nil,
	)

	out, err := svc.Plan(context.Background(), development.DevelopmentRequest{
		RepositoryID: repoID,
		Topic:        "Add OAuth login",
	})
	if err != nil {
		t.Fatalf("Plan failed: %v", err)
	}
	if len(out.SuggestedSteps) == 0 {
		t.Fatal("expected suggested steps")
	}
	if len(out.EntityBriefings) == 0 && len(out.StartingPoints) == 0 {
		t.Fatal("expected briefings or starting points for OAuth task")
	}
}
