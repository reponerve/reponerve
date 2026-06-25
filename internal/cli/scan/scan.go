package scancmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	codelinker "github.com/reponerve/reponerve/internal/code/linker"
	"github.com/reponerve/reponerve/internal/code/indexer"
	"github.com/reponerve/reponerve/internal/agent/discipline"
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
	var moduleFlags []string
	var changedFlag bool

	cmd := &cobra.Command{
		Use:   "scan",
		Short: "Scan the repository to build memory",
		Long:  `Scan repository artifacts (git history, ADRs, and Go source) to build and update repository memory and code intelligence.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			workspaceDir := config.GetWorkspaceDir()
			cfg, err := config.Load(workspaceDir)
			if err != nil {
				return fmt.Errorf("%s", config.FormatLoadError(workspaceDir, err))
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
			codeIndexer := indexer.New(db, codeEntityStore, codeRelStore, repoCodeRelStore, codeIndexStateStore)
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
			reg.Register("adr", adr.NewScanner(cfg.Ingestion.DocumentPaths...))

			pipeline := ingestion.NewPipeline(reg)
			coordOpts := []ingestion.CoordinatorOption{
				ingestion.WithOwnershipReaders(ingestion.OwnershipReaders{
					Sources:   storage.NewSQLiteSourceReader(db),
					Events:    storage.NewSQLiteEventReader(db),
					Decisions: storage.NewSQLiteDecisionReader(db),
					Facts:     storage.NewSQLiteFactReader(db),
				}),
			}

			moduleScope := moduleFlags
			if changedFlag {
				files, ferr := indexer.ChangedFiles(cfg.Repository.Path)
				if ferr != nil {
					return fmt.Errorf("list changed files: %w", ferr)
				}
				resolved, merr := indexer.ModulePathsForFiles(cfg.Repository.Path, files)
				if merr != nil {
					return fmt.Errorf("resolve changed modules: %w", merr)
				}
				moduleScope = resolved
			}
			if len(moduleScope) > 0 {
				coordOpts = append(coordOpts, ingestion.WithModuleScope(moduleScope))
			}

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
				coordOpts...,
			)

			cmd.Println("Scanning repository...")
			if len(moduleScope) > 0 {
				cmd.Printf("Scoped code index: %s\n", strings.Join(moduleScope, ", "))
			}

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
				storage.NewSQLiteSourceReader(db),
				memorySearchStore,
			); err != nil {
				return fmt.Errorf("failed to rebuild search index: %w", err)
			}

			codeEntities, err := storage.NewSQLiteCodeEntityReader(db).ListByRepository(cmd.Context(), result.RepositoryID)
			if err != nil {
				return fmt.Errorf("failed to list code entities for discipline policy: %w", err)
			}
			policy := discipline.Derive(cmd.Context(), discipline.DeriveInput{
				RepositoryID:   result.RepositoryID,
				RepositoryPath: cfg.Repository.Path,
				ADRsIndexed:    result.ADRsIndexed,
				CodeEntities:   codeEntities,
				DocumentPaths:  cfg.ResolvedDocumentPaths(),
			})
			if err := discipline.WritePolicy(workspaceDir, policy); err != nil {
				return fmt.Errorf("failed to write discipline policy: %w", err)
			}

			// 3. Print results
			cmd.Println("✓ Repository discovered")
			cmd.Printf("✓ %d commits indexed\n", result.CommitsIndexed)
			cmd.Printf("✓ %d ADRs indexed\n", result.ADRsIndexed)
			cmd.Println("✓ Code intelligence indexed")
			cmd.Println("✓ Repository-code links updated")
			cmd.Println("✓ Search index rebuilt")
			cmd.Println("✓ Discipline policy updated")
			cmd.Println("Scan completed.")

			return nil
		},
	}
	cmd.Flags().StringSliceVar(&moduleFlags, "modules", nil, "Go module paths to index (monorepo scoped scan)")
	cmd.Flags().BoolVar(&changedFlag, "changed", false, "Index only modules touched by git working tree changes")
	return cmd
}
