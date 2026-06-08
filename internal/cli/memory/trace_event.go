package memorycmd

import (
	"database/sql"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/reponerve/reponerve/internal/query/storage"
)

func newEventTraceCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "event <id>",
		Short: "Trace relationships for an event memory",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]
			db, err := loadDB()
			if err != nil {
				return err
			}
			defer db.Close()

			evtReader := storage.NewSQLiteEventReader(db)
			relReader := storage.NewSQLiteRelationshipReader(db)
			decReader := storage.NewSQLiteDecisionReader(db)
			intentReader := storage.NewSQLiteIntentReader(db)

			// 1. Get event
			evt, err := evtReader.GetByID(cmd.Context(), id)
			if err != nil {
				if err == sql.ErrNoRows {
					return fmt.Errorf("event with ID %q not found", id)
				}
				return err
			}

			// 2. Fetch all relationships
			allRels, err := relReader.ListAll(cmd.Context())
			if err != nil {
				return err
			}

			// 3. Filter relationships and resolve related entities
			var decisions []string
			var intents []string

			// Map of decision IDs we've already processed to avoid duplicates
			processedDecisions := make(map[string]bool)

			for _, r := range allRels {
				if r.ToID == evt.ID && r.Type == "DECISION_RESULTS_IN_EVENT" {
					decID := r.FromID
					dec, err := decReader.GetByID(cmd.Context(), decID)
					if err == nil {
						decisions = append(decisions, dec.Title)
					}

					if !processedDecisions[decID] {
						processedDecisions[decID] = true
						// Find intents for this decision
						for _, r2 := range allRels {
							if r2.ToID == decID && r2.Type == "INTENT_DRIVES_DECISION" {
								it, err := intentReader.GetByID(cmd.Context(), r2.FromID)
								if err == nil {
									intents = append(intents, it.Description)
								}
							}
						}
					}
				}
			}

			// 4. Output tree structures
			printTreeSection(cmd.OutOrStdout(), "Event", []string{evt.Title})
			printTreeSection(cmd.OutOrStdout(), "Decision", decisions)
			printTreeSection(cmd.OutOrStdout(), "Intent", intents)

			return nil
		},
	}
}
