package cli

import (
	"github.com/spf13/cobra"

	askcmd "github.com/reponerve/reponerve/internal/cli/ask"
	contextcmd "github.com/reponerve/reponerve/internal/cli/contextcmd"
	explaincmd "github.com/reponerve/reponerve/internal/cli/explain"
	explainfilecmd "github.com/reponerve/reponerve/internal/cli/explainfile"
	explainfunctioncmd "github.com/reponerve/reponerve/internal/cli/explainfunction"
	explaininterfacecmd "github.com/reponerve/reponerve/internal/cli/explaininterface"
	explainstructcmd "github.com/reponerve/reponerve/internal/cli/explainstruct"
	explaintypecmd "github.com/reponerve/reponerve/internal/cli/explaintype"
	impactcmd "github.com/reponerve/reponerve/internal/cli/impactcmd"
	onboardcmd "github.com/reponerve/reponerve/internal/cli/onboardcmd"
	plancmd "github.com/reponerve/reponerve/internal/cli/plancmd"
	reviewcmd "github.com/reponerve/reponerve/internal/cli/reviewcmd"
	hookcmd "github.com/reponerve/reponerve/internal/cli/hook"
	initcmd "github.com/reponerve/reponerve/internal/cli/init"
	integratecmd "github.com/reponerve/reponerve/internal/cli/integrate"
	searchcmd "github.com/reponerve/reponerve/internal/cli/search"
	mcpcmd "github.com/reponerve/reponerve/internal/cli/mcp"
	memorycmd "github.com/reponerve/reponerve/internal/cli/memory"
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
	rootCmd.AddCommand(integratecmd.NewCommand())
	rootCmd.AddCommand(scancmd.NewCommand())
	rootCmd.AddCommand(hookcmd.NewCommand())
	rootCmd.AddCommand(askcmd.NewCommand())
	rootCmd.AddCommand(searchcmd.NewCommand())
	rootCmd.AddCommand(explaincmd.NewCommand())
	rootCmd.AddCommand(explainfilecmd.NewCommand())
	rootCmd.AddCommand(explainfunctioncmd.NewCommand())
	rootCmd.AddCommand(explainstructcmd.NewCommand())
	rootCmd.AddCommand(explaininterfacecmd.NewCommand())
	rootCmd.AddCommand(explaintypecmd.NewCommand())
	rootCmd.AddCommand(plancmd.NewCommand())
	rootCmd.AddCommand(onboardcmd.NewCommand())
	rootCmd.AddCommand(reviewcmd.NewCommand())
	rootCmd.AddCommand(impactcmd.NewCommand())
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
