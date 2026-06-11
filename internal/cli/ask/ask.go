package askcmd

import (
	"strings"

	"github.com/spf13/cobra"

	"github.com/reponerve/reponerve/internal/agent/development"
	"github.com/reponerve/reponerve/internal/cli/devwire"
	"github.com/reponerve/reponerve/internal/config"
)

// NewCommand creates and returns the ask subcommand.
func NewCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "ask [question]",
		Short: "Ask a question about the repository",
		Long:  `Answer repository and development questions using Code Intelligence and Repository Intelligence with evidence-backed output.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			question := strings.TrimSpace(args[0])
			cmd.Printf("Querying repository memory for: %q...\n", question)

			session, err := devwire.Open(cmd.Context(), config.GetWorkspaceDir())
			if err != nil {
				return err
			}
			defer session.Close()

			answer, err := session.Service.Ask(cmd.Context(), development.DevelopmentRequest{
				RepositoryID: session.RepositoryID,
				Topic:        question,
			})
			if err != nil {
				return err
			}

			cmd.Print(development.FormatAnswer(answer))
			return nil
		},
	}
}
