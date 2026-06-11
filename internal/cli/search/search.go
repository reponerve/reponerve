package searchcmd

import (
	"fmt"

	"github.com/spf13/cobra"

	agentsearch "github.com/reponerve/reponerve/internal/agent/search"
	"github.com/reponerve/reponerve/internal/config"
	"github.com/reponerve/reponerve/internal/graph/impact"
	"github.com/reponerve/reponerve/internal/graph/relationships"
	"github.com/reponerve/reponerve/internal/graph/traversal"
	"github.com/reponerve/reponerve/internal/intelligence/discovery"
	"github.com/reponerve/reponerve/internal/query/storage"
	"github.com/reponerve/reponerve/internal/scanner/repository"
	"github.com/reponerve/reponerve/internal/storage/sqlite"
)

// NewCommand creates the search subcommand.
func NewCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "search [query]",
		Short: "Search repository knowledge deterministically",
		Long:  `Search decisions, facts, events, contributors, expertise, and relationships using FTS5 memory_search plus deterministic keyword matching. The search index is rebuilt on each scan.`,
		Args:  cobra.ExactArgs(1),
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

			discoverySvc := repository.NewGitDiscovery()
			repo, err := discoverySvc.Discover(cmd.Context(), cfg.Repository.Path)
			if err != nil {
				return fmt.Errorf("failed to discover repository: %w", err)
			}

			decisionReader := storage.NewSQLiteDecisionReader(db)
			factReader := storage.NewSQLiteFactReader(db)
			eventReader := storage.NewSQLiteEventReader(db)
			relationshipReader := storage.NewSQLiteRelationshipReader(db)
			contributorReader := storage.NewSQLiteContributorReader(db)
			expertiseReader := storage.NewSQLiteExpertiseReader(db)
			sourceReader := storage.NewSQLiteSourceReader(db)

			relEngine := relationships.NewEngine(
				decisionReader, storage.NewSQLiteIntentReader(db), factReader, eventReader,
				relationshipReader, contributorReader, expertiseReader, sourceReader,
			)
			travEngine := traversal.NewEngine(relEngine)
			impactSvc := impact.NewService(travEngine)
			discoveryService := discovery.NewService(
				decisionReader, factReader, eventReader, contributorReader, expertiseReader,
				relationshipReader, relEngine, travEngine, impactSvc,
			)

			memorySearchStore := sqlite.NewMemorySearchStore(db)

			searchSvc := agentsearch.NewService(
				decisionReader, factReader, eventReader, relationshipReader,
				contributorReader, expertiseReader, discoveryService, memorySearchStore,
			)

			result, err := searchSvc.Search(cmd.Context(), repo.ID, args[0])
			if err != nil {
				return err
			}

			if len(result.Hits) == 0 {
				cmd.Printf("No matches for %q.\n", args[0])
				return nil
			}

			cmd.Printf("Search results for %q (%d hits):\n", args[0], len(result.Hits))
			for _, hit := range result.Hits {
				cmd.Printf("- %s %s [%s] score=%d\n", hit.EntityType, hit.EntityID, hit.Source, hit.MatchScore)
			}
			return nil
		},
	}
}
