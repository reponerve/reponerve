package development

import (
	"github.com/reponerve/reponerve/internal/agent/discipline"
	"github.com/reponerve/reponerve/internal/config"
)

// DisciplineCheck is a structured review item from repository discipline policy.
type DisciplineCheck struct {
	Category string `json:"category"`
	Severity string `json:"severity"`
	Message  string `json:"message"`
}

func appendReviewDiscipline(out *DevelopmentReviewGuide) {
	if out == nil {
		return
	}
	out.DisciplineChecks = disciplineChecksFromPolicy()
	if len(out.RecommendedNextTools) == 0 {
		out.RecommendedNextTools = []string{"ship_check", "reuse_check", "analyze_topic_impact"}
	}
	if len(out.DisciplineChecks) > 0 {
		out.SourceServices = mergeSourceServices(out.SourceServices, []string{sourceDevelopmentDiscipline})
	}
}

func disciplineChecksFromPolicy() []DisciplineCheck {
	policy, err := discipline.LoadPolicy(config.GetWorkspaceDir())
	if err != nil || policy == nil {
		return nil
	}

	var checks []DisciplineCheck
	for _, hint := range policy.ShipCheckHints {
		checks = append(checks, DisciplineCheck{
			Category: "repository_policy",
			Severity: "advisory",
			Message:  hint,
		})
	}
	if policy.RequireADROnArchitecture && policy.ADRDirectory != "" {
		checks = append(checks, DisciplineCheck{
			Category: "adr",
			Severity: "required",
			Message:  "Confirm significant architecture changes are recorded in " + policy.ADRDirectory,
		})
	}
	for _, layer := range policy.LayerConventions {
		checks = append(checks, DisciplineCheck{
			Category: "layer_convention",
			Severity: "advisory",
			Message:  layer.Prefix + " — " + layer.Role,
		})
	}
	if len(policy.CIWorkflowFiles) > 0 {
		checks = append(checks, DisciplineCheck{
			Category: "ci",
			Severity: "advisory",
			Message:  "Verify CI workflows pass before merge",
		})
	}
	return checks
}
