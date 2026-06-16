package ingestion_test

import (
	"context"
	"testing"
	"time"

	memorystorage "github.com/reponerve/reponerve/internal/memory/storage"
	"github.com/reponerve/reponerve/internal/ingestion"
	querystorage "github.com/reponerve/reponerve/internal/query/storage"
	"github.com/reponerve/reponerve/internal/storage/migrations"
	"github.com/reponerve/reponerve/internal/storage/sqlite"
	"github.com/reponerve/reponerve/pkg/models"
)

func TestRecomputeOwnershipUsesAllCommitSources(t *testing.T) {
	ctx := context.Background()
	db, err := sqlite.Open(t.TempDir() + "/test.db")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()
	if err := migrations.RunUp(db); err != nil {
		t.Fatalf("migrations: %v", err)
	}

	repoID := "repo_own"
	repoStore := sqlite.NewRepositoryStore(db)
	if err := repoStore.UpsertRepository(ctx, &models.Repository{ID: repoID, Name: "test", Path: "."}); err != nil {
		t.Fatalf("upsert repository: %v", err)
	}
	sourceStore := sqlite.NewSourceStore(db)
	contributorStore := sqlite.NewSQLiteContributorStore(db)
	expertiseStore := sqlite.NewSQLiteExpertiseStore(db)

	now := time.Now().UTC()
	sources := []*models.Source{
		{ID: "c1", RepositoryID: repoID, SourceType: "commit", Reference: "c1", Title: "auth fix", Author: "Alice <alice@example.com>", Timestamp: now.Add(-48 * time.Hour)},
		{ID: "c2", RepositoryID: repoID, SourceType: "commit", Reference: "c2", Title: "auth tests", Author: "Alice <alice@example.com>", Timestamp: now.Add(-24 * time.Hour)},
	}
	for _, src := range sources {
		if err := sourceStore.UpsertSource(ctx, src); err != nil {
			t.Fatalf("upsert source: %v", err)
		}
	}

	coord := ingestion.NewCoordinator(
		nil, nil, sourceStore, nil, nil,
		memorystorage.NewSQLiteDecisionStore(db),
		memorystorage.NewSQLiteIntentStore(db),
		memorystorage.NewSQLiteFactStore(db),
		memorystorage.NewSQLiteRelationshipStore(db),
		contributorStore, expertiseStore, nil, nil, nil,
		ingestion.WithOwnershipReaders(ingestion.OwnershipReaders{
			Sources:   querystorage.NewSQLiteSourceReader(db),
			Events:    querystorage.NewSQLiteEventReader(db),
			Decisions: querystorage.NewSQLiteDecisionReader(db),
			Facts:     querystorage.NewSQLiteFactReader(db),
		}),
	)

	// Simulate incremental scan that only sees the latest commit.
	if err := coord.RecomputeOwnership(ctx, repoID); err != nil {
		t.Fatalf("recompute ownership: %v", err)
	}

	contribs, err := querystorage.NewSQLiteContributorReader(db).ListByRepository(ctx, repoID)
	if err != nil {
		t.Fatalf("list contributors: %v", err)
	}
	if len(contribs) != 1 {
		t.Fatalf("expected 1 contributor, got %d", len(contribs))
	}
	if contribs[0].CommitCount != 2 {
		t.Fatalf("expected commit_count=2 after full recompute, got %d", contribs[0].CommitCount)
	}
}
