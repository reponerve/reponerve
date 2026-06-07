package expertise_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"testing"
	"time"

	"reponerve/internal/ownership/expertise"
	memorymodels "reponerve/internal/memory/models"
	"reponerve/pkg/models"
)

func testContributorID(repositoryID, name, email string) string {
	var input string
	if email != "" {
		input = repositoryID + ":" + email
	} else {
		input = repositoryID + ":" + name
	}
	h := sha256.Sum256([]byte(input))
	return "ctr_" + hex.EncodeToString(h[:])
}

func TestDetector_Empty(t *testing.T) {
	d := expertise.NewDetector()
	ctx := context.Background()

	res, err := d.Detect(ctx, nil, nil, nil, nil, nil)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if len(res) != 0 {
		t.Errorf("expected empty result, got %d records", len(res))
	}
}

func TestDetector_Detect(t *testing.T) {
	d := expertise.NewDetector()
	ctx := context.Background()

	repoID := "repo_1"
	aliceName := "Alice"
	aliceEmail := "alice@example.com"
	aliceID := testContributorID(repoID, aliceName, aliceEmail)

	bobName := "Bob"
	bobEmail := "bob@example.com"
	bobID := testContributorID(repoID, bobName, bobEmail)

	contributors := []*models.Contributor{
		{
			ID:           aliceID,
			RepositoryID: repoID,
			Name:         aliceName,
			Email:        aliceEmail,
		},
		{
			ID:           bobID,
			RepositoryID: repoID,
			Name:         bobName,
			Email:        bobEmail,
		},
	}

	// Timestamps:
	// Anchor: 2026-06-02T12:00:00Z (latest commit timestamp)
	tAnchor := time.Date(2026, 6, 2, 12, 0, 0, 0, time.UTC)
	tAliceCommit := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC) // 1 day before anchor -> recent
	tBobAuthCommit := time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC) // 32 days before anchor -> not recent

	sources := []*models.Source{
		{
			ID:           "src_alice_auth",
			RepositoryID: repoID,
			SourceType:   "commit",
			Author:       "Alice <alice@example.com>",
			Title:        "feat(auth): implement token logins",
			Timestamp:    tAliceCommit,
		},
		{
			ID:           "src_bob_storage",
			RepositoryID: repoID,
			SourceType:   "commit",
			Author:       "Bob <bob@example.com>",
			Title:        "fix: configure db persistence queries",
			Timestamp:    tAnchor,
		},
		{
			ID:           "src_bob_auth",
			RepositoryID: repoID,
			SourceType:   "commit",
			Author:       "Bob <bob@example.com>",
			Title:        "chore: update session cookie token handling",
			Timestamp:    tBobAuthCommit,
		},
	}

	decisions := []*memorymodels.Decision{
		{
			ID:           "dec_1",
			RepositoryID: repoID,
			Title:        "Use JWT tokens for session auth",
			SourceID:     "src_alice_auth",
			CreatedAt:    tAliceCommit,
		},
	}

	facts := []*memorymodels.Fact{
		{
			ID:           "fact_1",
			RepositoryID: repoID,
			Subject:      "token auth",
			Predicate:    "uses",
			Object:       "symmetric jwt keys",
			SourceID:     "src_alice_auth",
			CreatedAt:    tAliceCommit,
		},
	}

	events := []*models.Event{
		{
			ID:           "ev_1",
			RepositoryID: repoID,
			EventType:    "FEATURE_INTRODUCED",
			Title:        "Storage engine initialized",
			Description:  "sqlite query database started",
			SourceID:     "src_bob_storage",
			Timestamp:    tAnchor,
		},
	}

	res, err := d.Detect(ctx, contributors, events, decisions, facts, sources)
	if err != nil {
		t.Fatalf("Detect failed: %v", err)
	}

	// We expect:
	// 1. Alice in "Authentication" domain (score 1.0, 1 commit, 1 decision, 1 fact, recent_activity: true)
	// 2. Bob in "Authentication" domain (score 1/8 = 0.125, 1 commit, recent_activity: false)
	// 3. Bob in "Storage" domain (score 1.0, 1 commit, 1 event, recent_activity: true)

	if len(res) != 3 {
		t.Fatalf("expected 3 expertise records, got %d", len(res))
	}

	// Verify records are sorted by ID deterministically
	for i := 0; i < len(res)-1; i++ {
		if res[i].ID >= res[i+1].ID {
			t.Errorf("records not sorted deterministically by ID: index %d vs %d", i, i+1)
		}
	}

	var aliceAuth, bobAuth, bobStorage *models.Expertise
	for _, e := range res {
		if e.ContributorID == aliceID && e.Domain == "Authentication" {
			aliceAuth = e
		} else if e.ContributorID == bobID && e.Domain == "Authentication" {
			bobAuth = e
		} else if e.ContributorID == bobID && e.Domain == "Storage" {
			bobStorage = e
		}
	}

	if aliceAuth == nil {
		t.Fatal("missing Alice Authentication record")
	}
	if bobAuth == nil {
		t.Fatal("missing Bob Authentication record")
	}
	if bobStorage == nil {
		t.Fatal("missing Bob Storage record")
	}

	// Validate Alice Authentication
	if aliceAuth.Score != 1.0 {
		t.Errorf("Alice Authentication score mismatch: got %f, expected 1.0", aliceAuth.Score)
	}
	var aliceAuthEvidence expertise.Evidence
	if err := json.Unmarshal([]byte(aliceAuth.EvidenceJSON), &aliceAuthEvidence); err != nil {
		t.Fatalf("failed to unmarshal Alice evidence: %v", err)
	}
	if aliceAuthEvidence.CommitCount != 1 || aliceAuthEvidence.DecisionCount != 1 || aliceAuthEvidence.FactCount != 1 || aliceAuthEvidence.EventCount != 0 {
		t.Errorf("Alice evidence count mismatch: %+v", aliceAuthEvidence)
	}
	if !aliceAuthEvidence.RecentActivity {
		t.Errorf("Alice RecentActivity should be true (activity 1 day before latest commit)")
	}

	// Validate Bob Authentication
	if bobAuth.Score != 0.125 {
		t.Errorf("Bob Authentication score mismatch: got %f, expected 0.125", bobAuth.Score)
	}
	var bobAuthEvidence expertise.Evidence
	if err := json.Unmarshal([]byte(bobAuth.EvidenceJSON), &bobAuthEvidence); err != nil {
		t.Fatalf("failed to unmarshal Bob auth evidence: %v", err)
	}
	if bobAuthEvidence.CommitCount != 1 || bobAuthEvidence.DecisionCount != 0 || bobAuthEvidence.FactCount != 0 || bobAuthEvidence.EventCount != 0 {
		t.Errorf("Bob auth evidence count mismatch: %+v", bobAuthEvidence)
	}
	if bobAuthEvidence.RecentActivity {
		t.Errorf("Bob RecentActivity should be false (activity 32 days before latest commit)")
	}

	// Validate Bob Storage
	if bobStorage.Score != 1.0 {
		t.Errorf("Bob Storage score mismatch: got %f, expected 1.0", bobStorage.Score)
	}
	var bobStorageEvidence expertise.Evidence
	if err := json.Unmarshal([]byte(bobStorage.EvidenceJSON), &bobStorageEvidence); err != nil {
		t.Fatalf("failed to unmarshal Bob storage evidence: %v", err)
	}
	if bobStorageEvidence.CommitCount != 1 || bobStorageEvidence.DecisionCount != 0 || bobStorageEvidence.FactCount != 0 || bobStorageEvidence.EventCount != 1 {
		t.Errorf("Bob storage evidence count mismatch: %+v", bobStorageEvidence)
	}
	if !bobStorageEvidence.RecentActivity {
		t.Errorf("Bob Storage RecentActivity should be true (activity on latest commit date)")
	}
}

func TestDetector_Determinism(t *testing.T) {
	d := expertise.NewDetector()
	ctx := context.Background()

	repoID := "repo_1"
	aliceName := "Alice"
	aliceEmail := "alice@example.com"
	aliceID := testContributorID(repoID, aliceName, aliceEmail)

	contributors := []*models.Contributor{
		{
			ID:           aliceID,
			RepositoryID: repoID,
			Name:         aliceName,
			Email:        aliceEmail,
		},
	}

	sources := []*models.Source{
		{
			ID:           "src_alice_auth",
			RepositoryID: repoID,
			SourceType:   "commit",
			Author:       "Alice <alice@example.com>",
			Title:        "feat(auth): implement token logins",
			Timestamp:    time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC),
		},
	}

	res1, err := d.Detect(ctx, contributors, nil, nil, nil, sources)
	if err != nil {
		t.Fatalf("first Detect failed: %v", err)
	}

	res2, err := d.Detect(ctx, contributors, nil, nil, nil, sources)
	if err != nil {
		t.Fatalf("second Detect failed: %v", err)
	}

	if len(res1) != len(res2) {
		t.Fatalf("result lengths mismatch: %d vs %d", len(res1), len(res2))
	}

	for i := range res1 {
		if res1[i].ID != res2[i].ID || res1[i].Score != res2[i].Score || res1[i].EvidenceJSON != res2[i].EvidenceJSON {
			t.Errorf("determinism violation at index %d: %+v vs %+v", i, res1[i], res2[i])
		}
	}
}
