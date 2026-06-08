package memorycmd

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/reponerve/reponerve/internal/query/storage"
)

func newIntentGetCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "intent <id>",
		Short: "Retrieve a single intent memory by ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]
			db, err := loadDB()
			if err != nil {
				return err
			}
			defer db.Close()

			reader := storage.NewSQLiteIntentReader(db)
			intent, err := reader.GetByID(cmd.Context(), id)
			if err != nil {
				if err == sql.ErrNoRows {
					return fmt.Errorf("intent with ID %q not found", id)
				}
				return err
			}

			cmd.Println("Intent")
			cmd.Println()
			cmd.Println("ID:")
			cmd.Println(intent.ID)
			cmd.Println()
			cmd.Println("Description:")
			cmd.Println(intent.Description)
			cmd.Println()
			cmd.Println("Source:")
			cmd.Println(intent.SourceID)
			cmd.Println()
			cmd.Println("Created:")
			cmd.Println(intent.CreatedAt.Format(time.RFC3339))

			return nil
		},
	}
}
