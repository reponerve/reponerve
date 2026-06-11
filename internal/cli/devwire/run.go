package devwire

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/reponerve/reponerve/internal/agent/development"
	"github.com/reponerve/reponerve/internal/config"
)

// RunExplanation executes a Development Experience explain workflow for CLI commands.
func RunExplanation(
	cmd *cobra.Command,
	arg string,
	run func(context.Context, *Handle, string) (*development.DevelopmentExplanation, error),
) error {
	session, err := Open(cmd.Context(), config.GetWorkspaceDir())
	if err != nil {
		return err
	}
	defer session.Close()

	out, err := run(cmd.Context(), session, arg)
	if err != nil {
		return err
	}

	cmd.Print(development.FormatExplanation(out))
	return nil
}
