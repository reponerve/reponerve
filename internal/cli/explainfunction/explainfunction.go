package explainfunctioncmd

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/reponerve/reponerve/internal/agent/development"
	"github.com/reponerve/reponerve/internal/cli/devwire"
)

// NewCommand creates the explain-function subcommand.
func NewCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "explain-function [symbol]",
		Short: "Explain a function or method symbol",
		Long:  `Resolve a function or method through Code Intelligence, including callers, callees, and repository-code links.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return devwire.RunExplanation(cmd, args[0], func(ctx context.Context, session *devwire.Handle, symbol string) (*development.DevelopmentExplanation, error) {
				return session.Service.ExplainFunction(ctx, session.RepositoryID, symbol)
			})
		},
	}
}
