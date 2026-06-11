package scancmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/reponerve/reponerve/internal/config"
	"github.com/reponerve/reponerve/internal/ingestion"
	"github.com/reponerve/reponerve/internal/memory/searchindex"
	memorystorage "github.com/reponerve/reponerve/internal/memory/storage"
	"github.com/reponerve/reponerve/internal/query/storage"
	"github.com/reponerve/reponerve/internal/scanner/adr"
	"github.com/reponerve/reponerve/internal/scanner/git"
	"github.com/reponerve/reponerve/internal/scanner/repository"
	"github.com/reponerve/reponerve/internal/storage/sqlite"
)

// NewCommand creates and returns the scan subcommand.
func NewCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "scan",
		Short: "Scan the repository to build memory",
		Long:  `Scan repository artifacts (git history and ADRs) to build and update repository memory.`,
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
			relationshipStore := memorystorage.NewSQLiteRelationshipStore(db)
			contributorStore := sqlite.NewSQLiteContributorStore(db)
			expertiseStore := sqlite.NewSQLiteExpertiseStore(db)
			memorySearchStore := sqlite.NewMemorySearchStore(db)

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
				relationshipStore,
				contributorStore,
				expertiseStore,
				pipeline,
			)

			cmd.Println("Scanning repository...")

			// 2. Call coordinator.Run()
			result, err := coord.Run(cmd.Context(), cfg.Repository.Path)
			if err != nil {
				return err
			}

			if err := searchindex.RebuildFromRepository(
				cmd.Context(),
				result.RepositoryID,
				storage.NewSQLiteEventReader(db),
				storage.NewSQLiteDecisionReader(db),
				storage.NewSQLiteFactReader(db),
				memorySearchStore,
			); err != nil {
				return fmt.Errorf("failed to rebuild search index: %w", err)
			}

			// 3. Print results
			cmd.Println("✓ Repository discovered")
			cmd.Printf("✓ %d commits indexed\n", result.CommitsIndexed)
			cmd.Printf("✓ %d ADRs indexed\n", result.ADRsIndexed)
			cmd.Println("✓ Search index rebuilt")
			cmd.Println("Scan completed.")

			return nil
		},
	}
}
