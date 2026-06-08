package mcpcmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"reponerve/internal/config"
	"reponerve/internal/context"
	"reponerve/internal/context/render"
	"reponerve/internal/graph/impact"
	"reponerve/internal/graph/relationships"
	"reponerve/internal/graph/traversal"
	"reponerve/internal/intelligence/changeplan"
	"reponerve/internal/intelligence/discovery"
	"reponerve/internal/intelligence/learning"
	"reponerve/internal/intelligence/reviewers"
	"reponerve/internal/mcp"
	"reponerve/internal/mcp/server"
	ownershipquery "reponerve/internal/ownership/query"
	"reponerve/internal/query/storage"
	"reponerve/internal/storage/sqlite"
)

// NewCommand creates and returns the mcp subcommand.
func NewCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "mcp",
		Short: "Start the RepoNerve MCP Server",
		Long:  `Start the RepoNerve Model Context Protocol (MCP) server over standard input and output (STDIO).`,
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

			// 3. Create Query Engine readers
			eventReader := storage.NewSQLiteEventReader(db)
			decisionReader := storage.NewSQLiteDecisionReader(db)
			intentReader := storage.NewSQLiteIntentReader(db)
			factReader := storage.NewSQLiteFactReader(db)
			relationshipReader := storage.NewSQLiteRelationshipReader(db)

			// 4. Create Context Reader & Generator & Renderer
			ctxReader := context.NewMemoryContextReader(eventReader, decisionReader, intentReader, factReader)
			generator := context.NewGenerator(ctxReader)
			renderer := render.NewRenderer()

			// 5. Create Ownership Reader
			qrContrib := storage.NewSQLiteContributorReader(db)
			qrExpertise := storage.NewSQLiteExpertiseReader(db)
			qrSource := storage.NewSQLiteSourceReader(db)
			ownershipReader := ownershipquery.NewReader(qrContrib, qrExpertise, qrSource, decisionReader, factReader, eventReader)

			// 6. Create Graph Services
			relEngine := relationships.NewEngine(
				decisionReader, intentReader, factReader, eventReader,
				relationshipReader, qrContrib, qrExpertise, qrSource,
			)
			travEngine := traversal.NewEngine(relEngine)
			impactSvc := impact.NewService(travEngine)

			// 6b. Create Intelligence Services
			discoverySvc := discovery.NewService(
				decisionReader, factReader, eventReader, qrContrib, qrExpertise,
				relationshipReader, relEngine, travEngine, impactSvc,
			)
			learningSvc := learning.NewService(
				discoverySvc, decisionReader, factReader, eventReader,
				qrContrib, qrExpertise, qrSource, relEngine,
			)
			reviewerSvc := reviewers.NewService(
				discoverySvc, decisionReader, factReader, eventReader,
				qrContrib, qrExpertise, qrSource, impactSvc,
			)
			changePlanSvc := changeplan.NewService(impactSvc)

			// 7. Create MCP Service & Registry & Server
			svc := mcp.NewService(
				decisionReader, intentReader, factReader, eventReader,
				relationshipReader, generator, renderer, ownershipReader,
				travEngine, impactSvc, discoverySvc, learningSvc, reviewerSvc, changePlanSvc,
			)
			registry := mcp.NewRegistry()
			srv := server.NewServer(registry, svc, cmd.InOrStdin(), cmd.OutOrStdout())

			// 8. Start server
			return srv.Start(cmd.Context())
		},
	}
}
