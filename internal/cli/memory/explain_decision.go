package memorycmd

import (
	"database/sql"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/reponerve/reponerve/internal/query/storage"
)

func newDecisionExplainCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "decision <id>",
		Short: "Explain a decision memory",
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

			// 4. Output the explanation
			cmd.Println("Decision:")
			cmd.Println(dec.Title)
			cmd.Println()

			if len(intents) > 0 {
				cmd.Println("Reason:")
				for _, it := range intents {
					cmd.Println(it)
				}
				cmd.Println()
			}

			if len(facts) > 0 {
				cmd.Println("Supporting Facts:")
				for _, f := range facts {
					cmd.Printf("- %s\n", f)
				}
				cmd.Println()
			}

			if len(events) > 0 {
				cmd.Println("Resulting Events:")
				for _, e := range events {
					cmd.Printf("- %s\n", e)
				}
				cmd.Println()
			}

			return nil
		},
	}
}
