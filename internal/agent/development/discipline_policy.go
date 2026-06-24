package development

import (
	"github.com/reponerve/reponerve/internal/agent/discipline"
	"github.com/reponerve/reponerve/internal/config"
)

func applyDisciplinePolicy(meta *AgentContextMeta) {
	if meta == nil {
		return
	}
	policy, err := discipline.LoadPolicy(config.GetWorkspaceDir())
	if err != nil || policy == nil {
		return
	}
	meta.DisciplinePolicy = policy.AgentSummary()
	if len(policy.ShipCheckHints) > 0 {
		meta.GuidanceForAgent = append(meta.GuidanceForAgent, policy.ShipCheckHints...)
	}
}
