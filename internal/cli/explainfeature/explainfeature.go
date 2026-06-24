package explainfeaturecmd

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/reponerve/reponerve/internal/agent/development"
	"github.com/reponerve/reponerve/internal/cli/devwire"
)

// NewCommand creates the explain-feature subcommand.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "explain-feature [name]",
		Short: "Explain a repository feature with code and memory context",
		Long:  `Explain a derived feature (domain/capability) with code, ownership, and decisions. Use --json for AI chat without MCP.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return devwire.RunExplanation(cmd, args[0], func(ctx context.Context, session *devwire.Handle, name string) (*development.DevelopmentExplanation, error) {
				return session.Service.ExplainFeature(ctx, session.RepositoryID, name)
			})
		},
	}
	return devwire.BindDECmd(cmd)
}
