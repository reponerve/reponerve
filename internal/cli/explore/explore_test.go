package explorecmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWriteExportFileRejectsSymlinkDestination(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "target.txt")
	if err := os.WriteFile(target, []byte("keep"), 0o644); err != nil {
		t.Fatal(err)
	}

	link := filepath.Join(dir, "reponerve-graph.html")
	if err := os.Symlink(target, link); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}

	err := writeExportFile(link, []byte("overwrite"))
	if err == nil {
		t.Fatal("expected symlink destination to be rejected")
	}
	if !strings.Contains(err.Error(), "refusing to overwrite symlink") {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "keep" {
		t.Fatalf("symlink target was overwritten: %q", got)
	}
}

func TestWriteExportFileWritesRegularDestination(t *testing.T) {
	out := filepath.Join(t.TempDir(), "reponerve-graph.html")
	if err := writeExportFile(out, []byte("<html></html>")); err != nil {
		t.Fatalf("write export: %v", err)
	}
	got, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "<html></html>" {
		t.Fatalf("got %q", got)
	}
}
