package listfeaturescmd

import (
	"github.com/spf13/cobra"

	"github.com/reponerve/reponerve/internal/agent/development"
	"github.com/reponerve/reponerve/internal/cli/devwire"
	"github.com/reponerve/reponerve/internal/config"
)

// NewCommand creates the list-features subcommand.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list-features",
		Short: "List derived repository features",
		Long:  `List features derived from expertise domains, feature events, and decisions. Use --json for AI chat without MCP.`,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			session, err := devwire.Open(cmd.Context(), config.GetWorkspaceDir())
			if err != nil {
				return err
			}
			defer session.Close()

			out, err := session.Service.ListFeatures(cmd.Context(), session.RepositoryID)
			if err != nil {
				return err
			}
			return devwire.WriteDEResult(cmd, development.FormatFeatureList(out), out)
		},
	}
	return devwire.BindDECmd(cmd)
}
