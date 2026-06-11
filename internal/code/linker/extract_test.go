package linker

import "testing"

func TestExtractGoFilePaths(t *testing.T) {
	text := "Updated internal/cli/search/search.go and internal/code/indexer/indexer.go"
	matches := extractGoFilePaths(text, "title")
	if len(matches) != 2 {
		t.Fatalf("expected 2 paths, got %d", len(matches))
	}
	if matches[0].Value != "internal/cli/search/search.go" {
		t.Fatalf("unexpected first path: %q", matches[0].Value)
	}
}

func TestExtractQualifiedSymbols(t *testing.T) {
	text := "Call internal/auth.Service.Login from handler"
	matches := extractQualifiedSymbols(text, "body", []string{
		"internal/auth.Service",
		"internal/auth.Service.Login",
	})
	if len(matches) != 1 {
		t.Fatalf("expected 1 symbol match, got %d", len(matches))
	}
	if matches[0].Value != "internal/auth.Service.Login" {
		t.Fatalf("expected longest symbol match, got %q", matches[0].Value)
	}
}
