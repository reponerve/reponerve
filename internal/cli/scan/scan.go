package scancmd

import (
	"fmt"

	"github.com/spf13/cobra"

	codelinker "github.com/reponerve/reponerve/internal/code/linker"
	"github.com/reponerve/reponerve/internal/code/indexer"
	"github.com/reponerve/reponerve/internal/config"
	"github.com/reponerve/reponerve/internal/ingestion"
	"github.com/reponerve/reponerve/internal/memory/searchindex"
	memorystorage "github.com/reponerve/reponerve/internal/memory/storage"
	"github.com/reponerve/reponerve/internal/query/storage"
	"github.com/reponerve/reponerve/internal/scanner/adr"
	"github.com/reponerve/reponerve/internal/scanner/git"
	"github.com/reponerve/reponerve/internal/scanner/repository"
	"github.com/reponerve/reponerve/internal/storage/migrations"
	"github.com/reponerve/reponerve/internal/storage/sqlite"
)

// NewCommand creates and returns the scan subcommand.
func NewCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "scan",
		Short: "Scan the repository to build memory",
		Long:  `Scan repository artifacts (git history, ADRs, and Go source) to build and update repository memory and code intelligence.`,
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

			if err := migrations.RunUp(db); err != nil {
				return fmt.Errorf("failed to run database migrations: %w", err)
			}

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
			codeEntityStore := sqlite.NewSQLiteCodeEntityStore(db)
			codeRelStore := sqlite.NewSQLiteCodeRelationshipStore(db)
			codeIndexStateStore := sqlite.NewSQLiteCodeIndexStateStore(db)
			repoCodeRelStore := sqlite.NewSQLiteRepositoryCodeRelationshipStore(db)
			codeIndexer := indexer.New(codeEntityStore, codeRelStore, repoCodeRelStore, codeIndexStateStore)
			codeLinker := codelinker.New(
				storage.NewSQLiteEventReader(db),
				storage.NewSQLiteDecisionReader(db),
				storage.NewSQLiteFactReader(db),
				storage.NewSQLiteSourceReader(db),
				storage.NewSQLiteCodeEntityReader(db),
				repoCodeRelStore,
				codeIndexStateStore,
			)

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
				codeIndexer,
				codeLinker,
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
			cmd.Println("✓ Code intelligence indexed")
			cmd.Println("✓ Repository-code links updated")
			cmd.Println("✓ Search index rebuilt")
			cmd.Println("Scan completed.")

			return nil
		},
	}
}
