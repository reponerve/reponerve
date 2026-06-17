package devwire

import (
	"encoding/json"

	"github.com/spf13/cobra"

	"github.com/reponerve/reponerve/internal/agent/development"
)

// BindJSONFlag registers --json on a Development Experience CLI command.
func BindJSONFlag(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Emit MCP-compatible JSON (structured, agent, formatted) for AI chat without MCP")
}

// BindDECmd registers --json on a Development Experience command.
func BindDECmd(cmd *cobra.Command) *cobra.Command {
	BindJSONFlag(cmd)
	return cmd
}

// WriteDEResult prints human text or the MCP envelope based on --json.
func WriteDEResult(cmd *cobra.Command, formatted string, structured any) error {
	useJSON, err := cmd.Flags().GetBool("json")
	if err != nil {
		return err
	}
	if useJSON {
		enc := json.NewEncoder(cmd.OutOrStdout())
		enc.SetIndent("", "  ")
		return enc.Encode(development.NewMCPResult(formatted, structured))
	}
	cmd.Print(formatted)
	return nil
}
