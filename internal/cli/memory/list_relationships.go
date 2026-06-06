package memorycmd

import (
	"github.com/spf13/cobra"

	memorymodels "reponerve/internal/memory/models"
	"reponerve/internal/query/storage"
)

func newRelationshipsCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "relationships",
		Short: "List relationship memories",
		RunE: func(cmd *cobra.Command, args []string) error {
			db, err := loadDB()
			if err != nil {
				return err
			}
			defer db.Close()

			reader := storage.NewSQLiteRelationshipReader(db)
			var relationships []*memorymodels.Relationship
			if repositoryID != "" {
				relationships, err = reader.ListByRepository(cmd.Context(), repositoryID)
			} else {
				relationships, err = reader.ListAll(cmd.Context())
			}
			if err != nil {
				return err
			}

			headers := []string{"ID", "TYPE", "FROM", "TO"}
			rows := make([][]string, len(relationships))
			for i, val := range relationships {
				rows[i] = []string{
					val.ID,
					val.Type,
					val.FromID,
					val.ToID,
				}
			}

			printTable(cmd.OutOrStdout(), headers, rows)
			return nil
		},
	}
}
