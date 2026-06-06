package memorycmd

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"
)

// newTraceCommand creates and returns the trace subcommand.
func newTraceCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "trace",
		Short: "Traverse memory relationships",
		Long:  `Traverse relationships between memories in the knowledge graph.`,
	}

	// Register subcommands
	cmd.AddCommand(newDecisionTraceCommand())
	cmd.AddCommand(newEventTraceCommand())
	cmd.AddCommand(newIntentTraceCommand())

	return cmd
}

// printTreeSection writes a header and its children with tree characters to out.
func printTreeSection(out io.Writer, header string, items []string) {
	if len(items) == 0 {
		return
	}
	fmt.Fprintln(out, header)
	for i, item := range items {
		prefix := "├── "
		if i == len(items)-1 {
			prefix = "└── "
		}
		fmt.Fprintf(out, "%s%s\n", prefix, item)
	}
	fmt.Fprintln(out)
}
