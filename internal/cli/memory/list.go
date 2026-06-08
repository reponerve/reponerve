package memorycmd

import (
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/reponerve/reponerve/internal/config"
	"github.com/reponerve/reponerve/internal/storage/sqlite"
)

var repositoryID string

// newListCommand creates and returns the list subcommand.
func newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List repository memories",
		Long:  `Retrieve and list repository memory structures like events, decisions, intents, facts, and relationships.`,
	}

	// Register subcommands
	cmd.AddCommand(newEventsCommand())
	cmd.AddCommand(newDecisionsCommand())
	cmd.AddCommand(newIntentsCommand())
	cmd.AddCommand(newFactsCommand())
	cmd.AddCommand(newRelationshipsCommand())

	// Add repository persistent flag
	cmd.PersistentFlags().StringVar(&repositoryID, "repository", "", "Filter by repository ID")

	return cmd
}

// printTable formats and writes tabular memory data to out.
func printTable(out io.Writer, headers []string, rows [][]string) {
	if len(rows) == 0 {
		fmt.Fprintln(out, "No records found.")
		return
	}

	// We use text/tabwriter with tab separators. The column text includes " | " at boundaries.
	w := tabwriter.NewWriter(out, 0, 0, 1, ' ', 0)

	// Print headers
	for i, h := range headers {
		if i > 0 {
			fmt.Fprint(w, "\t| ")
		}
		fmt.Fprint(w, h)
	}
	fmt.Fprintln(w)

	// Print rows
	for _, row := range rows {
		for i, val := range row {
			if i > 0 {
				fmt.Fprint(w, "\t| ")
			}
			fmt.Fprint(w, val)
		}
		fmt.Fprintln(w)
	}

	w.Flush()
}

// loadDB loads the workspace configuration and opens the database.
func loadDB() (*sqlite.Database, error) {
	workspaceDir := config.GetWorkspaceDir()
	cfg, err := config.Load(workspaceDir)
	if err != nil {
		return nil, fmt.Errorf("workspace not initialized; run 'reponerve init' first")
	}

	db, err := sqlite.Open(cfg.Storage.SQLitePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	return db, nil
}

