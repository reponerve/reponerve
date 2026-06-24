package development

// MCPResult is the structured MCP tool payload for Development Experience commands.
type MCPResult struct {
	Formatted  string           `json:"formatted"`
	Structured any              `json:"structured"`
	Agent      AgentContextMeta `json:"agent"`
}

// NewMCPResult builds the agent context envelope for MCP and JSON consumers.
func NewMCPResult(formatted string, structured any) MCPResult {
	meta := BuildAgentContextMeta(structured)
	applyDisciplinePolicy(&meta)
	return MCPResult{
		Formatted:  formatted,
		Structured: structured,
		Agent:      meta,
	}
}
