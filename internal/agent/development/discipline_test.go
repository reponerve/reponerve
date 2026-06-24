package development_test

import (
	"context"
	"strings"
	"testing"

	"github.com/reponerve/reponerve/internal/agent/development"
	codemodels "github.com/reponerve/reponerve/internal/code/models"
	"github.com/reponerve/reponerve/internal/code"
	memorymodels "github.com/reponerve/reponerve/internal/memory/models"
	models "github.com/reponerve/reponerve/pkg/models"
)

func TestReuseCheck_ReturnsCandidatesWithDefinedIn(t *testing.T) {
	repoID := "repo-1"
	fn := &codemodels.CodeEntity{
		ID: "fn-auth", RepositoryID: repoID, EntityType: codemodels.EntityTypeFunction,
		QualifiedName: "internal/auth.Authenticate", Name: "Authenticate",
		PackagePath: "internal/auth", FilePath: "internal/auth/service.go",
		StartLine: 10, EndLine: 20,
	}
	decision := &memorymodels.Decision{
		ID: "dec-auth", RepositoryID: repoID, Title: "Use JWT for authentication",
	}

	codeEntityReader := &mockCodeEntityReader{entities: []*codemodels.CodeEntity{fn}}
	searchSvc := newTestSearchService([]*memorymodels.Decision{decision})
	codeSvc := code.NewService(codeEntityReader, &mockRelReader{}, &mockRepoCodeReader{})

	svc := development.NewService(
		codeSvc, searchSvc, codeEntityReader, &mockRelReader{}, &mockRepoCodeReader{},
		&mockDecisionReader{decisions: []*memorymodels.Decision{decision}},
		&mockFactReader{}, &mockEventReader{}, &mockExpertiseReader{},
		nil, &mockContributorReader{}, nil, "", nil, nil, nil, nil, nil, nil,
	)

	out, err := svc.ReuseCheck(context.Background(), development.DevelopmentRequest{
		RepositoryID: repoID,
		Topic:        "authentication",
	})
	if err != nil {
		t.Fatalf("ReuseCheck: %v", err)
	}
	if len(out.ReuseCandidates) == 0 {
		t.Fatalf("expected reuse candidates, got none")
	}
	found := false
	for _, c := range out.ReuseCandidates {
		if c.DefinedIn != "" && c.QualifiedName != "" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected candidate with defined_in: %+v", out.ReuseCandidates)
	}
}

func TestShipCheck_MigrationTopicProducesBlocker(t *testing.T) {
	repoID := "repo-1"
	decision := &memorymodels.Decision{
		ID: "dec-migrate", Title: "Adopt SQLite schema migration v2",
	}
	searchSvc := newTestSearchService([]*memorymodels.Decision{decision})
	codeSvc := code.NewService(&mockCodeEntityReader{}, &mockRelReader{}, &mockRepoCodeReader{})

	svc := development.NewService(
		codeSvc, searchSvc, &mockCodeEntityReader{}, &mockRelReader{}, &mockRepoCodeReader{},
		&mockDecisionReader{decisions: []*memorymodels.Decision{decision}},
		&mockFactReader{}, &mockEventReader{},
		&mockExpertiseReader{expertise: []*models.Expertise{
			{ID: "exp-db", RepositoryID: repoID, Domain: "Storage", Score: 10},
		}},
		nil, &mockContributorReader{}, nil, "", nil, nil, nil, nil, nil, nil,
	)

	out, err := svc.ShipCheck(context.Background(), development.DevelopmentRequest{
		RepositoryID: repoID,
		Topic:        "database migration",
	})
	if err != nil {
		t.Fatalf("ShipCheck: %v", err)
	}
	if len(out.ShipBlockers) == 0 {
		t.Fatalf("expected ship blockers for migration topic, got %+v", out)
	}
	if !strings.Contains(strings.ToLower(out.ShipBlockers[0].Message), "migration") {
		t.Fatalf("unexpected blocker message: %s", out.ShipBlockers[0].Message)
	}
}
