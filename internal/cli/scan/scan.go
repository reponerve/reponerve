package scancmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"reponerve/internal/config"
	"reponerve/internal/ingestion"
	memorystorage "reponerve/internal/memory/storage"
	"reponerve/internal/scanner/adr"
	"reponerve/internal/scanner/git"
	"reponerve/internal/scanner/repository"
	"reponerve/internal/storage/sqlite"
)

// NewCommand creates and returns the scan subcommand.
func NewCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "scan",
		Short: "Scan the repository to build memory",
		Long:  `Scan repository artifacts (code, git history, PRs, ADRs) to build and update repository memory.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			workspaceDir := config.GetWorkspaceDir()
			cfg, err := config.Load(workspaceDir)
			if err != nil {
				return fmt.Errorf("workspace not initialized; run 'reponerve init' first")
			}

			db, err := sqlite.Open(cfg.Storage.SQLitePath)
			if err != nil {
				return fmt.Errorf("failed to open database: %w", err)
			}
			defer db.Close()

			// 1. Initialize dependencies
			repoStore := sqlite.NewRepositoryStore(db)
			sourceStore := sqlite.NewSourceStore(db)
			scanStateStore := sqlite.NewScanStateStore(db)
			eventStore := sqlite.NewEventStore(db)
			decisionStore := memorystorage.NewSQLiteDecisionStore(db)
			intentStore := memorystorage.NewSQLiteIntentStore(db)
			factStore := memorystorage.NewSQLiteFactStore(db)

			reg := ingestion.NewRegistry()
			reg.Register("git", git.NewScanner(scanStateStore))
			reg.Register("adr", adr.NewScanner())

			pipeline := ingestion.NewPipeline(reg)
			coord := ingestion.NewCoordinator(
				repository.NewGitDiscovery(),
				repoStore,
				sourceStore,
				scanStateStore,
				eventStore,
				decisionStore,
				intentStore,
				factStore,
				pipeline,
			)

			cmd.Println("Scanning repository...")

			// 2. Call coordinator.Run()
			result, err := coord.Run(cmd.Context(), cfg.Repository.Path)
			if err != nil {
				return err
			}

			// 3. Print results
			cmd.Println("✓ Repository discovered")
			cmd.Printf("✓ %d commits indexed\n", result.CommitsIndexed)
			cmd.Printf("✓ %d ADRs indexed\n", result.ADRsIndexed)
			cmd.Println("Scan completed.")

			return nil
		},
	}
}
