package codeowners

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseFile(t *testing.T) {
	content := `# Global owners
* @global-owner

# Core component
/hugolib/ @alice @bob
internal/storage/ @charlie
*.md @docs-team
`
	tmpFile := filepath.Join(t.TempDir(), "CODEOWNERS")
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	rules, err := ParseFile(tmpFile)
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}

	if len(rules) != 4 {
		t.Fatalf("expected 4 rules, got %d", len(rules))
	}

	if rules[0].Pattern != "*" || len(rules[0].Owners) != 1 || rules[0].Owners[0] != "@global-owner" {
		t.Errorf("unexpected rule 0: %+v", rules[0])
	}
	if rules[1].Pattern != "/hugolib/" || len(rules[1].Owners) != 2 {
		t.Errorf("unexpected rule 1: %+v", rules[1])
	}
}

func TestMatches(t *testing.T) {
	if !Matches("*", "any/file.go") {
		t.Errorf("expected * to match any/file.go")
	}
	if !Matches("/hugolib/", "hugolib/page.go") {
		t.Errorf("expected /hugolib/ to match hugolib/page.go")
	}
	if !Matches("internal/storage", "internal/storage/sqlite.go") {
		t.Errorf("expected internal/storage to match internal/storage/sqlite.go")
	}
	if Matches("internal/storage", "internal/code/walk.go") {
		t.Errorf("did not expect internal/storage to match internal/code/walk.go")
	}
}
