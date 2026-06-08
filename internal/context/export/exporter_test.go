package export

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/reponerve/reponerve/internal/context"
	"github.com/reponerve/reponerve/internal/context/render"
	memorymodels "github.com/reponerve/reponerve/internal/memory/models"
)

func TestExporter_Unit(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "reponerve-exporter-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	renderer := render.NewRenderer()
	exporter := NewExporter(renderer)
	genTime := time.Date(2026, 6, 6, 12, 0, 0, 0, time.UTC)

	t.Run("Empty context returns error", func(t *testing.T) {
		rc := &context.RepositoryContext{
			RepositoryID: "test_repo",
			GeneratedAt:  genTime,
		}

		outputPath := filepath.Join(tempDir, "empty.md")
		err := exporter.Export(rc, outputPath)
		if err == nil {
			t.Fatal("expected error for exporting empty context, got nil")
		}
		if !strings.Contains(err.Error(), "no repository context available to export") {
			t.Errorf("expected empty context error, got: %v", err)
		}
	})

	t.Run("Successful export to custom path", func(t *testing.T) {
		rc := &context.RepositoryContext{
			RepositoryID: "test_repo",
			GeneratedAt:  genTime,
			Decisions: []*memorymodels.Decision{
				{Title: "Adopt Go"},
			},
		}

		outputPath := filepath.Join(tempDir, "subfolder", "exported-context.md")
		err := exporter.Export(rc, outputPath)
		if err != nil {
			t.Fatalf("unexpected export error: %v", err)
		}

		// Verify file content
		data, err := os.ReadFile(outputPath)
		if err != nil {
			t.Fatalf("failed to read exported file: %v", err)
		}

		content := string(data)
		if !strings.Contains(content, "# Repository Context") || !strings.Contains(content, "Adopt Go") {
			t.Errorf("exported file content is incorrect:\n%s", content)
		}
	})

	t.Run("Permission failure", func(t *testing.T) {
		rc := &context.RepositoryContext{
			RepositoryID: "test_repo",
			GeneratedAt:  genTime,
			Decisions: []*memorymodels.Decision{
				{Title: "Adopt Go"},
			},
		}

		// Create a read-only directory to simulate permission failures
		readOnlyDir := filepath.Join(tempDir, "readonly")
		err := os.Mkdir(readOnlyDir, 0400) // read-only permissions
		if err != nil {
			t.Fatalf("failed to create readonly dir: %v", err)
		}

		// Cleanup permissions so defer os.RemoveAll can delete it
		defer os.Chmod(readOnlyDir, 0755)

		outputPath := filepath.Join(readOnlyDir, "denied.md")
		err = exporter.Export(rc, outputPath)
		if err == nil {
			t.Fatal("expected writing to read-only directory to fail, got nil error")
		}
	})
}
