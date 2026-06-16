package impactcmd

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/reponerve/reponerve/internal/agent/development"
	"github.com/reponerve/reponerve/internal/cli/devwire"
)

// NewCommand creates the impact subcommand.
func NewCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "impact [subject]",
		Short: "Analyze impact of a service, feature, or area",
		Long: `Analyze repository and code impact for a natural-language subject by orchestrating knowledge graph impact, agent impact analysis, code intelligence, and ownership intelligence.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return devwire.RunImpact(cmd, args[0], func(ctx context.Context, session *devwire.Handle, subject string) (*development.DevelopmentImpactReport, error) {
				return session.Service.AnalyzeImpact(ctx, development.DevelopmentRequest{
					RepositoryID: session.RepositoryID,
					Topic:        subject,
				})
			})
		},
	}
}
