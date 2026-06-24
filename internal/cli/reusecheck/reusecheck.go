package reusecheckcmd

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/reponerve/reponerve/internal/agent/development"
	"github.com/reponerve/reponerve/internal/cli/devwire"
)

// NewCommand creates the reuse-check subcommand.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reuse-check [intent]",
		Short: "Find existing symbols and patterns to reuse before writing new code",
		Long:  `Reuse Protocol — evidence-backed reuse candidates from code intelligence and repository memory.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return devwire.RunReuseCheck(cmd, args[0], func(ctx context.Context, session *devwire.Handle, intent string) (*development.ReuseCheckResult, error) {
				return session.Service.ReuseCheck(ctx, development.DevelopmentRequest{
					RepositoryID: session.RepositoryID,
					Topic:        intent,
				})
			})
		},
	}
	return devwire.BindDECmd(cmd)
}
