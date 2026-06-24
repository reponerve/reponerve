package development_test

import (
	"context"
	"strings"
	"testing"

	"github.com/reponerve/reponerve/internal/agent/development"
	"github.com/reponerve/reponerve/internal/code"
	codemodels "github.com/reponerve/reponerve/internal/code/models"
)

func TestAsk_WhatIsConceptRoutesToBriefing(t *testing.T) {
	repoID := "repo-1"
	ctxStruct := &codemodels.CodeEntity{
		ID:            "ctx-struct",
		RepositoryID:  repoID,
		EntityType:    codemodels.EntityTypeStruct,
		Name:          "RepositoryContext",
		QualifiedName: "internal/context.RepositoryContext",
		FilePath:      "internal/context/models.go",
		PackagePath:   "internal/context",
		StartLine:     10,
		EndLine:       17,
		Signature:     "RepositoryID string; GeneratedAt time.Time",
	}
	devStruct := &codemodels.CodeEntity{
		ID:            "dev-struct",
		RepositoryID:  repoID,
		EntityType:    codemodels.EntityTypeStruct,
		Name:          "RepositoryContext",
		QualifiedName: "internal/agent/development.RepositoryContext",
		FilePath:      "internal/agent/development/models.go",
		PackagePath:   "internal/agent/development",
		StartLine:     62,
		EndLine:       71,
		Signature:     "Decisions []EntityRef; Facts []EntityRef",
	}

	codeEntityReader := &mockCodeEntityReader{entities: []*codemodels.CodeEntity{ctxStruct, devStruct}}
	codeSvc := code.NewService(codeEntityReader, &mockRelReader{}, &mockRepoCodeReader{})
	searchSvc := newTestSearchService(nil)

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

	out, err := svc.Ask(context.Background(), development.DevelopmentRequest{
		RepositoryID: repoID,
		Topic:        "What is RepositoryContext?",
	})
	if err != nil {
		t.Fatalf("Ask failed: %v", err)
	}
	if out.AnswerType != "concept_explanation" {
		t.Fatalf("expected concept_explanation, got %q", out.AnswerType)
	}
	if len(out.EntityBriefings) < 2 {
		t.Fatalf("expected briefings for ambiguous RepositoryContext, got %d", len(out.EntityBriefings))
	}
	if !strings.Contains(out.Summary, "internal/context.RepositoryContext") {
		t.Fatalf("expected context struct in summary, got: %s", out.Summary)
	}
	if !strings.Contains(out.Summary, "internal/agent/development.RepositoryContext") {
		t.Fatalf("expected development struct in summary, got: %s", out.Summary)
	}
	if len(out.EntityBriefings[0].Fields) == 0 {
		t.Fatalf("expected struct fields in briefing")
	}
}

func TestExplainStruct_AmbiguousReturnsBriefings(t *testing.T) {
	repoID := "repo-1"
	a := &codemodels.CodeEntity{
		ID: "a", RepositoryID: repoID, EntityType: codemodels.EntityTypeStruct,
		Name: "Service", QualifiedName: "internal/foo.Service",
		FilePath: "internal/foo/service.go", PackagePath: "internal/foo",
	}
	b := &codemodels.CodeEntity{
		ID: "b", RepositoryID: repoID, EntityType: codemodels.EntityTypeStruct,
		Name: "Service", QualifiedName: "internal/bar.Service",
		FilePath: "internal/bar/service.go", PackagePath: "internal/bar",
	}

	codeEntityReader := &mockCodeEntityReader{entities: []*codemodels.CodeEntity{a, b}}
	codeSvc := code.NewService(codeEntityReader, &mockRelReader{}, &mockRepoCodeReader{})
	searchSvc := newTestSearchService(nil)
	svc := development.NewService(
		codeSvc, searchSvc, codeEntityReader,
		&mockRelReader{}, &mockRepoCodeReader{},
		&mockDecisionReader{}, &mockFactReader{}, &mockEventReader{},
		&mockExpertiseReader{},
		nil,
		&mockContributorReader{},
		nil,
		"",
		nil, nil, nil, nil, nil, nil,
	)

	out, err := svc.ExplainStruct(context.Background(), repoID, "Service", "")
	if err != nil {
		t.Fatalf("ExplainStruct failed: %v", err)
	}
	if len(out.EntityBriefings) < 2 {
		t.Fatalf("expected compare briefings, got %d", len(out.EntityBriefings))
	}
}
