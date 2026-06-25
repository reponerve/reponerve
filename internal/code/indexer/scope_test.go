package indexer

import (
	"os"
	"path/filepath"
	"testing"
)

func TestModulePathsForFiles(t *testing.T) {
	repoPath := t.TempDir()
	writeTestFile(t, filepath.Join(repoPath, "go.mod"), "module example.com/root\n\ngo 1.22\n")
	sub := filepath.Join(repoPath, "sub")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	writeTestFile(t, filepath.Join(sub, "go.mod"), "module example.com/sub\n\ngo 1.22\n")
	writeTestFile(t, filepath.Join(sub, "main.go"), "package main\n")

	paths, err := ModulePathsForFiles(repoPath, []string{"sub/main.go"})
	if err != nil {
		t.Fatalf("ModulePathsForFiles: %v", err)
	}
	if len(paths) != 1 || paths[0] != "example.com/sub" {
		t.Fatalf("got %v, want [example.com/sub]", paths)
	}
}

func TestFilterModuleRoots(t *testing.T) {
	roots := []moduleRoot{
		{modulePath: "a"},
		{modulePath: "b"},
	}
	filtered, err := FilterModuleRoots(roots, []string{"b"})
	if err != nil {
		t.Fatalf("FilterModuleRoots: %v", err)
	}
	if len(filtered) != 1 || filtered[0].modulePath != "b" {
		t.Fatalf("got %+v", filtered)
	}
}

func writeTestFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
