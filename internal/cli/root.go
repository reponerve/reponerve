package cli

import (
	"github.com/spf13/cobra"

	askcmd "github.com/reponerve/reponerve/internal/cli/ask"
	contextcmd "github.com/reponerve/reponerve/internal/cli/contextcmd"
	explaincmd "github.com/reponerve/reponerve/internal/cli/explain"
	initcmd "github.com/reponerve/reponerve/internal/cli/init"
	memorycmd "github.com/reponerve/reponerve/internal/cli/memory"
	mcpcmd "github.com/reponerve/reponerve/internal/cli/mcp"
	scancmd "github.com/reponerve/reponerve/internal/cli/scan"
)

// NewRootCmd creates the root command for the reponerve CLI.
func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "reponerve",
		Short: "RepoNerve is a memory and context engine for software repositories",
		Long:  `RepoNerve is an open-source memory and context engine that preserves repository knowledge and generates optimized context.`,
	}

	// Register subcommands
	rootCmd.AddCommand(initcmd.NewCommand())
	rootCmd.AddCommand(scancmd.NewCommand())
	rootCmd.AddCommand(askcmd.NewCommand())
	rootCmd.AddCommand(explaincmd.NewCommand())
	rootCmd.AddCommand(memorycmd.NewCommand())
	rootCmd.AddCommand(contextcmd.NewCommand())
	rootCmd.AddCommand(mcpcmd.NewCommand())

	return rootCmd
}

// Execute runs the root command.
func Execute() error {
	rootCmd := NewRootCmd()
	return rootCmd.Execute()
}
