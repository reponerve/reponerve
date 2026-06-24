package discipline

import "time"

// LayerConvention describes a repository layout prefix and its role.
type LayerConvention struct {
	Prefix string `json:"prefix"`
	Role   string `json:"role"`
}

// Policy is the repo-adaptive Native Development Discipline contract.
type Policy struct {
	RepositoryID             string            `json:"repository_id"`
	GeneratedAt              time.Time         `json:"generated_at"`
	ADRDirectory             string            `json:"adr_directory,omitempty"`
	RequireADROnArchitecture bool              `json:"require_adr_on_architecture"`
	CIWorkflowFiles          []string          `json:"ci_workflow_files,omitempty"`
	DominantLanguage         string            `json:"dominant_language,omitempty"`
	LayerConventions         []LayerConvention `json:"layer_conventions,omitempty"`
	ShipCheckHints           []string          `json:"ship_check_hints,omitempty"`
	SourceServices           []string          `json:"source_services"`
}

// AgentSummary is the subset exposed on DE agent envelopes.
type AgentSummary struct {
	ADRDirectory             string            `json:"adr_directory,omitempty"`
	RequireADROnArchitecture bool              `json:"require_adr_on_architecture,omitempty"`
	CIWorkflowFiles          []string          `json:"ci_workflow_files,omitempty"`
	DominantLanguage         string            `json:"dominant_language,omitempty"`
	LayerConventions         []LayerConvention `json:"layer_conventions,omitempty"`
	ShipCheckHints           []string          `json:"ship_check_hints,omitempty"`
}

// AgentSummary returns envelope-safe policy fields.
func (p *Policy) AgentSummary() *AgentSummary {
	if p == nil {
		return nil
	}
	return &AgentSummary{
		ADRDirectory:             p.ADRDirectory,
		RequireADROnArchitecture: p.RequireADROnArchitecture,
		CIWorkflowFiles:          append([]string(nil), p.CIWorkflowFiles...),
		DominantLanguage:         p.DominantLanguage,
		LayerConventions:         append([]LayerConvention(nil), p.LayerConventions...),
		ShipCheckHints:           append([]string(nil), p.ShipCheckHints...),
	}
}
