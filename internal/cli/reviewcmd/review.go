package reviewcmd

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/reponerve/reponerve/internal/agent/development"
	"github.com/reponerve/reponerve/internal/cli/devwire"
)

// NewCommand creates the review subcommand.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "review [topic]",
		Short: "Prepare a code review guide for a feature or area",
		Long:  `Prepare review guidance by orchestrating reviewer recommendations, ownership intelligence, repository search, and code intelligence.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return devwire.RunReview(cmd, args[0], func(ctx context.Context, session *devwire.Handle, topic string) (*development.DevelopmentReviewGuide, error) {
				return session.Service.PrepareReview(ctx, development.DevelopmentRequest{
					RepositoryID: session.RepositoryID,
					Topic:        topic,
				})
			})
		},
	}
	return devwire.BindDECmd(cmd)
}
