package integration

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInstallProjectFiles(t *testing.T) {
	root := t.TempDir()

	result, err := Install(Options{ProjectRoot: root, GlobalSkill: false})
	if err != nil {
		t.Fatalf("Install() error = %v", err)
	}

	if len(result.Installed) < 8 {
		t.Fatalf("expected at least 8 installed paths, got %v", result.Installed)
	}

	required := []string{
		".cursor/mcp.json",
		".vscode/mcp.json",
		".continue/mcpServers/reponerve.json",
		".cursor/skills/reponerve/SKILL.md",
		".cursor/rules/reponerve.mdc",
		".cursor/rules/coding-guidelines.mdc",
		".cursor/rules/development-discipline.mdc",
	}
	for _, rel := range required {
		if _, err := os.Stat(filepath.Join(root, rel)); err != nil {
			t.Fatalf("missing %s: %v", rel, err)
		}
	}

	cursorMCP, err := os.ReadFile(filepath.Join(root, ".cursor/mcp.json"))
	if err != nil {
		t.Fatalf("read cursor mcp: %v", err)
	}
	if !strings.Contains(string(cursorMCP), `"reponerve"`) {
		t.Fatalf("cursor mcp missing reponerve server: %s", cursorMCP)
	}
}

func TestInstallIdempotentWithoutForce(t *testing.T) {
	root := t.TempDir()

	first, err := Install(Options{ProjectRoot: root})
	if err != nil {
		t.Fatalf("first Install() error = %v", err)
	}
	if len(first.Installed) == 0 {
		t.Fatal("expected first install to write files")
	}

	second, err := Install(Options{ProjectRoot: root})
	if err != nil {
		t.Fatalf("second Install() error = %v", err)
	}
	if len(second.Skipped) == 0 {
		t.Fatalf("expected skipped files on second run, got installed=%v updated=%v", second.Installed, second.Updated)
	}
}

func TestMergeCursorMCPPreservesExistingServers(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, ".cursor", "mcp.json")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	existing := `{
  "mcpServers": {
    "other": {
      "command": "other-mcp",
      "args": []
    }
  }
}
`
	if err := os.WriteFile(path, []byte(existing), 0o644); err != nil {
		t.Fatalf("write existing mcp: %v", err)
	}

	bundle, err := os.ReadFile(filepath.Join("bundle", "cursor-mcp.json"))
	if err != nil {
		t.Fatalf("read bundle: %v", err)
	}

	merged, err := mergeCursorMCP(path, bundle)
	if err != nil {
		t.Fatalf("mergeCursorMCP() error = %v", err)
	}
	mergedText := string(merged)
	if !strings.Contains(mergedText, `"other"`) {
		t.Fatalf("expected existing server preserved: %s", mergedText)
	}
	if !strings.Contains(mergedText, `"reponerve"`) {
		t.Fatalf("expected reponerve server merged: %s", mergedText)
	}
}

func TestInstallGlobalSkill(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	// installGlobalSkill uses os.UserHomeDir(), not HOME on all platforms;
	// run full Install with GlobalSkill in a project and verify via home override only on unix.
	if home == "" {
		t.Skip("temp home unavailable")
	}

	// UserHomeDir reads HOME on unix — validate via direct path construction in installGlobalSkill
	// by calling Install with GlobalSkill and checking ~/.cursor under temp HOME.
	root := t.TempDir()
	result, err := Install(Options{ProjectRoot: root, GlobalSkill: true})
	if err != nil {
		t.Fatalf("Install() error = %v", err)
	}

	foundGlobal := false
	for _, path := range append(result.Installed, result.Updated...) {
		if strings.Contains(path, "~/.cursor/skills/reponerve/SKILL.md") {
			foundGlobal = true
			break
		}
	}
	skillPath := filepath.Join(home, ".cursor", "skills", "reponerve", "SKILL.md")
	if _, err := os.Stat(skillPath); err == nil {
		foundGlobal = true
	}
	if !foundGlobal {
		t.Fatalf("expected global skill install, got result=%+v", result)
	}
}
