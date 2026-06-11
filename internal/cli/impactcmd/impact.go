package impactcmd

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/reponerve/reponerve/internal/config"
	"github.com/reponerve/reponerve/internal/graph/impact"
	"github.com/reponerve/reponerve/internal/graph/relationships"
	"github.com/reponerve/reponerve/internal/graph/traversal"
	"github.com/reponerve/reponerve/internal/query/storage"
	"github.com/reponerve/reponerve/internal/scanner/repository"
	"github.com/reponerve/reponerve/internal/storage/sqlite"
)

var supportedNodeTypes = map[string]bool{
	"DECISION":    true,
	"FACT":        true,
	"EVENT":       true,
	"CONTRIBUTOR": true,
}

var idPrefixToNodeType = map[string]string{
	"dec_":  "DECISION",
	"fact_": "FACT",
	"evt_":  "EVENT",
	"ctr_":  "CONTRIBUTOR",
}

// NewCommand creates the graph impact subcommand.
func NewCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "impact <NODE_TYPE> <NODE_ID>",
		Short: "Analyze graph impact for a repository entity",
		Long: `Analyze downstream graph impact for a repository memory entity using the knowledge graph.

Supported NODE_TYPE values: DECISION, FACT, EVENT, CONTRIBUTOR.

This command uses graph traversal impact analysis (canonical for structural impact).
Memory-relationship impact used by ask is a separate analysis path.`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			nodeType := strings.ToUpper(strings.TrimSpace(args[0]))
			nodeID := strings.TrimSpace(args[1])
			if !supportedNodeTypes[nodeType] {
				return fmt.Errorf("unsupported node type %q (use DECISION, FACT, EVENT, or CONTRIBUTOR)", args[0])
			}

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

			repoDiscovery := repository.NewGitDiscovery()
			repo, err := repoDiscovery.Discover(cmd.Context(), cfg.Repository.Path)
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
			impactSvc := impact.NewService(traversal.NewEngine(relEngine))

			if hint := idTypeHint(nodeType, nodeID); hint != "" {
				cmd.Println(hint)
			}
			if err := verifyEntityExists(cmd.Context(), nodeType, nodeID, repo.ID, decisionReader, factReader, eventReader, contributorReader); err != nil {
				return err
			}

			var report *impact.ImpactReport
			switch nodeType {
			case "DECISION":
				report, err = impactSvc.AnalyzeDecisionImpact(cmd.Context(), repo.ID, nodeID)
			case "FACT":
				report, err = impactSvc.AnalyzeFactImpact(cmd.Context(), repo.ID, nodeID)
			case "EVENT":
				report, err = impactSvc.AnalyzeEventImpact(cmd.Context(), repo.ID, nodeID)
			case "CONTRIBUTOR":
				report, err = impactSvc.AnalyzeContributorImpact(cmd.Context(), repo.ID, nodeID)
			}
			if err != nil {
				return err
			}

			if report == nil || len(report.ImpactPaths) == 0 {
				cmd.Printf("No impact paths found for %s %s.\n", nodeType, nodeID)
				cmd.Println("The entity exists but has no downstream graph edges.")
				cmd.Println("Graph impact needs linked decisions, facts, intents, expertise, or stored relationships from scan.")
				return nil
			}

			cmd.Printf("Impact analysis for %s %s (%d paths):\n", nodeType, nodeID, len(report.ImpactPaths))
			for i, p := range report.ImpactPaths {
				if p.Path == nil {
					continue
				}
				cmd.Printf("%d. nodes=%d edges=%d — %s\n", i+1, len(p.Path.Nodes), len(p.Path.Edges), p.Reason)
			}
			return nil
		},
	}
}

func idTypeHint(nodeType, nodeID string) string {
	for prefix, inferred := range idPrefixToNodeType {
		if strings.HasPrefix(nodeID, prefix) && inferred != nodeType {
			return fmt.Sprintf("Hint: ID %q looks like %s (prefix %q), but NODE_TYPE is %s. Try: reponerve impact %s %q",
				nodeID, inferred, prefix, nodeType, inferred, nodeID)
		}
	}
	return ""
}

func verifyEntityExists(
	ctx context.Context,
	nodeType, nodeID, repositoryID string,
	decisionReader storage.DecisionReader,
	factReader storage.FactReader,
	eventReader storage.EventReader,
	contributorReader storage.ContributorReader,
) error {
	var err error
	switch nodeType {
	case "DECISION":
		_, err = decisionReader.GetByID(ctx, nodeID)
	case "FACT":
		_, err = factReader.GetByID(ctx, nodeID)
	case "EVENT":
		_, err = eventReader.GetByID(ctx, nodeID)
	case "CONTRIBUTOR":
		_, err = contributorReader.GetByID(ctx, repositoryID, nodeID)
	}
	if err == nil {
		return nil
	}
	if errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("%s %q not found in repository memory; list entities with reponerve memory list %s",
			nodeType, nodeID, strings.ToLower(nodeType)+"s")
	}
	return fmt.Errorf("failed to verify %s %s: %w", nodeType, nodeID, err)
}
