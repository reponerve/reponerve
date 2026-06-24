package mcp

import (
	"fmt"
	"sort"
	"sync"
)

// Registry manages and provides access to registered MCP tools.
type Registry struct {
	mu    sync.RWMutex
	tools map[string]ToolDefinition
}

// NewRegistry creates a new Registry and registers default initial tools.
func NewRegistry() *Registry {
	r := &Registry{
		tools: make(map[string]ToolDefinition),
	}

	// Register initial default tool definitions
	_ = r.Register(ToolDefinition{
		Name:        "list_decisions",
		Description: "List all architectural decisions for a repository",
	})
	_ = r.Register(ToolDefinition{
		Name:        "get_decision",
		Description: "Retrieve a specific decision memory by its ID",
	})
	_ = r.Register(ToolDefinition{
		Name:        "list_events",
		Description: "List all events for a repository",
	})
	_ = r.Register(ToolDefinition{
		Name:        "get_event",
		Description: "Retrieve a specific event memory by its ID",
	})
	_ = r.Register(ToolDefinition{
		Name:        "list_intents",
		Description: "List all intents for a repository",
	})
	_ = r.Register(ToolDefinition{
		Name:        "get_intent",
		Description: "Retrieve a specific intent memory by its ID",
	})
	_ = r.Register(ToolDefinition{
		Name:        "list_facts",
		Description: "List all facts for a repository",
	})
	_ = r.Register(ToolDefinition{
		Name:        "get_fact",
		Description: "Retrieve a specific fact memory by its ID",
	})
	_ = r.Register(ToolDefinition{
		Name:        "trace_decision",
		Description: "Trace relationships for a specific decision memory",
	})
	_ = r.Register(ToolDefinition{
		Name:        "trace_event",
		Description: "Trace relationships for a specific event memory",
	})
	_ = r.Register(ToolDefinition{
		Name:        "explain_decision",
		Description: "Explain a decision memory",
	})
	_ = r.Register(ToolDefinition{
		Name:        "explain_event",
		Description: "Explain an event memory",
	})
	_ = r.Register(ToolDefinition{
		Name:        "generate_context",
		Description: "Generate a structured repository context briefing",
	})
	_ = r.Register(ToolDefinition{
		Name:        "export_context",
		Description: "Export repository context as rendered markdown",
	})
	_ = r.Register(ToolDefinition{
		Name:        "list_contributors",
		Description: "List all repository contributors",
	})
	_ = r.Register(ToolDefinition{
		Name:        "get_contributor",
		Description: "Retrieve details of a specific contributor",
	})
	_ = r.Register(ToolDefinition{
		Name:        "list_expertise",
		Description: "List all detected expertise records for a repository",
	})
	_ = r.Register(ToolDefinition{
		Name:        "trace_contributor",
		Description: "Trace a contributor's complete ownership, decisions, facts, and events",
	})
	_ = r.Register(ToolDefinition{
		Name:        "recommend_reviewers",
		Description: "Recommend reviewers using repository intelligence and ownership",
	})
	_ = r.Register(ToolDefinition{
		Name:        "discover_knowledge",
		Description: "Discover important repository knowledge and rank key artifacts",
	})
	_ = r.Register(ToolDefinition{
		Name:        "generate_learning_path",
		Description: "Generate a structured learning path for repository onboarding",
	})
	_ = r.Register(ToolDefinition{
		Name:        "generate_change_plan",
		Description: "Generate a change plan listing repository entities to examine before modifying code",
	})
	_ = r.Register(ToolDefinition{
		Name:        "trace_graph",
		Description: "Trace all reachable graph paths originating from a node",
	})
	_ = r.Register(ToolDefinition{
		Name:        "trace_path",
		Description: "Find graph paths connecting a start node to an end node",
	})
	_ = r.Register(ToolDefinition{
		Name:        "find_dependencies",
		Description: "Find outbound dependency paths from a node in the knowledge graph",
	})
	_ = r.Register(ToolDefinition{
		Name:        "find_dependents",
		Description: "Find inbound dependency paths pointing to a node in the knowledge graph",
	})
	_ = r.Register(ToolDefinition{
		Name:        "analyze_impact",
		Description: "Analyze the impact of a decision, fact, event, or contributor through the knowledge graph",
	})
	_ = r.Register(ToolDefinition{
		Name:        "ask",
		Description: "Answer a repository or development question using Code Intelligence and Repository Intelligence",
	})
	_ = r.Register(ToolDefinition{
		Name:        "explain",
		Description: "Explain a repository topic using evidence-backed code and memory context",
	})
	_ = r.Register(ToolDefinition{
		Name:        "explain_file",
		Description: "Explain a source file using Code Intelligence",
	})
	_ = r.Register(ToolDefinition{
		Name:        "explain_function",
		Description: "Explain a function symbol using Code Intelligence",
	})
	_ = r.Register(ToolDefinition{
		Name:        "explain_struct",
		Description: "Explain a struct symbol using Code Intelligence",
	})
	_ = r.Register(ToolDefinition{
		Name:        "explain_interface",
		Description: "Explain an interface symbol using Code Intelligence",
	})
	_ = r.Register(ToolDefinition{
		Name:        "explain_type",
		Description: "Explain a type alias symbol using Code Intelligence",
	})
	_ = r.Register(ToolDefinition{
		Name:        "plan",
		Description: "Generate a development plan for a task using repository and code intelligence",
	})
	_ = r.Register(ToolDefinition{
		Name:        "review",
		Description: "Prepare a review guide for a topic using repository and code intelligence",
	})
	_ = r.Register(ToolDefinition{
		Name:        "analyze_topic_impact",
		Description: "Analyze the impact of a topic, symbol, or area across code and repository memory",
	})
	_ = r.Register(ToolDefinition{
		Name:        "onboard",
		Description: "First-day repository context with key decisions and optional assignment plan",
	})
	_ = r.Register(ToolDefinition{
		Name:        "discover_surprises",
		Description: "Discover god nodes and surprising cross-community graph connections",
	})
	_ = r.Register(ToolDefinition{
		Name:        "suggest_questions",
		Description: "Suggest evidence-backed questions from graph structure",
	})
	_ = r.Register(ToolDefinition{
		Name:        "query_graph",
		Description: "Traverse the knowledge graph from a start node within a token budget",
	})
	_ = r.Register(ToolDefinition{
		Name:        "list_features",
		Description: "List derived repository features (domains and capabilities)",
	})
	_ = r.Register(ToolDefinition{
		Name:        "explain_feature",
		Description: "Explain a derived feature with code, ownership, and decision context",
	})
	_ = r.Register(ToolDefinition{
		Name:        "reuse_check",
		Description: "Find existing symbols and patterns to reuse before writing new code (Reuse Protocol)",
	})
	_ = r.Register(ToolDefinition{
		Name:        "ship_check",
		Description: "Assess ship readiness with blockers and advisories from repository evidence",
	})
	_ = r.Register(ToolDefinition{
		Name:        "remember",
		Description: "Store session knowledge as an evidence-backed fact with provenance",
	})
	_ = r.Register(ToolDefinition{
		Name:        "forget",
		Description: "Remove a session memory fact by ID",
	})

	return r
}

// Register registers a new tool definition in the registry.
func (r *Registry) Register(tool ToolDefinition) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if tool.Name == "" {
		return fmt.Errorf("tool name cannot be empty")
	}

	if _, exists := r.tools[tool.Name]; exists {
		return fmt.Errorf("tool %q already registered", tool.Name)
	}

	r.tools[tool.Name] = tool
	return nil
}

// Get retrieves a tool definition by name.
func (r *Registry) Get(name string) (ToolDefinition, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	t, exists := r.tools[name]
	return t, exists
}

// List returns all registered tool definitions, sorted alphabetically by name.
func (r *Registry) List() []ToolDefinition {
	r.mu.RLock()
	defer r.mu.RUnlock()

	list := make([]ToolDefinition, 0, len(r.tools))
	for _, t := range r.tools {
		list = append(list, t)
	}

	sort.Slice(list, func(i, j int) bool {
		return list[i].Name < list[j].Name
	})

	return list
}
