package memorycmd

import (
	"github.com/spf13/cobra"

	memorymodels "github.com/reponerve/reponerve/internal/memory/models"
	"github.com/reponerve/reponerve/internal/query/storage"
)

func newFactsCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "facts",
		Short: "List fact memories",
		RunE: func(cmd *cobra.Command, args []string) error {
			db, err := loadDB()
			if err != nil {
				return err
			}
			defer db.Close()

			reader := storage.NewSQLiteFactReader(db)
			var facts []*memorymodels.Fact
			if repositoryID != "" {
				facts, err = reader.ListByRepository(cmd.Context(), repositoryID)
			} else {
				facts, err = reader.ListAll(cmd.Context())
			}
			if err != nil {
				return err
			}

			headers := []string{"ID", "SUBJECT", "PREDICATE", "OBJECT"}
			rows := make([][]string, len(facts))
			for i, val := range facts {
				rows[i] = []string{
					val.ID,
					val.Subject,
					val.Predicate,
					val.Object,
				}
			}

			printTable(cmd.OutOrStdout(), headers, rows)
			return nil
		},
	}
}
