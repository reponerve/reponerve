package integration

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInstallChatHookCLAUDE(t *testing.T) {
	root := t.TempDir()
	path, err := installChatHooks(root, false)
	if err != nil {
		t.Fatalf("installChatHooks: %v", err)
	}
	if path != "CLAUDE.md" {
		t.Fatalf("expected CLAUDE.md hook, got %q", path)
	}

	data, err := os.ReadFile(filepath.Join(root, "CLAUDE.md"))
	if err != nil {
		t.Fatalf("read CLAUDE.md: %v", err)
	}
	text := string(data)
	if !strings.Contains(text, chatHookMarker) {
		t.Fatalf("missing marker in CLAUDE.md: %s", text)
	}
	if !strings.Contains(text, "reponerve ask") {
		t.Fatalf("missing query instructions: %s", text)
	}
}
