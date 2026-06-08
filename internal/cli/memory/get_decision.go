package memorycmd

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/reponerve/reponerve/internal/query/storage"
)

func newDecisionGetCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "decision <id>",
		Short: "Retrieve a single decision memory by ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]
			db, err := loadDB()
			if err != nil {
				return err
			}
			defer db.Close()

			reader := storage.NewSQLiteDecisionReader(db)
			dec, err := reader.GetByID(cmd.Context(), id)
			if err != nil {
				if err == sql.ErrNoRows {
					return fmt.Errorf("decision with ID %q not found", id)
				}
				return err
			}

			cmd.Println("Decision")
			cmd.Println()
			cmd.Println("ID:")
			cmd.Println(dec.ID)
			cmd.Println()
			cmd.Println("Title:")
			cmd.Println(dec.Title)
			cmd.Println()
			cmd.Println("Status:")
			cmd.Println(dec.Status)
			cmd.Println()
			cmd.Println("Source:")
			cmd.Println(dec.SourceID)
			cmd.Println()
			cmd.Println("Created:")
			cmd.Println(dec.CreatedAt.Format(time.RFC3339))

			return nil
		},
	}
}
