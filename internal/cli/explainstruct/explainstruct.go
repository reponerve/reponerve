package explainstructcmd

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/reponerve/reponerve/internal/agent/development"
	"github.com/reponerve/reponerve/internal/cli/devwire"
)

// NewCommand creates the explain-struct subcommand.
func NewCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "explain-struct [symbol]",
		Short: "Explain a struct symbol",
		Long:  `Resolve a struct through Code Intelligence and attach related repository context.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return devwire.RunExplanation(cmd, args[0], func(ctx context.Context, session *devwire.Handle, symbol string) (*development.DevelopmentExplanation, error) {
				return session.Service.ExplainStruct(ctx, session.RepositoryID, symbol)
			})
		},
	}
}
