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

	memorymodels "github.com/reponerve/reponerve/internal/memory/models"
	"github.com/reponerve/reponerve/internal/ownership/expertise"
	ownerextraction "github.com/reponerve/reponerve/internal/ownership/extraction"
	ownershipquery "github.com/reponerve/reponerve/internal/ownership/query"
	querystorage "github.com/reponerve/reponerve/internal/query/storage"
	"github.com/reponerve/reponerve/internal/storage/migrations"
	"github.com/reponerve/reponerve/internal/storage/sqlite"
	"github.com/reponerve/reponerve/pkg/models"
)

func TestOwnershipQueryIntegration(t *testing.T) {
	// 1. Setup SQLite database
	tempDir, err := os.MkdirTemp("", "reponerve-ownership-query-integration-*")
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
	repoID := "repo_query_test"

	// Insert Repository
	_, err = db.Exec(
		"INSERT INTO repositories (id, name, path, default_branch, created_at, updated_at) VALUES (?, ?, ?, ?, datetime(), datetime())",
		repoID, "Query Test Repository", tempDir, "main",
	)
	if err != nil {
		t.Fatalf("failed to insert repository: %v", err)
	}

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
	aliceID := testContributorID(repoID, "Alice", "alice@example.com")

	// 2. Contributor Extraction: Seed Source Commits
	sources := []*models.Source{
		{
			ID:           "src_alice_auth",
			RepositoryID: repoID,
			SourceType:   "commit",
			Reference:    "commit_1",
			Title:        "feat(auth): add oauth credentials and token login handler",
			Author:       "Alice <alice@example.com>",
			Timestamp:    tAliceCommit,
		},
		{
			ID:           "src_alice_storage",
			RepositoryID: repoID,
			SourceType:   "commit",
			Reference:    "commit_2",
			Title:        "fix: sqlite database query configuration",
			Author:       "Alice <alice@example.com>",
			Timestamp:    tAnchor,
		},
	}

	for _, src := range sources {
		_, err = db.Exec(
			"INSERT INTO sources (id, repository_id, source_type, reference, title, author, timestamp, metadata_json, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, NULL, datetime())",
			src.ID, src.RepositoryID, src.SourceType, src.Reference, src.Title, src.Author, src.Timestamp,
		)
		if err != nil {
			t.Fatalf("failed to seed source %s: %v", src.ID, err)
		}
	}

	extractor := ownerextraction.NewExtractor()
	contribs, err := extractor.Extract(ctx, sources)
	if err != nil {
		t.Fatalf("failed to extract contributors: %v", err)
	}

	contribStore := sqlite.NewSQLiteContributorStore(db)
	for _, c := range contribs {
		if err := contribStore.UpsertContributor(ctx, c); err != nil {
			t.Fatalf("failed to upsert contributor %s: %v", c.ID, err)
		}
	}

	// 3. Expertise Detection: Seed Decisions, Facts, Events
	decisions := []*memorymodels.Decision{
		{
			ID:           "dec_auth",
			RepositoryID: repoID,
			Title:        "Use JWT token for credentials verification",
			SourceID:     "src_alice_auth",
			CreatedAt:    tAliceCommit,
		},
	}
	for _, dec := range decisions {
		_, err = db.Exec(
			"INSERT INTO memory_decisions (id, repository_id, title, status, source_id, created_at) VALUES (?, ?, ?, 'approved', ?, ?)",
			dec.ID, dec.RepositoryID, dec.Title, dec.SourceID, dec.CreatedAt,
		)
		if err != nil {
			t.Fatalf("failed to seed decision: %v", err)
		}
	}

	facts := []*memorymodels.Fact{
		{
			ID:           "fact_storage",
			RepositoryID: repoID,
			Subject:      "sqlite database",
			Predicate:    "stores",
			Object:       "repository intelligence data",
			SourceID:     "src_alice_storage",
			CreatedAt:    tAnchor,
		},
	}
	for _, f := range facts {
		_, err = db.Exec(
			"INSERT INTO memory_facts (id, repository_id, subject, predicate, object, source_id, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)",
			f.ID, f.RepositoryID, f.Subject, f.Predicate, f.Object, f.SourceID, f.CreatedAt,
		)
		if err != nil {
			t.Fatalf("failed to seed fact: %v", err)
		}
	}

	events := []*models.Event{
		{
			ID:           "ev_auth",
			RepositoryID: repoID,
			EventType:    "FEATURE_INTRODUCED",
			Title:        "Auth configuration completed",
			Description:  "Token security authentication configured",
			SourceID:     "src_alice_auth",
			Timestamp:    tAliceCommit,
		},
	}
	for _, e := range events {
		_, err = db.Exec(
			"INSERT INTO memory_events (id, repository_id, event_type, title, description, source_id, timestamp, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, datetime())",
			e.ID, e.RepositoryID, e.EventType, e.Title, e.Description, e.SourceID, e.Timestamp,
		)
		if err != nil {
			t.Fatalf("failed to seed event: %v", err)
		}
	}

	detector := expertise.NewDetector()
	exps, err := detector.Detect(ctx, contribs, events, decisions, facts, sources)
	if err != nil {
		t.Fatalf("failed to detect expertise: %v", err)
	}

	expStore := sqlite.NewSQLiteExpertiseStore(db)
	for _, exp := range exps {
		if err := expStore.UpsertExpertise(ctx, exp); err != nil {
			t.Fatalf("failed to upsert expertise %s: %v", exp.ID, err)
		}
	}

	// 4. Ownership Query Engine: Instantiate and run
	qrContrib := querystorage.NewSQLiteContributorReader(db)
	qrExpertise := querystorage.NewSQLiteExpertiseReader(db)
	qrSource := querystorage.NewSQLiteSourceReader(db)
	qrDecision := querystorage.NewSQLiteDecisionReader(db)
	qrFact := querystorage.NewSQLiteFactReader(db)
	qrEvent := querystorage.NewSQLiteEventReader(db)

	engine := ownershipquery.NewReader(qrContrib, qrExpertise, qrSource, qrDecision, qrFact, qrEvent)

	// Test GetContributor
	alice, err := engine.GetContributor(ctx, repoID, aliceID)
	if err != nil {
		t.Fatalf("GetContributor failed: %v", err)
	}
	if alice.Name != "Alice" || alice.Email != "alice@example.com" {
		t.Errorf("incorrect contributor details: %+v", alice)
	}

	// Test ListContributors
	allContribs, err := engine.ListContributors(ctx, repoID)
	if err != nil {
		t.Fatalf("ListContributors failed: %v", err)
	}
	if len(allContribs) != 1 || allContribs[0].ID != aliceID {
		t.Errorf("incorrect contributors list: %+v", allContribs)
	}

	// Test ListExpertise
	allExps, err := engine.ListExpertise(ctx, repoID)
	if err != nil {
		t.Fatalf("ListExpertise failed: %v", err)
	}
	// Alice should have expertise in both "Authentication" and "Storage" domains
	if len(allExps) != 2 {
		t.Fatalf("expected 2 expertise records, got %d: %+v", len(allExps), allExps)
	}

	var authExp, storageExp *models.Expertise
	for _, exp := range allExps {
		if exp.Domain == "Authentication" {
			authExp = exp
		} else if exp.Domain == "Storage" {
			storageExp = exp
		}
	}

	if authExp == nil || storageExp == nil {
		t.Fatal("missing expected domains (Authentication / Storage)")
	}

	// Verify Evidence Preservation
	var authEv expertise.Evidence
	if err := json.Unmarshal([]byte(authExp.EvidenceJSON), &authEv); err != nil {
		t.Fatalf("failed to parse auth evidence: %v", err)
	}
	if authEv.CommitCount != 1 || authEv.DecisionCount != 1 || authEv.EventCount != 1 || authEv.FactCount != 0 {
		t.Errorf("incorrect auth evidence counts: %+v", authEv)
	}

	var storageEv expertise.Evidence
	if err := json.Unmarshal([]byte(storageExp.EvidenceJSON), &storageEv); err != nil {
		t.Fatalf("failed to parse storage evidence: %v", err)
	}
	if storageEv.CommitCount != 1 || storageEv.DecisionCount != 0 || storageEv.EventCount != 0 || storageEv.FactCount != 1 {
		t.Errorf("incorrect storage evidence counts: %+v", storageEv)
	}

	// Test TraceContributor
	trace, err := engine.TraceContributor(ctx, repoID, aliceID)
	if err != nil {
		t.Fatalf("TraceContributor failed: %v", err)
	}

	if trace.Contributor.ID != aliceID {
		t.Errorf("trace contributor ID mismatch: got %s, expected %s", trace.Contributor.ID, aliceID)
	}

	if len(trace.Expertise) != 2 {
		t.Errorf("expected 2 expertise in trace, got %d", len(trace.Expertise))
	}

	if len(trace.Decisions) != 1 || trace.Decisions[0].ID != "dec_auth" {
		t.Errorf("expected decision dec_auth in trace, got %+v", trace.Decisions)
	}

	if len(trace.Facts) != 1 || trace.Facts[0].ID != "fact_storage" {
		t.Errorf("expected fact fact_storage in trace, got %+v", trace.Facts)
	}

	if len(trace.Events) != 1 || trace.Events[0].ID != "ev_auth" {
		t.Errorf("expected event ev_auth in trace, got %+v", trace.Events)
	}
}
