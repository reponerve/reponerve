package askcmd

import (
	"strings"

	"github.com/spf13/cobra"

	"github.com/reponerve/reponerve/internal/agent/development"
	"github.com/reponerve/reponerve/internal/cli/devwire"
	"github.com/reponerve/reponerve/internal/config"
)

func newAskCommand(use, short string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   use,
		Short: short,
		Long:  `Answer repository and development questions using Code Intelligence and Repository Intelligence. Use --json for AI chat without MCP (same envelope as MCP tools).`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			question := strings.TrimSpace(args[0])
			format, err := devwire.ResolveFormat(cmd)
			if err != nil {
				return err
			}
			if format == devwire.FormatProse {
				cmd.Printf("Querying repository memory for: %q...\n", question)
			}

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

			return devwire.WriteDEResult(cmd, development.FormatAnswer(answer), answer)
		},
	}
	return cmd
}

func NewCommand() *cobra.Command {
	return devwire.BindDECmd(newAskCommand("ask [question]", "Ask a question about the repository"))
}
