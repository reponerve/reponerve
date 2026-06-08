package memorycmd

import (
	"database/sql"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/reponerve/reponerve/internal/query/storage"
)

func newIntentTraceCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "intent <id>",
		Short: "Trace relationships for an intent memory",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]
			db, err := loadDB()
			if err != nil {
				return err
			}
			defer db.Close()

			intentReader := storage.NewSQLiteIntentReader(db)
			relReader := storage.NewSQLiteRelationshipReader(db)
			decReader := storage.NewSQLiteDecisionReader(db)
			eventReader := storage.NewSQLiteEventReader(db)

			// 1. Get intent
			intent, err := intentReader.GetByID(cmd.Context(), id)
			if err != nil {
				if err == sql.ErrNoRows {
					return fmt.Errorf("intent with ID %q not found", id)
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
			var events []string

			// Map of decision IDs we've already processed to avoid duplicates
			processedDecisions := make(map[string]bool)

			for _, r := range allRels {
				if r.FromID == intent.ID && r.Type == "INTENT_DRIVES_DECISION" {
					decID := r.ToID
					dec, err := decReader.GetByID(cmd.Context(), decID)
					if err == nil {
						decisions = append(decisions, dec.Title)
					}

					if !processedDecisions[decID] {
						processedDecisions[decID] = true
						// Find events for this decision
						for _, r2 := range allRels {
							if r2.FromID == decID && r2.Type == "DECISION_RESULTS_IN_EVENT" {
								e, err := eventReader.GetByID(cmd.Context(), r2.ToID)
								if err == nil {
									events = append(events, e.Title)
								}
							}
						}
					}
				}
			}

			// 4. Output tree structures
			printTreeSection(cmd.OutOrStdout(), "Intent", []string{intent.Description})
			printTreeSection(cmd.OutOrStdout(), "Decision", decisions)
			printTreeSection(cmd.OutOrStdout(), "Event", events)

			return nil
		},
	}
}
