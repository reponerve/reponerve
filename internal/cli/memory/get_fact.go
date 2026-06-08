package memorycmd

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/reponerve/reponerve/internal/query/storage"
)

func newFactGetCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "fact <id>",
		Short: "Retrieve a single fact memory by ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]
			db, err := loadDB()
			if err != nil {
				return err
			}
			defer db.Close()

			reader := storage.NewSQLiteFactReader(db)
			fact, err := reader.GetByID(cmd.Context(), id)
			if err != nil {
				if err == sql.ErrNoRows {
					return fmt.Errorf("fact with ID %q not found", id)
				}
				return err
			}

			cmd.Println("Fact")
			cmd.Println()
			cmd.Println("ID:")
			cmd.Println(fact.ID)
			cmd.Println()
			cmd.Println("Subject:")
			cmd.Println(fact.Subject)
			cmd.Println()
			cmd.Println("Predicate:")
			cmd.Println(fact.Predicate)
			cmd.Println()
			cmd.Println("Object:")
			cmd.Println(fact.Object)
			cmd.Println()
			cmd.Println("Source:")
			cmd.Println(fact.SourceID)
			cmd.Println()
			cmd.Println("Created:")
			cmd.Println(fact.CreatedAt.Format(time.RFC3339))

			return nil
		},
	}
}
