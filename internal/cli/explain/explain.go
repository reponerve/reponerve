package explaincmd

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/reponerve/reponerve/internal/agent/development"
	"github.com/reponerve/reponerve/internal/cli/devwire"
)

// NewCommand creates and returns the explain subcommand.
func NewCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "explain [topic]",
		Short: "Explain a repository topic with code and repository context",
		Long:  `Combine Code Intelligence and Repository Intelligence into a unified explanation with evidence and repository-code links.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return devwire.RunExplanation(cmd, args[0], func(ctx context.Context, session *devwire.Handle, topic string) (*development.DevelopmentExplanation, error) {
				return session.Service.Explain(ctx, development.DevelopmentRequest{
					RepositoryID: session.RepositoryID,
					Topic:        topic,
				})
			})
		},
	}
}
