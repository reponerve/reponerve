package sqlite_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"reponerve/internal/storage/migrations"
	"reponerve/internal/storage/sqlite"
	"reponerve/pkg/models"
)

func TestOwnershipStores(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "reponerve-ownership-store-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	dbPath := filepath.Join(tempDir, "test.db")
	db, err := sqlite.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	if err := migrations.RunUp(db); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	ctx := context.Background()
	repoID := "test_repo"

	// Insert Repository (for foreign key check)
	repoStore := sqlite.NewRepositoryStore(db)
	err = repoStore.UpsertRepository(ctx, &models.Repository{
		ID:            repoID,
		Name:          "Test Repository",
		Path:          tempDir,
		DefaultBranch: "main",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	})
	if err != nil {
		t.Fatalf("failed to insert repository: %v", err)
	}

	t.Run("Test SQLiteContributorStore", func(t *testing.T) {
		store := sqlite.NewSQLiteContributorStore(db)

		c := &models.Contributor{
			ID:           "c_1",
			RepositoryID: repoID,
			Name:         "Alice",
			Email:        "alice@example.com",
			FirstSeen:    time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			LastSeen:     time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC),
			CommitCount:  10,
		}

		// Insert
		err := store.UpsertContributor(ctx, c)
		if err != nil {
			t.Fatalf("failed to insert contributor: %v", err)
		}

		// Retrieve and verify
		var gotID, gotRepoID, gotName, gotEmail string
		var gotFirstSeen, gotLastSeen time.Time
		var gotCommitCount int
		err = db.QueryRow("SELECT id, repository_id, name, email, first_seen, last_seen, commit_count FROM contributors WHERE id = ?", c.ID).
			Scan(&gotID, &gotRepoID, &gotName, &gotEmail, &gotFirstSeen, &gotLastSeen, &gotCommitCount)
		if err != nil {
			t.Fatalf("failed to retrieve contributor: %v", err)
		}

		if gotID != c.ID || gotRepoID != c.RepositoryID || gotName != c.Name || gotEmail != c.Email || gotCommitCount != c.CommitCount {
			t.Errorf("contributor mismatch. got %+v, expected %+v", gotName, c.Name)
		}
		if !gotFirstSeen.Equal(c.FirstSeen) || !gotLastSeen.Equal(c.LastSeen) {
			t.Errorf("contributor timestamp mismatch")
		}

		// Update (Idempotence)
		c.Name = "Alice Smith"
		c.CommitCount = 15
		err = store.UpsertContributor(ctx, c)
		if err != nil {
			t.Fatalf("failed to update contributor: %v", err)
		}

		err = db.QueryRow("SELECT name, commit_count FROM contributors WHERE id = ?", c.ID).Scan(&gotName, &gotCommitCount)
		if err != nil {
			t.Fatalf("failed to query contributor after update: %v", err)
		}

		if gotName != "Alice Smith" || gotCommitCount != 15 {
			t.Errorf("upsert did not update name or commit count correctly, got name: %q, commit_count: %d", gotName, gotCommitCount)
		}
	})

	t.Run("Test SQLiteExpertiseStore", func(t *testing.T) {
		store := sqlite.NewSQLiteExpertiseStore(db)

		e := &models.Expertise{
			ID:            "exp_1",
			RepositoryID:  repoID,
			ContributorID: "c_1",
			Domain:        "authentication",
			Score:         0.95,
			EvidenceJSON:  `{"commits": 5}`,
		}

		// Insert
		err := store.UpsertExpertise(ctx, e)
		if err != nil {
			t.Fatalf("failed to insert expertise: %v", err)
		}

		// Retrieve and verify
		var gotID, gotRepoID, gotContrID, gotDomain, gotEvidenceJSON string
		var gotScore float64
		err = db.QueryRow("SELECT id, repository_id, contributor_id, domain, score, evidence_json FROM expertise WHERE id = ?", e.ID).
			Scan(&gotID, &gotRepoID, &gotContrID, &gotDomain, &gotScore, &gotEvidenceJSON)
		if err != nil {
			t.Fatalf("failed to retrieve expertise: %v", err)
		}

		if gotID != e.ID || gotRepoID != e.RepositoryID || gotContrID != e.ContributorID || gotDomain != e.Domain || gotScore != e.Score || gotEvidenceJSON != e.EvidenceJSON {
			t.Errorf("expertise mismatch. got score %f, evidence %q, expected %f, evidence %q", gotScore, gotEvidenceJSON, e.Score, e.EvidenceJSON)
		}

		// Update (Idempotence)
		e.Score = 0.99
		e.EvidenceJSON = `{"commits": 6}`
		err = store.UpsertExpertise(ctx, e)
		if err != nil {
			t.Fatalf("failed to update expertise: %v", err)
		}

		err = db.QueryRow("SELECT score, evidence_json FROM expertise WHERE id = ?", e.ID).Scan(&gotScore, &gotEvidenceJSON)
		if err != nil {
			t.Fatalf("failed to query expertise after update: %v", err)
		}

		if gotScore != 0.99 || gotEvidenceJSON != e.EvidenceJSON {
			t.Errorf("upsert did not update score or evidence correctly, got score %f, evidence %q", gotScore, gotEvidenceJSON)
		}
	})
}
