package extraction

import (
	"context"
	"testing"
	"time"

	"reponerve/pkg/models"
)

func TestExtractor_Extract(t *testing.T) {
	ctx := context.Background()
	extractor := NewExtractor()
	repoID := "repo_1"

	t.Run("Empty source sets", func(t *testing.T) {
		contribs, err := extractor.Extract(ctx, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(contribs) != 0 {
			t.Errorf("expected 0 contributors, got %d", len(contribs))
		}

		contribs2, err := extractor.Extract(ctx, []*models.Source{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(contribs2) != 0 {
			t.Errorf("expected 0 contributors, got %d", len(contribs2))
		}
	})

	t.Run("Filtering out non-commit sources", func(t *testing.T) {
		sources := []*models.Source{
			{
				ID:           "adr_1",
				RepositoryID: repoID,
				SourceType:   "adr",
				Author:       "Alice <alice@example.com>",
				Timestamp:    time.Now(),
			},
			{
				ID:           "doc_1",
				RepositoryID: repoID,
				SourceType:   "doc",
				Author:       "Bob <bob@example.com>",
				Timestamp:    time.Now(),
			},
		}

		contribs, err := extractor.Extract(ctx, sources)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(contribs) != 0 {
			t.Errorf("expected 0 contributors, got %d", len(contribs))
		}
	})

	t.Run("Single contributor", func(t *testing.T) {
		timestamp := time.Date(2026, 1, 10, 12, 0, 0, 0, time.UTC)
		sources := []*models.Source{
			{
				ID:           "c_1",
				RepositoryID: repoID,
				SourceType:   "commit",
				Author:       "Alice Smith <alice@example.com>",
				Timestamp:    timestamp,
			},
		}

		contribs, err := extractor.Extract(ctx, sources)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(contribs) != 1 {
			t.Fatalf("expected 1 contributor, got %d", len(contribs))
		}

		c := contribs[0]
		expectedID := contributorID(repoID, "Alice Smith", "alice@example.com")
		if c.ID != expectedID {
			t.Errorf("expected ID %q, got %q", expectedID, c.ID)
		}
		if c.RepositoryID != repoID {
			t.Errorf("expected repo ID %q, got %q", repoID, c.RepositoryID)
		}
		if c.Name != "Alice Smith" {
			t.Errorf("expected name 'Alice Smith', got %q", c.Name)
		}
		if c.Email != "alice@example.com" {
			t.Errorf("expected email 'alice@example.com', got %q", c.Email)
		}
		if !c.FirstSeen.Equal(timestamp) {
			t.Errorf("expected FirstSeen %v, got %v", timestamp, c.FirstSeen)
		}
		if !c.LastSeen.Equal(timestamp) {
			t.Errorf("expected LastSeen %v, got %v", timestamp, c.LastSeen)
		}
		if c.CommitCount != 1 {
			t.Errorf("expected CommitCount 1, got %d", c.CommitCount)
		}
	})

	t.Run("Multiple contributors and deduplication", func(t *testing.T) {
		t1 := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
		t2 := time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC)
		t3 := time.Date(2026, 1, 3, 0, 0, 0, 0, time.UTC)

		sources := []*models.Source{
			{
				ID:           "c_1",
				RepositoryID: repoID,
				SourceType:   "commit",
				Author:       "Alice <alice@example.com>",
				Timestamp:    t2,
			},
			{
				ID:           "c_2",
				RepositoryID: repoID,
				SourceType:   "commit",
				Author:       "Bob <bob@example.com>",
				Timestamp:    t3,
			},
			{
				ID:           "c_3",
				RepositoryID: repoID,
				SourceType:   "commit",
				Author:       "Alice Smith <alice@example.com>",
				Timestamp:    t1,
			},
		}

		contribs, err := extractor.Extract(ctx, sources)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Expected output contains exactly 2 contributors: Alice and Bob
		if len(contribs) != 2 {
			t.Fatalf("expected 2 contributors, got %d", len(contribs))
		}

		// Since they are sorted by ID, let's find them
		var alice, bob *models.Contributor
		for _, c := range contribs {
			if c.Email == "alice@example.com" {
				alice = c
			} else if c.Email == "bob@example.com" {
				bob = c
			}
		}

		if alice == nil || bob == nil {
			t.Fatal("could not find both contributors by email")
		}

		// Alice assertions:
		// Deduplicated to Alice Smith (longer name)
		if alice.Name != "Alice Smith" {
			t.Errorf("expected name 'Alice Smith', got %q", alice.Name)
		}
		if alice.CommitCount != 2 {
			t.Errorf("expected Alice commit count 2, got %d", alice.CommitCount)
		}
		if !alice.FirstSeen.Equal(t1) {
			t.Errorf("expected Alice FirstSeen %v, got %v", t1, alice.FirstSeen)
		}
		if !alice.LastSeen.Equal(t2) {
			t.Errorf("expected Alice LastSeen %v, got %v", t2, alice.LastSeen)
		}

		// Bob assertions:
		if bob.Name != "Bob" {
			t.Errorf("expected Bob's name to be 'Bob', got %q", bob.Name)
		}
		if bob.CommitCount != 1 {
			t.Errorf("expected Bob commit count 1, got %d", bob.CommitCount)
		}
		if !bob.FirstSeen.Equal(t3) || !bob.LastSeen.Equal(t3) {
			t.Errorf("expected Bob activity to be at %v", t3)
		}
	})

	t.Run("Deterministic IDs", func(t *testing.T) {
		id1 := contributorID("repo_a", "Alice", "alice@example.com")
		id2 := contributorID("repo_a", "Alice Smith", "alice@example.com")
		if id1 != id2 {
			t.Errorf("IDs should match for identical RepositoryID and Email despite name change: %q vs %q", id1, id2)
		}

		id3 := contributorID("repo_b", "Alice", "alice@example.com")
		if id1 == id3 {
			t.Errorf("IDs should not match for different RepositoryID: %q vs %q", id1, id3)
		}

		id4 := contributorID("repo_a", "Alice", "another@example.com")
		if id1 == id4 {
			t.Errorf("IDs should not match for different Email: %q vs %q", id1, id4)
		}
	})
}
