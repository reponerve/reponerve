package contextcmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"reponerve/internal/config"
	"reponerve/internal/context"
	"reponerve/internal/context/export"
	"reponerve/internal/context/render"
	"reponerve/internal/query/storage"
	"reponerve/internal/scanner/repository"
	"reponerve/internal/storage/sqlite"
)

func newExportCommand() *cobra.Command {
	var outputPath string
	var format string

	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export repository context briefing to a file",
		Long:  `Generate and export a structured repository context briefing to a file.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if format != "markdown" {
				return fmt.Errorf("unsupported format: %q. Only 'markdown' is supported", format)
			}

			// 1. Load active workspace
			workspaceDir := config.GetWorkspaceDir()
			cfg, err := config.Load(workspaceDir)
			if err != nil {
				return fmt.Errorf("workspace not initialized; run 'reponerve init' first")
			}

			// 2. Open configured storage
			db, err := sqlite.Open(cfg.Storage.SQLitePath)
			if err != nil {
				return fmt.Errorf("failed to open database: %w", err)
			}
			defer db.Close()

			// 3. Discover repository
			discovery := repository.NewGitDiscovery()
			repo, err := discovery.Discover(cmd.Context(), cfg.Repository.Path)
			if err != nil {
				return fmt.Errorf("failed to discover repository: %w", err)
			}

			// 4. Create Query Engine readers
			eventReader := storage.NewSQLiteEventReader(db)
			decisionReader := storage.NewSQLiteDecisionReader(db)
			intentReader := storage.NewSQLiteIntentReader(db)
			factReader := storage.NewSQLiteFactReader(db)

			// 5. Create Context Reader & Generator
			ctxReader := context.NewMemoryContextReader(eventReader, decisionReader, intentReader, factReader)
			generator := context.NewGenerator(ctxReader)

			// 6. Generate RepositoryContext
			rc, err := generator.Generate(cmd.Context(), repo.ID)
			if err != nil {
				return fmt.Errorf("failed to generate context: %w", err)
			}

			// 7. Instantiate Exporter and execute export
			renderer := render.NewRenderer()
			exp := export.NewExporter(renderer)
			err = exp.Export(rc, outputPath)
			if err != nil {
				return err
			}

			cmd.Printf("✓ Repository context exported to %s\n", outputPath)
			return nil
		},
	}

	cmd.Flags().StringVarP(&outputPath, "output", "o", "repository-context.md", "Destination path for the exported context file")
	cmd.Flags().StringVarP(&format, "format", "f", "markdown", "Export format (only 'markdown' is supported)")

	return cmd
}
