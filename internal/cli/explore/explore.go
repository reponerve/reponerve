package explorecmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/reponerve/reponerve/internal/config"
	"github.com/reponerve/reponerve/internal/graph/communities"
	graphdiscovery "github.com/reponerve/reponerve/internal/graph/discovery"
	"github.com/reponerve/reponerve/internal/graph/relationships"
	"github.com/reponerve/reponerve/internal/graph/traversal"
	"github.com/reponerve/reponerve/internal/query/storage"
	"github.com/reponerve/reponerve/internal/scanner/repository"
	"github.com/reponerve/reponerve/internal/storage/sqlite"
)

// NewCommand creates the explore subcommand.
func NewCommand() *cobra.Command {
	var outputPath string
	cmd := &cobra.Command{
		Use:   "explore",
		Short: "Export the knowledge graph as HTML",
		Long:  `Load the repository knowledge graph, detect communities, and write a self-contained HTML export.`,
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

			discoverySvc := repository.NewGitDiscovery()
			repo, err := discoverySvc.Discover(cmd.Context(), cfg.Repository.Path)
			if err != nil {
				return fmt.Errorf("failed to discover repository: %w", err)
			}

			decisionReader := storage.NewSQLiteDecisionReader(db)
			intentReader := storage.NewSQLiteIntentReader(db)
			factReader := storage.NewSQLiteFactReader(db)
			eventReader := storage.NewSQLiteEventReader(db)
			relationshipReader := storage.NewSQLiteRelationshipReader(db)
			contribReader := storage.NewSQLiteContributorReader(db)
			expertiseReader := storage.NewSQLiteExpertiseReader(db)
			sourceReader := storage.NewSQLiteSourceReader(db)

			relEngine := relationships.NewEngine(
				decisionReader, intentReader, factReader, eventReader,
				relationshipReader, contribReader, expertiseReader, sourceReader,
			)
			travEngine := traversal.NewEngine(relEngine)

			snapshot, err := travEngine.LoadGraphSnapshot(cmd.Context(), repo.ID, traversal.TraversalOptions{
				IncludeStored:  true,
				IncludeDerived: true,
			})
			if err != nil {
				return fmt.Errorf("failed to load graph: %w", err)
			}

			communityResult := communities.Detect(repo.ID, snapshot.Nodes, snapshot.Edges)
			report, err := graphdiscovery.Analyze(repo.ID, snapshot.Nodes, snapshot.Edges, communityResult)
			if err != nil {
				return fmt.Errorf("failed to analyze graph: %w", err)
			}

			payload := explorePayload{
				RepositoryID: repo.ID,
				NodeCount:    len(snapshot.Nodes),
				EdgeCount:    len(snapshot.Edges),
				Communities:  len(communityResult.Communities),
				GodNodes:     len(report.GodNodes),
				Surprises:    len(report.SurprisingConnections),
				Nodes:        make([]exploreNode, 0, len(snapshot.Nodes)),
				Edges:        make([]exploreEdge, 0, len(snapshot.Edges)),
			}
			for _, n := range snapshot.Nodes {
				payload.Nodes = append(payload.Nodes, exploreNode{
					ID: n.ID, Type: string(n.NodeType), EntityID: n.EntityID,
				})
			}
			for _, e := range snapshot.Edges {
				payload.Edges = append(payload.Edges, exploreEdge{
					ID: e.ID, From: e.FromNodeID, To: e.ToNodeID, Type: string(e.EdgeType),
				})
			}

			html, err := renderExploreHTML(payload)
			if err != nil {
				return fmt.Errorf("failed to render HTML: %w", err)
			}

			out := outputPath
			if !filepath.IsAbs(out) {
				out = filepath.Join(cfg.Repository.Path, out)
			}
			if err := os.WriteFile(out, []byte(html), 0o644); err != nil {
				return fmt.Errorf("failed to write HTML: %w", err)
			}

			cmd.Printf("✓ Graph exported to %s\n", out)
			cmd.Printf("  nodes=%d edges=%d communities=%d surprises=%d\n",
				payload.NodeCount, payload.EdgeCount, payload.Communities, payload.Surprises)
			return nil
		},
	}
	cmd.Flags().StringVarP(&outputPath, "output", "o", "reponerve-graph.html", "HTML output path")
	return cmd
}
