package development_test

import (
	"context"
	"testing"

	"github.com/reponerve/reponerve/internal/agent/development"
	codemodels "github.com/reponerve/reponerve/internal/code/models"
	memorymodels "github.com/reponerve/reponerve/internal/memory/models"
)

func TestRouter_ResolveTopic_LinksRepositoryAndCode(t *testing.T) {
	repoID := "repo-1"
	decision := &memorymodels.Decision{ID: "decision-1", Title: "authentication service"}
	file := &codemodels.CodeEntity{
		ID: "file-1", EntityType: codemodels.EntityTypeFile,
		Name: "service.go", QualifiedName: "internal/auth/service.go",
		FilePath: "internal/auth/service.go",
	}
	link := &codemodels.RepositoryCodeRelationship{
		ID: "link-1", RepositoryID: repoID,
		RepositoryEntityID: decision.ID, CodeEntityID: file.ID,
	}

	router := development.NewRouter(
		newTestSearchService([]*memorymodels.Decision{decision}),
		&mockCodeEntityReader{entities: []*codemodels.CodeEntity{file}},
		&mockRepoCodeReader{links: []*codemodels.RepositoryCodeRelationship{link}},
	)

	topic, err := router.ResolveTopic(context.Background(), repoID, "authentication")
	if err != nil {
		t.Fatalf("ResolveTopic failed: %v", err)
	}
	if len(topic.RepositoryHitIDs) == 0 {
		t.Fatalf("expected repository hits")
	}
	if len(topic.CodeEntityIDs) == 0 {
		t.Fatalf("expected code entity hits")
	}
	if len(topic.RepositoryCodeLinks) == 0 {
		t.Fatalf("expected repository-code links")
	}
}
