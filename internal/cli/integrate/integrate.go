package integratecmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/reponerve/reponerve/internal/integration"
)

// NewCommand installs or refreshes IDE integration (skill + MCP configs).
func NewCommand() *cobra.Command {
	var force bool
	var skipGlobal bool

	cmd := &cobra.Command{
		Use:   "integrate",
		Short: "Install Cursor skill and MCP configs for AI chat",
		Long: `Install or update RepoNerve IDE integration in the current repository.

Writes project files:
  .cursor/mcp.json, .cursor/skills/reponerve/, .cursor/rules/reponerve.mdc
  .vscode/mcp.json, .continue/mcpServers/reponerve.json

Also installs the global Cursor skill to ~/.cursor/skills/reponerve/ unless --skip-global is set.

Run automatically by 'reponerve init'. Safe to re-run; merges MCP entries without removing other servers.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			result, err := integration.Install(integration.Options{
				Force:       force,
				GlobalSkill: !skipGlobal,
			})
			if err != nil {
				return fmt.Errorf("failed to install IDE integration: %w", err)
			}

			lines := integration.FormatSummary(result)
			if len(lines) == 0 {
				cmd.Println("✓ IDE integration already up to date (use --force to overwrite skill files)")
				return nil
			}
			for _, line := range lines {
				cmd.Println(line)
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "Overwrite existing skill and rule files")
	cmd.Flags().BoolVar(&skipGlobal, "skip-global", false, "Do not install ~/.cursor/skills/reponerve/")

	return cmd
}
