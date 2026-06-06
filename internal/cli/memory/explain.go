package memorycmd

import (
	"github.com/spf13/cobra"
)

// newExplainCommand creates and returns the explain subcommand.
func newExplainCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "explain",
		Short: "Explain repository memories",
		Long:  `Provide human-readable explanations of decisions and events from the relationship graph.`,
	}

	// Register subcommands
	cmd.AddCommand(newDecisionExplainCommand())
	cmd.AddCommand(newEventExplainCommand())

	return cmd
}
