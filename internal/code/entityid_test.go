package code

import "testing"

func TestEntityIDDeterministic(t *testing.T) {
	a := EntityID("repo1", "function", "internal/auth.Login")
	b := EntityID("repo1", "function", "internal/auth.Login")
	if a != b {
		t.Fatalf("expected deterministic entity ID, got %q and %q", a, b)
	}
	if a == EntityID("repo2", "function", "internal/auth.Login") {
		t.Fatal("expected different repository to produce different ID")
	}
}

func TestRelationshipIDDeterministic(t *testing.T) {
	a := RelationshipID("repo1", "CALLS", "from", "to")
	b := RelationshipID("repo1", "CALLS", "from", "to")
	if a != b {
		t.Fatalf("expected deterministic relationship ID, got %q and %q", a, b)
	}
}
