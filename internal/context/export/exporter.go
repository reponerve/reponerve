package export

import (
	"fmt"
	"os"
	"path/filepath"

	"reponerve/internal/context"
	"reponerve/internal/context/render"
)

type Exporter struct {
	renderer *render.Renderer
}

func NewExporter(r *render.Renderer) *Exporter {
	return &Exporter{renderer: r}
}

// Export renders the RepositoryContext and writes it to the specified output path.
func (e *Exporter) Export(rc *context.RepositoryContext, outputPath string) error {
	// Check if context is empty
	if len(rc.Decisions) == 0 && len(rc.Intents) == 0 && len(rc.Facts) == 0 && len(rc.Events) == 0 {
		return fmt.Errorf("no repository context available to export")
	}

	markdown, err := e.renderer.Render(rc)
	if err != nil {
		return fmt.Errorf("failed to render context: %w", err)
	}

	// Ensure destination directory exists
	dir := filepath.Dir(outputPath)
	if dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %q: %w", dir, err)
		}
	}

	// Write markdown safely to destination file
	err = os.WriteFile(outputPath, []byte(markdown), 0644)
	if err != nil {
		return fmt.Errorf("failed to write export file: %w", err)
	}

	return nil
}
