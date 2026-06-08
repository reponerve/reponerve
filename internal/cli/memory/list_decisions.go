package memorycmd

import (
	"time"

	"github.com/spf13/cobra"

	memorymodels "github.com/reponerve/reponerve/internal/memory/models"
	"github.com/reponerve/reponerve/internal/query/storage"
)

func newDecisionsCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "decisions",
		Short: "List decision memories",
		RunE: func(cmd *cobra.Command, args []string) error {
			db, err := loadDB()
			if err != nil {
				return err
			}
			defer db.Close()

			reader := storage.NewSQLiteDecisionReader(db)
			var decisions []*memorymodels.Decision
			if repositoryID != "" {
				decisions, err = reader.ListByRepository(cmd.Context(), repositoryID)
			} else {
				decisions, err = reader.ListAll(cmd.Context())
			}
			if err != nil {
				return err
			}

			headers := []string{"ID", "STATUS", "TITLE", "CREATED"}
			rows := make([][]string, len(decisions))
			for i, d := range decisions {
				rows[i] = []string{
					d.ID,
					d.Status,
					d.Title,
					d.CreatedAt.Format(time.RFC3339),
				}
			}

			printTable(cmd.OutOrStdout(), headers, rows)
			return nil
		},
	}
}
