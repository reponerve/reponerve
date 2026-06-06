package memorycmd

import (
	"time"

	"github.com/spf13/cobra"

	"reponerve/internal/query/storage"
	models "reponerve/pkg/models"
)

func newEventsCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "events",
		Short: "List event memories",
		RunE: func(cmd *cobra.Command, args []string) error {
			db, err := loadDB()
			if err != nil {
				return err
			}
			defer db.Close()

			reader := storage.NewSQLiteEventReader(db)
			var events []*models.Event
			if repositoryID != "" {
				events, err = reader.ListByRepository(cmd.Context(), repositoryID)
			} else {
				events, err = reader.ListAll(cmd.Context())
			}
			if err != nil {
				return err
			}

			headers := []string{"ID", "TYPE", "TITLE", "CREATED"}
			rows := make([][]string, len(events))
			for i, e := range events {
				rows[i] = []string{
					e.ID,
					e.EventType,
					e.Title,
					e.Timestamp.Format(time.RFC3339),
				}
			}

			printTable(cmd.OutOrStdout(), headers, rows)
			return nil
		},
	}
}
