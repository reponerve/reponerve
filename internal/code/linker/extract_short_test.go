package linker

import "testing"

func TestExtractShortSymbols_UniqueNameOnly(t *testing.T) {
	index := map[string][]string{
		"Service": {"internal/auth.Service"},
		"Store":   {"internal/store.Store", "internal/other.Store"},
	}
	matches := extractShortSymbols("Refactor Service login flow", "title", index)
	if len(matches) != 1 {
		t.Fatalf("expected 1 short symbol match, got %d", len(matches))
	}
	if matches[0].Value != "internal/auth.Service" {
		t.Fatalf("unexpected match: %q", matches[0].Value)
	}
}
