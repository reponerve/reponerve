package explainfilecmd

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/reponerve/reponerve/internal/agent/development"
	"github.com/reponerve/reponerve/internal/cli/devwire"
)

// NewCommand creates the explain-file subcommand.
func NewCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "explain-file [path]",
		Short: "Explain an indexed source file",
		Long:  `Resolve a file path through Code Intelligence and attach related repository context via repository-code links.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return devwire.RunExplanation(cmd, args[0], func(ctx context.Context, session *devwire.Handle, path string) (*development.DevelopmentExplanation, error) {
				return session.Service.ExplainFile(ctx, session.RepositoryID, path)
			})
		},
	}
}
