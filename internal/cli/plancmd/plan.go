package plancmd

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/reponerve/reponerve/internal/agent/development"
	"github.com/reponerve/reponerve/internal/cli/devwire"
)

// NewCommand creates the plan subcommand.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "plan [task]",
		Short: "Plan implementation for a development task",
		Long:  `Prepare implementation guidance by orchestrating search, code intelligence, learning paths, reviewers, and change planning.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return devwire.RunPlan(cmd, args[0], func(ctx context.Context, session *devwire.Handle, task string) (*development.DevelopmentPlan, error) {
				return session.Service.Plan(ctx, development.DevelopmentRequest{
					RepositoryID: session.RepositoryID,
					Topic:        task,
				})
			})
		},
	}
	return devwire.BindDECmd(cmd)
}
