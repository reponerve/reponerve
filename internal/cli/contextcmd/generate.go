package contextcmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/reponerve/reponerve/internal/config"
	"github.com/reponerve/reponerve/internal/context"
	"github.com/reponerve/reponerve/internal/context/render"
	"github.com/reponerve/reponerve/internal/query/storage"
	"github.com/reponerve/reponerve/internal/scanner/repository"
	"github.com/reponerve/reponerve/internal/storage/sqlite"
)

func newGenerateCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "generate",
		Short: "Generate repository context briefing",
		Long:  `Retrieve, generate, and render structured repository context briefings for users and AI coding agents.`,
		RunE: func(cmd *cobra.Command, args []string) error {
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

			// 7. Check if empty
			if len(rc.Decisions) == 0 && len(rc.Intents) == 0 && len(rc.Facts) == 0 && len(rc.Events) == 0 {
				cmd.Println("No repository context available.")
				return nil
			}

			// 8. Render and print
			renderer := render.NewRenderer()
			markdown, err := renderer.Render(rc)
			if err != nil {
				return fmt.Errorf("failed to render context: %w", err)
			}

			cmd.Print(markdown)
			return nil
		},
	}
}
