package memorycmd

import (
	"time"

	"github.com/spf13/cobra"

	memorymodels "reponerve/internal/memory/models"
	"reponerve/internal/query/storage"
)

func newIntentsCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "intents",
		Short: "List intent memories",
		RunE: func(cmd *cobra.Command, args []string) error {
			db, err := loadDB()
			if err != nil {
				return err
			}
			defer db.Close()

			reader := storage.NewSQLiteIntentReader(db)
			var intents []*memorymodels.Intent
			if repositoryID != "" {
				intents, err = reader.ListByRepository(cmd.Context(), repositoryID)
			} else {
				intents, err = reader.ListAll(cmd.Context())
			}
			if err != nil {
				return err
			}

			headers := []string{"ID", "DESCRIPTION", "CREATED"}
			rows := make([][]string, len(intents))
			for i, val := range intents {
				rows[i] = []string{
					val.ID,
					val.Description,
					val.CreatedAt.Format(time.RFC3339),
				}
			}

			printTable(cmd.OutOrStdout(), headers, rows)
			return nil
		},
	}
}
