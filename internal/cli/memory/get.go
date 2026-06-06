package memorycmd

import (
	"github.com/spf13/cobra"
)

// newGetCommand creates and returns the get subcommand.
func newGetCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Retrieve individual repository memories by ID",
		Long:  `Retrieve details of a single memory by ID (event, decision, intent, fact).`,
	}

	// Register subcommands
	cmd.AddCommand(newEventGetCommand())
	cmd.AddCommand(newDecisionGetCommand())
	cmd.AddCommand(newIntentGetCommand())
	cmd.AddCommand(newFactGetCommand())

	return cmd
}
