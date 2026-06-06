package memorycmd

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"reponerve/internal/query/storage"
)

func newEventGetCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "event <id>",
		Short: "Retrieve a single event memory by ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]
			db, err := loadDB()
			if err != nil {
				return err
			}
			defer db.Close()

			reader := storage.NewSQLiteEventReader(db)
			event, err := reader.GetByID(cmd.Context(), id)
			if err != nil {
				if err == sql.ErrNoRows {
					return fmt.Errorf("event with ID %q not found", id)
				}
				return err
			}

			cmd.Println("Event")
			cmd.Println()
			cmd.Println("ID:")
			cmd.Println(event.ID)
			cmd.Println()
			cmd.Println("Type:")
			cmd.Println(event.EventType)
			cmd.Println()
			cmd.Println("Title:")
			cmd.Println(event.Title)
			cmd.Println()
			cmd.Println("Description:")
			cmd.Println(event.Description)
			cmd.Println()
			cmd.Println("Source:")
			cmd.Println(event.SourceID)
			cmd.Println()
			cmd.Println("Timestamp:")
			cmd.Println(event.Timestamp.Format(time.RFC3339))

			return nil
		},
	}
}
