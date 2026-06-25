package cli

import (
	"github.com/spf13/cobra"

	askcmd "github.com/reponerve/reponerve/internal/cli/ask"
	contextcmd "github.com/reponerve/reponerve/internal/cli/contextcmd"
	explorecmd "github.com/reponerve/reponerve/internal/cli/explore"
	forgetcmd "github.com/reponerve/reponerve/internal/cli/forget"
	handoffcmd "github.com/reponerve/reponerve/internal/cli/handoff"
	remembercmd "github.com/reponerve/reponerve/internal/cli/remember"
	workflowcmd "github.com/reponerve/reponerve/internal/cli/workflowcmd"
	explaincmd "github.com/reponerve/reponerve/internal/cli/explain"
	explainfilecmd "github.com/reponerve/reponerve/internal/cli/explainfile"
	explainfunctioncmd "github.com/reponerve/reponerve/internal/cli/explainfunction"
	explaininterfacecmd "github.com/reponerve/reponerve/internal/cli/explaininterface"
	explainstructcmd "github.com/reponerve/reponerve/internal/cli/explainstruct"
	explaintypecmd "github.com/reponerve/reponerve/internal/cli/explaintype"
	explainfeaturecmd "github.com/reponerve/reponerve/internal/cli/explainfeature"
	listfeaturescmd "github.com/reponerve/reponerve/internal/cli/listfeatures"
	impactcmd "github.com/reponerve/reponerve/internal/cli/impactcmd"
	onboardcmd "github.com/reponerve/reponerve/internal/cli/onboardcmd"
	plancmd "github.com/reponerve/reponerve/internal/cli/plancmd"
	reviewcmd "github.com/reponerve/reponerve/internal/cli/reviewcmd"
	reusecheckcmd "github.com/reponerve/reponerve/internal/cli/reusecheck"
	shipcheckcmd "github.com/reponerve/reponerve/internal/cli/shipcheck"
	disciplinepolicycmd "github.com/reponerve/reponerve/internal/cli/disciplinepolicy"
	prcontextcmd "github.com/reponerve/reponerve/internal/cli/prcontext"
	doctorcmd "github.com/reponerve/reponerve/internal/cli/doctor"
	hookcmd "github.com/reponerve/reponerve/internal/cli/hook"
	initcmd "github.com/reponerve/reponerve/internal/cli/init"
	integratecmd "github.com/reponerve/reponerve/internal/cli/integrate"
	searchcmd "github.com/reponerve/reponerve/internal/cli/search"
	mcpcmd "github.com/reponerve/reponerve/internal/cli/mcp"
	memorycmd "github.com/reponerve/reponerve/internal/cli/memory"
	scancmd "github.com/reponerve/reponerve/internal/cli/scan"
	versioncmd "github.com/reponerve/reponerve/internal/cli/versioncmd"
	"github.com/reponerve/reponerve/internal/version"
)

// NewRootCmd creates the root command for the reponerve CLI.
func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "reponerve",
		Short: "RepoNerve is a memory and context engine for software repositories",
		Long:  `RepoNerve is an open-source memory and context engine that preserves repository knowledge and generates optimized context.`,
		Version: version.String(),
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
