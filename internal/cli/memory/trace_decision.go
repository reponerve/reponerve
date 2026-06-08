package memorycmd

import (
	"database/sql"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/reponerve/reponerve/internal/query/storage"
)

func newDecisionTraceCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "decision <id>",
		Short: "Trace relationships for a decision memory",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]
			db, err := loadDB()
			if err != nil {
				return err
			}
			defer db.Close()

			decReader := storage.NewSQLiteDecisionReader(db)
			relReader := storage.NewSQLiteRelationshipReader(db)
			intentReader := storage.NewSQLiteIntentReader(db)
			factReader := storage.NewSQLiteFactReader(db)
			eventReader := storage.NewSQLiteEventReader(db)

			// 1. Get decision
			dec, err := decReader.GetByID(cmd.Context(), id)
			if err != nil {
				if err == sql.ErrNoRows {
					return fmt.Errorf("decision with ID %q not found", id)
				}
				return err
			}

			// 2. Fetch all relationships
			allRels, err := relReader.ListAll(cmd.Context())
			if err != nil {
				return err
			}

			// 3. Filter relationships and resolve related entities
			var intents []string
			var facts []string
			var events []string

			for _, r := range allRels {
				if r.ToID == dec.ID && r.Type == "INTENT_DRIVES_DECISION" {
					it, err := intentReader.GetByID(cmd.Context(), r.FromID)
					if err == nil {
						intents = append(intents, it.Description)
					}
				} else if r.ToID == dec.ID && r.Type == "FACT_SUPPORTS_DECISION" {
					f, err := factReader.GetByID(cmd.Context(), r.FromID)
					if err == nil {
						facts = append(facts, fmt.Sprintf("%s %s %s", f.Subject, f.Predicate, f.Object))
					}
				} else if r.FromID == dec.ID && r.Type == "DECISION_RESULTS_IN_EVENT" {
					e, err := eventReader.GetByID(cmd.Context(), r.ToID)
					if err == nil {
						events = append(events, e.Title)
					}
				}
			}

			// 4. Output tree structures
			printTreeSection(cmd.OutOrStdout(), "Decision", []string{dec.Title})
			printTreeSection(cmd.OutOrStdout(), "Intent", intents)
			printTreeSection(cmd.OutOrStdout(), "Fact", facts)
			printTreeSection(cmd.OutOrStdout(), "Event", events)

			return nil
		},
	}
}
