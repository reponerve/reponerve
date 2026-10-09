package cli

import (
	"github.com/spf13/cobra"

	askcmd "github.com/reponerve/reponerve/internal/cli/ask"
	contextcmd "github.com/reponerve/reponerve/internal/cli/contextcmd"
	disciplinepolicycmd "github.com/reponerve/reponerve/internal/cli/disciplinepolicy"
	doctorcmd "github.com/reponerve/reponerve/internal/cli/doctor"
	explaincmd "github.com/reponerve/reponerve/internal/cli/explain"
	explainfeaturecmd "github.com/reponerve/reponerve/internal/cli/explainfeature"
	explainfilecmd "github.com/reponerve/reponerve/internal/cli/explainfile"
	explainfunctioncmd "github.com/reponerve/reponerve/internal/cli/explainfunction"
	explaininterfacecmd "github.com/reponerve/reponerve/internal/cli/explaininterface"
	explainstructcmd "github.com/reponerve/reponerve/internal/cli/explainstruct"
	explaintypecmd "github.com/reponerve/reponerve/internal/cli/explaintype"
	explorecmd "github.com/reponerve/reponerve/internal/cli/explore"
	forgetcmd "github.com/reponerve/reponerve/internal/cli/forget"
	handoffcmd "github.com/reponerve/reponerve/internal/cli/handoff"
	hookcmd "github.com/reponerve/reponerve/internal/cli/hook"
	impactcmd "github.com/reponerve/reponerve/internal/cli/impactcmd"
	initcmd "github.com/reponerve/reponerve/internal/cli/init"
	integratecmd "github.com/reponerve/reponerve/internal/cli/integrate"
	listfeaturescmd "github.com/reponerve/reponerve/internal/cli/listfeatures"
	mcpcmd "github.com/reponerve/reponerve/internal/cli/mcp"
	memorycmd "github.com/reponerve/reponerve/internal/cli/memory"
	onboardcmd "github.com/reponerve/reponerve/internal/cli/onboardcmd"
	plancmd "github.com/reponerve/reponerve/internal/cli/plancmd"
	prcontextcmd "github.com/reponerve/reponerve/internal/cli/prcontext"
	remembercmd "github.com/reponerve/reponerve/internal/cli/remember"
	reusecheckcmd "github.com/reponerve/reponerve/internal/cli/reusecheck"
	reviewcmd "github.com/reponerve/reponerve/internal/cli/reviewcmd"
	scancmd "github.com/reponerve/reponerve/internal/cli/scan"
	searchcmd "github.com/reponerve/reponerve/internal/cli/search"
	shipcheckcmd "github.com/reponerve/reponerve/internal/cli/shipcheck"
	versioncmd "github.com/reponerve/reponerve/internal/cli/versioncmd"
	workflowcmd "github.com/reponerve/reponerve/internal/cli/workflowcmd"
	"github.com/reponerve/reponerve/internal/version"
)

// NewRootCmd creates the root command for the reponerve CLI.
func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:          "reponerve",
		Short:        "RepoNerve is a memory and context engine for software repositories",
		Long:         `RepoNerve is an open-source memory and context engine that preserves repository knowledge and generates optimized context.`,
		Version:      version.String(),
		SilenceUsage: true,
	}

	// Register subcommands
	rootCmd.AddCommand(initcmd.NewCommand())
	rootCmd.AddCommand(integratecmd.NewCommand())
	rootCmd.AddCommand(scancmd.NewCommand())
	rootCmd.AddCommand(hookcmd.NewCommand())
	rootCmd.AddCommand(askcmd.NewCommand())
	rootCmd.AddCommand(searchcmd.NewCommand())
	rootCmd.AddCommand(explorecmd.NewCommand())
	rootCmd.AddCommand(remembercmd.NewCommand())
	rootCmd.AddCommand(forgetcmd.NewCommand())
	rootCmd.AddCommand(handoffcmd.NewCommand())
	rootCmd.AddCommand(workflowcmd.NewCommand())
	rootCmd.AddCommand(explaincmd.NewCommand())
	rootCmd.AddCommand(explainfilecmd.NewCommand())
	rootCmd.AddCommand(explainfunctioncmd.NewCommand())
	rootCmd.AddCommand(explainstructcmd.NewCommand())
	rootCmd.AddCommand(explaininterfacecmd.NewCommand())
	rootCmd.AddCommand(explaintypecmd.NewCommand())
	rootCmd.AddCommand(explainfeaturecmd.NewCommand())
	rootCmd.AddCommand(listfeaturescmd.NewCommand())
	rootCmd.AddCommand(plancmd.NewCommand())
	rootCmd.AddCommand(onboardcmd.NewCommand())
	rootCmd.AddCommand(reviewcmd.NewCommand())
	rootCmd.AddCommand(reusecheckcmd.NewCommand())
	rootCmd.AddCommand(shipcheckcmd.NewCommand())
	rootCmd.AddCommand(disciplinepolicycmd.NewCommand())
	rootCmd.AddCommand(prcontextcmd.NewCommand())
	rootCmd.AddCommand(doctorcmd.NewCommand())
	rootCmd.AddCommand(impactcmd.NewCommand())
	rootCmd.AddCommand(memorycmd.NewCommand())
	rootCmd.AddCommand(contextcmd.NewCommand())
	rootCmd.AddCommand(mcpcmd.NewCommand())
	rootCmd.AddCommand(versioncmd.NewCommand())

	return rootCmd
}

// Execute runs the root command.
func Execute() error {
	rootCmd := NewRootCmd()
	return rootCmd.Execute()
}
