package integration

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"reponerve/internal/ownership/expertise"
	"reponerve/internal/storage/migrations"
	"reponerve/internal/storage/sqlite"
	memorymodels "reponerve/internal/memory/models"
	"reponerve/pkg/models"
)

func TestExpertiseDetectorIntegration(t *testing.T) {
	// 1. Setup temp directory and SQLite DB
	tempDir, err := os.MkdirTemp("", "reponerve-expertise-detector-integration-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	dbPath := filepath.Join(tempDir, "test.db")
	db, err := sqlite.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open sqlite db: %v", err)
	}
	defer db.Close()

	if err := migrations.RunUp(db); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	ctx := context.Background()
	repoID := "repo_expertise_test"

	// Insert test repository
	_, err = db.Exec(
		"INSERT INTO repositories (id, name, path, default_branch, created_at, updated_at) VALUES (?, ?, ?, ?, datetime(), datetime())",
		repoID, "test-repo", tempDir, "main",
	)
	if err != nil {
		t.Fatalf("failed to insert test repository: %v", err)
	}

	// 2. Setup mock data
	tAnchor := time.Date(2026, 6, 2, 12, 0, 0, 0, time.UTC)
	tAliceCommit := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)

	// Helper to calculate contributor ID
	testContributorID := func(repositoryID, name, email string) string {
		var input string
		if email != "" {
			input = repositoryID + ":" + email
		} else {
			input = repositoryID + ":" + name
		}
		hash := sha256.Sum256([]byte(input))
		return "ctr_" + hex.EncodeToString(hash[:])
	}

	alice := &models.Contributor{
		ID:           testContributorID(repoID, "Alice", "alice@example.com"),
		RepositoryID: repoID,
		Name:         "Alice",
		Email:        "alice@example.com",
		FirstSeen:    tAliceCommit,
		LastSeen:     tAliceCommit,
		CommitCount:  1,
	}

	contribStore := sqlite.NewSQLiteContributorStore(db)
	if err := contribStore.UpsertContributor(ctx, alice); err != nil {
		t.Fatalf("failed to insert contributor: %v", err)
	}

	sources := []*models.Source{
		{
			ID:           "commit_alice_auth",
			RepositoryID: repoID,
			SourceType:   "commit",
			Reference:    "commit_alice_auth",
			Title:        "feat(auth): implement oauth token login flow",
			Author:       "Alice <alice@example.com>",
			Timestamp:    tAliceCommit,
		},
		{
			ID:           "commit_anchor",
			RepositoryID: repoID,
			SourceType:   "commit",
			Reference:    "commit_anchor",
			Title:        "chore: update readme",
			Author:       "Alice <alice@example.com>",
			Timestamp:    tAnchor,
		},
	}

	decisions := []*memorymodels.Decision{
		{
			ID:           "dec_auth_choice",
			RepositoryID: repoID,
			Title:        "Support token-based oauth credentials",
			SourceID:     "commit_alice_auth",
			CreatedAt:    tAliceCommit,
		},
	}

	facts := []*memorymodels.Fact{
		{
			ID:           "fact_auth_config",
			RepositoryID: repoID,
			Subject:      "credential store",
			Predicate:    "secures",
			Object:       "jwt tokens",
			SourceID:     "commit_alice_auth",
			CreatedAt:    tAliceCommit,
		},
	}

	events := []*models.Event{
		{
			ID:           "ev_auth_release",
			RepositoryID: repoID,
			EventType:    "FEATURE_INTRODUCED",
			Title:        "Authentication system configured with token auth",
			Description:  "jwt credentials enabled for api login",
			SourceID:     "commit_alice_auth",
			Timestamp:    tAliceCommit,
		},
	}

	// 3. Run detection
	detector := expertise.NewDetector()
	results, err := detector.Detect(ctx, []*models.Contributor{alice}, events, decisions, facts, sources)
	if err != nil {
		t.Fatalf("detection failed: %v", err)
	}

	// Alice matches keywords for "Authentication" domain in:
	// - Commit (Title: feat(auth): implement oauth token login flow) -> contains "auth", "token", "login"
	// - Decision (Title: Support token-based oauth credentials) -> contains "token", "credential"
	// - Fact (Subject/Object: credential store / jwt tokens) -> contains "credential", "jwt", "token"
	// - Event (Title/Description: Authentication system configured with token auth / jwt credentials enabled for api login) -> contains "auth", "token", "jwt", "credential", "login"
	//
	// So Alice has:
	// - 1 commit match in Authentication
	// - 1 decision match in Authentication
	// - 1 fact match in Authentication
	// - 1 event match in Authentication
	//
	// Raw score = (1 * 1.0) + (1 * 5.0) + (1 * 2.0) + (1 * 3.0) = 11.0.
	// Since Alice is the only contributor, score will be normalized to 1.0.
	// Last activity: tAliceCommit. latest commit: tAnchor. Diff: 1 day <= 30 days -> RecentActivity: true.

	if len(results) == 0 {
		t.Fatalf("expected at least one expertise record, got 0")
	}

	var authRecord *models.Expertise
	for _, r := range results {
		if r.Domain == "Authentication" {
			authRecord = r
			break
		}
	}

	if authRecord == nil {
		t.Fatalf("expected expertise record in Authentication domain, but none was generated")
	}

	if authRecord.Score != 1.0 {
		t.Errorf("expected score 1.0, got %f", authRecord.Score)
	}

	// 4. Save to sqlite
	expStore := sqlite.NewSQLiteExpertiseStore(db)
	if err := expStore.UpsertExpertise(ctx, authRecord); err != nil {
		t.Fatalf("failed to persist expertise record: %v", err)
	}

	// 5. Query from DB and verify
	var gotID, gotRepoID, gotContributorID, gotDomain, gotEvidenceJSON string
	var gotScore float64
	err = db.QueryRow("SELECT id, repository_id, contributor_id, domain, score, evidence_json FROM expertise WHERE id = ?", authRecord.ID).
		Scan(&gotID, &gotRepoID, &gotContributorID, &gotDomain, &gotScore, &gotEvidenceJSON)
	if err != nil {
		t.Fatalf("failed to retrieve expertise from SQLite: %v", err)
	}

	if gotID != authRecord.ID {
		t.Errorf("ID mismatch: got %q, expected %q", gotID, authRecord.ID)
	}
	if gotRepoID != repoID {
		t.Errorf("RepositoryID mismatch: got %q, expected %q", gotRepoID, repoID)
	}
	if gotContributorID != alice.ID {
		t.Errorf("ContributorID mismatch: got %q, expected %q", gotContributorID, alice.ID)
	}
	if gotDomain != "Authentication" {
		t.Errorf("Domain mismatch: got %q, expected %q", gotDomain, "Authentication")
	}
	if gotScore != 1.0 {
		t.Errorf("Score mismatch: got %f, expected 1.0", gotScore)
	}

	var ev expertise.Evidence
	if err := json.Unmarshal([]byte(gotEvidenceJSON), &ev); err != nil {
		t.Fatalf("failed to parse retrieved evidence JSON: %v", err)
	}

	if ev.CommitCount != 1 || ev.DecisionCount != 1 || ev.FactCount != 1 || ev.EventCount != 1 {
		t.Errorf("evidence count mismatch: %+v", ev)
	}
	if !ev.RecentActivity {
		t.Errorf("RecentActivity should be true")
	}
}
