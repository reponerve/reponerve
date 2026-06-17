package onboardcmd

import (
	"context"
	"strings"

	"github.com/spf13/cobra"

	"github.com/reponerve/reponerve/internal/agent/development"
	"github.com/reponerve/reponerve/internal/cli/devwire"
)

// NewCommand creates the onboard subcommand.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "onboard [assignment]",
		Short: "First-day repository context with optional assignment plan",
		Long:  `Assemble day-one orientation, key decisions, and an optional task plan from repository and code intelligence.`,
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			assignment := ""
			if len(args) > 0 {
				assignment = strings.TrimSpace(args[0])
			}
			return devwire.RunOnboarding(cmd, assignment, func(ctx context.Context, session *devwire.Handle, topic string) (*development.DevelopmentOnboardingGuide, error) {
				return session.Service.Onboard(ctx, development.DevelopmentRequest{
					RepositoryID: session.RepositoryID,
					Topic:        topic,
				})
			})
		},
	}
	return devwire.BindDECmd(cmd)
}
