package workflowcmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/reponerve/reponerve/internal/agent/workflow"
	"github.com/reponerve/reponerve/internal/cli/devwire"
	"github.com/reponerve/reponerve/internal/config"
	"github.com/reponerve/reponerve/internal/storage/sqlite"
)

// NewCommand creates workflow template commands.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "workflow",
		Short: "Run fixed workflow templates (onboarding, review, change)",
	}
	cmd.AddCommand(
		newRunCommand(workflow.WorkflowTypeOnboarding, "onboarding", "First-day repository onboarding workflow", nil),
		newRunCommand(workflow.WorkflowTypeReviewPreparation, "review", "Review preparation workflow", []string{"topic"}),
		newRunCommand(workflow.WorkflowTypeChangePreparation, "change", "Change preparation workflow", []string{"entity"}),
	)
	return cmd
}

func newRunCommand(workflowType, name, description string, requiredFlags []string) *cobra.Command {
	var topic, entity string
	cmd := &cobra.Command{
		Use:   name,
		Short: description,
		RunE: func(cmd *cobra.Command, args []string) error {
			workspaceDir := config.GetWorkspaceDir()
			cfg, err := config.Load(workspaceDir)
			if err != nil {
				return fmt.Errorf("%s", config.FormatLoadError(workspaceDir, err))
			}
			db, err := sqlite.Open(cfg.Storage.SQLitePath)
			if err != nil {
				return err
			}
			defer db.Close()

			repoID, wf, err := devwire.WireWorkflowService(cmd.Context(), db, cfg.Repository.Path)
			if err != nil {
				return err
			}

			var pkg *workflow.WorkflowPackage
			switch workflowType {
			case workflow.WorkflowTypeOnboarding:
				pkg, err = wf.BuildOnboardingWorkflow(cmd.Context(), repoID)
			case workflow.WorkflowTypeReviewPreparation:
				query := strings.TrimSpace(topic)
				if query == "" {
					return fmt.Errorf("--topic is required")
				}
				pkg, err = wf.BuildReviewPreparationWorkflow(cmd.Context(), repoID, query)
			case workflow.WorkflowTypeChangePreparation:
				id := strings.TrimSpace(entity)
				if id == "" {
					return fmt.Errorf("--entity is required")
				}
				pkg, err = wf.BuildChangePreparationWorkflow(cmd.Context(), repoID, id)
			default:
				return fmt.Errorf("unsupported workflow type %q", workflowType)
			}
			if err != nil {
				return err
			}

			enc := json.NewEncoder(cmd.OutOrStdout())
			enc.SetIndent("", "  ")
			return enc.Encode(pkg)
		},
	}
	if contains(requiredFlags, "topic") {
		cmd.Flags().StringVar(&topic, "topic", "", "Review topic")
	}
	if contains(requiredFlags, "entity") {
		cmd.Flags().StringVar(&entity, "entity", "", "Entity ID to change")
	}
	return cmd
}

func contains(items []string, target string) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}
