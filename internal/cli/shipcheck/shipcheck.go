package shipcheckcmd

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/reponerve/reponerve/internal/agent/development"
	"github.com/reponerve/reponerve/internal/cli/devwire"
)

// NewCommand creates the ship-check subcommand.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ship-check [topic]",
		Short: "Assess ship readiness with blockers and advisories from repository evidence",
		Long:  `Ship Readiness — structured pre-merge checks from review, impact, and ownership intelligence.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return devwire.RunShipCheck(cmd, args[0], func(ctx context.Context, session *devwire.Handle, topic string) (*development.ShipCheckResult, error) {
				return session.Service.ShipCheck(ctx, development.DevelopmentRequest{
					RepositoryID: session.RepositoryID,
					Topic:        topic,
				})
			})
		},
	}
	return devwire.BindDECmd(cmd)
}
