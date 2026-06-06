package server

import (
	"bufio"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"reponerve/internal/mcp"
	memorymodels "reponerve/internal/memory/models"
	models "reponerve/pkg/models"
)

// JSONRPCRequest represents a JSON-RPC 2.0 request or notification.
type JSONRPCRequest struct {
	JSONRPC string           `json:"jsonrpc"`
	ID      *json.RawMessage `json:"id,omitempty"`
	Method  string           `json:"method"`
	Params  json.RawMessage  `json:"params,omitempty"`
}

// JSONRPCResponse represents a JSON-RPC 2.0 response.
type JSONRPCResponse struct {
	JSONRPC string           `json:"jsonrpc"`
	ID      *json.RawMessage `json:"id"`
	Result  interface{}      `json:"result,omitempty"`
	Error   *JSONRPCError    `json:"error,omitempty"`
}

// JSONRPCError represents a JSON-RPC 2.0 error.
type JSONRPCError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// InitializeResult is the result returned by the initialize method.
type InitializeResult struct {
	ProtocolVersion string             `json:"protocolVersion"`
	Capabilities    ServerCapabilities `json:"capabilities"`
	ServerInfo      ServerInfo         `json:"serverInfo"`
}

// ServerCapabilities defines the capabilities supported by the server.
type ServerCapabilities struct {
	Tools *ToolsCapability `json:"tools,omitempty"`
}

// ToolsCapability represents capabilities related to tools.
type ToolsCapability struct {
	ListChanged bool `json:"listChanged"`
}

// ServerInfo represents metadata about the server implementation.
type ServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// ListToolsResult is the result returned by the tools/list method.
type ListToolsResult struct {
	Tools []Tool `json:"tools"`
}

// Tool represents an MCP tool definition.
type Tool struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	InputSchema InputSchema `json:"inputSchema"`
}

// InputSchema defines the input schema for a tool.
type InputSchema struct {
	Type       string                 `json:"type"`
	Properties map[string]interface{} `json:"properties"`
	Required   []string               `json:"required,omitempty"`
}

// CallToolParams represents the parameters for a tools/call request.
type CallToolParams struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments,omitempty"`
}

// CallToolResult represents the result of a tools/call request.
type CallToolResult struct {
	Content []ContentBlock `json:"content"`
	IsError bool           `json:"isError,omitempty"`
}

// ContentBlock represents a content block returned by a tool.
type ContentBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// DecisionTraceResult represents the structured JSON output of a decision trace.
type DecisionTraceResult struct {
	Decision *memorymodels.Decision `json:"decision"`
	Intents  []*memorymodels.Intent `json:"intents"`
	Facts    []*memorymodels.Fact   `json:"facts"`
	Events   []*models.Event        `json:"events"`
}

// EventTraceResult represents the structured JSON output of an event trace.
type EventTraceResult struct {
	Event     *models.Event            `json:"event"`
	Decisions []*memorymodels.Decision `json:"decisions"`
	Intents   []*memorymodels.Intent   `json:"intents"`
}

// DecisionExplanationResult represents the structured JSON output of a decision explanation.
type DecisionExplanationResult struct {
	Decision        *memorymodels.Decision `json:"decision"`
	Reason          []string               `json:"reason"`
	SupportingFacts []string               `json:"supportingFacts"`
	ResultingEvents []string               `json:"resultingEvents"`
}

// EventExplanationResult represents the structured JSON output of an event explanation.
type EventExplanationResult struct {
	Event    *models.Event `json:"event"`
	CausedBy []string      `json:"causedBy"`
	Reason   []string      `json:"reason"`
}

// Server handles MCP communication over standard input and output (STDIO).
type Server struct {
	registry *mcp.Registry
	service  *mcp.Service
	stdin    io.Reader
	stdout   io.Writer
}

// NewServer creates a new MCP Server instance.
func NewServer(registry *mcp.Registry, service *mcp.Service, stdin io.Reader, stdout io.Writer) *Server {
	return &Server{
		registry: registry,
		service:  service,
		stdin:    stdin,
		stdout:   stdout,
	}
}

// Start runs the JSON-RPC read-write loop on STDIO.
func (s *Server) Start(ctx context.Context) error {
	scanner := bufio.NewScanner(s.stdin)
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var req JSONRPCRequest
		if err := json.Unmarshal(line, &req); err != nil {
			s.sendError(nil, -32700, "Parse error: invalid JSON", nil)
			continue
		}

		// Validate jsonrpc version
		if req.JSONRPC != "2.0" {
			if req.ID != nil {
				s.sendError(req.ID, -32600, "Invalid Request: expected jsonrpc version '2.0'", nil)
			}
			continue
		}

		// Handle notifications (no id)
		if req.ID == nil {
			continue
		}

		// Route request
		switch req.Method {
		case "initialize":
			s.handleInitialize(req.ID)
		case "tools/list":
			s.handleToolsList(req.ID)
		case "tools/call":
			var params CallToolParams
			if err := json.Unmarshal(req.Params, &params); err != nil {
				s.sendError(req.ID, -32602, "Invalid params: failed to parse CallToolParams", nil)
				continue
			}
			s.handleCallTool(ctx, req.ID, params)
		default:
			s.sendError(req.ID, -32601, fmt.Sprintf("Method not found: %q", req.Method), nil)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("stdio read error: %w", err)
	}

	return nil
}

func (s *Server) handleInitialize(id *json.RawMessage) {
	result := InitializeResult{
		ProtocolVersion: "2024-11-05",
		Capabilities: ServerCapabilities{
			Tools: &ToolsCapability{
				ListChanged: false,
			},
		},
		ServerInfo: ServerInfo{
			Name:    "reponerve",
			Version: "0.1.0-alpha",
		},
	}
	s.sendResult(id, result)
}

func (s *Server) handleToolsList(id *json.RawMessage) {
	registeredTools := s.registry.List()
	tools := make([]Tool, len(registeredTools))
	for i, t := range registeredTools {
		tools[i] = Tool{
			Name:        t.Name,
			Description: t.Description,
			InputSchema: getInputSchema(t.Name),
		}
	}

	result := ListToolsResult{
		Tools: tools,
	}
	s.sendResult(id, result)
}

func (s *Server) handleCallTool(ctx context.Context, id *json.RawMessage, params CallToolParams) {
	if s.service == nil {
		s.sendToolError(id, "internal error: MCP service is not configured")
		return
	}

	var args map[string]string
	if len(params.Arguments) > 0 && string(params.Arguments) != "null" {
		if err := json.Unmarshal(params.Arguments, &args); err != nil {
			s.sendToolError(id, fmt.Sprintf("invalid arguments structure: %v", err))
			return
		}
	}

	getArg := func(key string, required bool) (string, error) {
		val, exists := args[key]
		if !exists || val == "" {
			if required {
				return "", fmt.Errorf("missing required argument: %s", key)
			}
			return "", nil
		}
		return val, nil
	}

	resolveRepoID := func() (string, error) {
		repoID, _ := getArg("repository_id", false)
		if repoID != "" {
			return repoID, nil
		}

		decs, err := s.service.DecisionReader.ListAll(ctx)
		if err == nil && len(decs) > 0 {
			return decs[0].RepositoryID, nil
		}

		evts, err := s.service.EventReader.ListAll(ctx)
		if err == nil && len(evts) > 0 {
			return evts[0].RepositoryID, nil
		}

		ints, err := s.service.IntentReader.ListAll(ctx)
		if err == nil && len(ints) > 0 {
			return ints[0].RepositoryID, nil
		}

		facts, err := s.service.FactReader.ListAll(ctx)
		if err == nil && len(facts) > 0 {
			return facts[0].RepositoryID, nil
		}

		return "", nil
	}

	switch params.Name {
	case "list_decisions":
		repoID, _ := getArg("repository_id", false)
		var decisions []*memorymodels.Decision
		var err error
		if repoID != "" {
			decisions, err = s.service.DecisionReader.ListByRepository(ctx, repoID)
		} else {
			decisions, err = s.service.DecisionReader.ListAll(ctx)
		}
		if err != nil {
			s.sendToolError(id, fmt.Sprintf("failed to list decisions: %v", err))
			return
		}
		if decisions == nil {
			decisions = []*memorymodels.Decision{}
		}
		s.sendToolSuccess(id, decisions)

	case "get_decision":
		decID, err := getArg("decision_id", true)
		if err != nil {
			s.sendToolError(id, err.Error())
			return
		}
		decision, err := s.service.DecisionReader.GetByID(ctx, decID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				s.sendToolError(id, fmt.Sprintf("decision with ID %q not found", decID))
			} else {
				s.sendToolError(id, fmt.Sprintf("failed to get decision: %v", err))
			}
			return
		}
		s.sendToolSuccess(id, decision)

	case "list_events":
		repoID, _ := getArg("repository_id", false)
		var events []*models.Event
		var err error
		if repoID != "" {
			events, err = s.service.EventReader.ListByRepository(ctx, repoID)
		} else {
			events, err = s.service.EventReader.ListAll(ctx)
		}
		if err != nil {
			s.sendToolError(id, fmt.Sprintf("failed to list events: %v", err))
			return
		}
		if events == nil {
			events = []*models.Event{}
		}
		s.sendToolSuccess(id, events)

	case "get_event":
		evtID, err := getArg("event_id", true)
		if err != nil {
			s.sendToolError(id, err.Error())
			return
		}
		event, err := s.service.EventReader.GetByID(ctx, evtID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				s.sendToolError(id, fmt.Sprintf("event with ID %q not found", evtID))
			} else {
				s.sendToolError(id, fmt.Sprintf("failed to get event: %v", err))
			}
			return
		}
		s.sendToolSuccess(id, event)

	case "list_intents":
		repoID, _ := getArg("repository_id", false)
		var intents []*memorymodels.Intent
		var err error
		if repoID != "" {
			intents, err = s.service.IntentReader.ListByRepository(ctx, repoID)
		} else {
			intents, err = s.service.IntentReader.ListAll(ctx)
		}
		if err != nil {
			s.sendToolError(id, fmt.Sprintf("failed to list intents: %v", err))
			return
		}
		if intents == nil {
			intents = []*memorymodels.Intent{}
		}
		s.sendToolSuccess(id, intents)

	case "get_intent":
		intID, err := getArg("intent_id", true)
		if err != nil {
			s.sendToolError(id, err.Error())
			return
		}
		intent, err := s.service.IntentReader.GetByID(ctx, intID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				s.sendToolError(id, fmt.Sprintf("intent with ID %q not found", intID))
			} else {
				s.sendToolError(id, fmt.Sprintf("failed to get intent: %v", err))
			}
			return
		}
		s.sendToolSuccess(id, intent)

	case "list_facts":
		repoID, _ := getArg("repository_id", false)
		var facts []*memorymodels.Fact
		var err error
		if repoID != "" {
			facts, err = s.service.FactReader.ListByRepository(ctx, repoID)
		} else {
			facts, err = s.service.FactReader.ListAll(ctx)
		}
		if err != nil {
			s.sendToolError(id, fmt.Sprintf("failed to list facts: %v", err))
			return
		}
		if facts == nil {
			facts = []*memorymodels.Fact{}
		}
		s.sendToolSuccess(id, facts)

	case "get_fact":
		factID, err := getArg("fact_id", true)
		if err != nil {
			s.sendToolError(id, err.Error())
			return
		}
		fact, err := s.service.FactReader.GetByID(ctx, factID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				s.sendToolError(id, fmt.Sprintf("fact with ID %q not found", factID))
			} else {
				s.sendToolError(id, fmt.Sprintf("failed to get fact: %v", err))
			}
			return
		}
		s.sendToolSuccess(id, fact)

	case "trace_decision":
		decID, err := getArg("decision_id", true)
		if err != nil {
			s.sendToolError(id, err.Error())
			return
		}
		dec, err := s.service.DecisionReader.GetByID(ctx, decID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				s.sendToolError(id, fmt.Sprintf("decision with ID %q not found", decID))
			} else {
				s.sendToolError(id, fmt.Sprintf("failed to get decision: %v", err))
			}
			return
		}
		allRels, err := s.service.RelationshipReader.ListAll(ctx)
		if err != nil {
			s.sendToolError(id, fmt.Sprintf("failed to list relationships: %v", err))
			return
		}

		var intents []*memorymodels.Intent
		var facts []*memorymodels.Fact
		var events []*models.Event

		for _, r := range allRels {
			if r.ToID == dec.ID && r.Type == "INTENT_DRIVES_DECISION" {
				it, err := s.service.IntentReader.GetByID(ctx, r.FromID)
				if err == nil {
					intents = append(intents, it)
				}
			} else if r.ToID == dec.ID && r.Type == "FACT_SUPPORTS_DECISION" {
				f, err := s.service.FactReader.GetByID(ctx, r.FromID)
				if err == nil {
					facts = append(facts, f)
				}
			} else if r.FromID == dec.ID && r.Type == "DECISION_RESULTS_IN_EVENT" {
				e, err := s.service.EventReader.GetByID(ctx, r.ToID)
				if err == nil {
					events = append(events, e)
				}
			}
		}

		if intents == nil {
			intents = []*memorymodels.Intent{}
		}
		if facts == nil {
			facts = []*memorymodels.Fact{}
		}
		if events == nil {
			events = []*models.Event{}
		}

		s.sendToolSuccess(id, DecisionTraceResult{
			Decision: dec,
			Intents:  intents,
			Facts:    facts,
			Events:   events,
		})

	case "trace_event":
		evtID, err := getArg("event_id", true)
		if err != nil {
			s.sendToolError(id, err.Error())
			return
		}
		evt, err := s.service.EventReader.GetByID(ctx, evtID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				s.sendToolError(id, fmt.Sprintf("event with ID %q not found", evtID))
			} else {
				s.sendToolError(id, fmt.Sprintf("failed to get event: %v", err))
			}
			return
		}
		allRels, err := s.service.RelationshipReader.ListAll(ctx)
		if err != nil {
			s.sendToolError(id, fmt.Sprintf("failed to list relationships: %v", err))
			return
		}

		var decisions []*memorymodels.Decision
		var intents []*memorymodels.Intent
		processedDecisions := make(map[string]bool)

		for _, r := range allRels {
			if r.ToID == evt.ID && r.Type == "DECISION_RESULTS_IN_EVENT" {
				decID := r.FromID
				dec, err := s.service.DecisionReader.GetByID(ctx, decID)
				if err == nil {
					decisions = append(decisions, dec)
				}

				if !processedDecisions[decID] {
					processedDecisions[decID] = true
					for _, r2 := range allRels {
						if r2.ToID == decID && r2.Type == "INTENT_DRIVES_DECISION" {
							it, err := s.service.IntentReader.GetByID(ctx, r2.FromID)
							if err == nil {
								intents = append(intents, it)
							}
						}
					}
				}
			}
		}

		if decisions == nil {
			decisions = []*memorymodels.Decision{}
		}
		if intents == nil {
			intents = []*memorymodels.Intent{}
		}

		s.sendToolSuccess(id, EventTraceResult{
			Event:     evt,
			Decisions: decisions,
			Intents:   intents,
		})

	case "explain_decision":
		decID, err := getArg("decision_id", true)
		if err != nil {
			s.sendToolError(id, err.Error())
			return
		}
		dec, err := s.service.DecisionReader.GetByID(ctx, decID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				s.sendToolError(id, fmt.Sprintf("decision with ID %q not found", decID))
			} else {
				s.sendToolError(id, fmt.Sprintf("failed to get decision: %v", err))
			}
			return
		}
		allRels, err := s.service.RelationshipReader.ListAll(ctx)
		if err != nil {
			s.sendToolError(id, fmt.Sprintf("failed to list relationships: %v", err))
			return
		}

		var intents []string
		var facts []string
		var events []string

		for _, r := range allRels {
			if r.ToID == dec.ID && r.Type == "INTENT_DRIVES_DECISION" {
				it, err := s.service.IntentReader.GetByID(ctx, r.FromID)
				if err == nil {
					intents = append(intents, it.Description)
				}
			} else if r.ToID == dec.ID && r.Type == "FACT_SUPPORTS_DECISION" {
				f, err := s.service.FactReader.GetByID(ctx, r.FromID)
				if err == nil {
					facts = append(facts, fmt.Sprintf("%s %s %s", f.Subject, f.Predicate, f.Object))
				}
			} else if r.FromID == dec.ID && r.Type == "DECISION_RESULTS_IN_EVENT" {
				e, err := s.service.EventReader.GetByID(ctx, r.ToID)
				if err == nil {
					events = append(events, e.Title)
				}
			}
		}

		if intents == nil {
			intents = []string{}
		}
		if facts == nil {
			facts = []string{}
		}
		if events == nil {
			events = []string{}
		}

		s.sendToolSuccess(id, DecisionExplanationResult{
			Decision:        dec,
			Reason:          intents,
			SupportingFacts: facts,
			ResultingEvents: events,
		})

	case "explain_event":
		evtID, err := getArg("event_id", true)
		if err != nil {
			s.sendToolError(id, err.Error())
			return
		}
		evt, err := s.service.EventReader.GetByID(ctx, evtID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				s.sendToolError(id, fmt.Sprintf("event with ID %q not found", evtID))
			} else {
				s.sendToolError(id, fmt.Sprintf("failed to get event: %v", err))
			}
			return
		}
		allRels, err := s.service.RelationshipReader.ListAll(ctx)
		if err != nil {
			s.sendToolError(id, fmt.Sprintf("failed to list relationships: %v", err))
			return
		}

		var decisions []string
		var intents []string
		processedDecisions := make(map[string]bool)

		for _, r := range allRels {
			if r.ToID == evt.ID && r.Type == "DECISION_RESULTS_IN_EVENT" {
				decID := r.FromID
				dec, err := s.service.DecisionReader.GetByID(ctx, decID)
				if err == nil {
					decisions = append(decisions, dec.Title)
				}

				if !processedDecisions[decID] {
					processedDecisions[decID] = true
					for _, r2 := range allRels {
						if r2.ToID == decID && r2.Type == "INTENT_DRIVES_DECISION" {
							it, err := s.service.IntentReader.GetByID(ctx, r2.FromID)
							if err == nil {
								intents = append(intents, it.Description)
							}
						}
					}
				}
			}
		}

		if decisions == nil {
			decisions = []string{}
		}
		if intents == nil {
			intents = []string{}
		}

		s.sendToolSuccess(id, EventExplanationResult{
			Event:    evt,
			CausedBy: decisions,
			Reason:   intents,
		})

	case "generate_context":
		repoID, err := resolveRepoID()
		if err != nil {
			s.sendToolError(id, fmt.Sprintf("failed to resolve repository ID: %v", err))
			return
		}
		rc, err := s.service.Generator.Generate(ctx, repoID)
		if err != nil {
			s.sendToolError(id, fmt.Sprintf("failed to generate context: %v", err))
			return
		}
		s.sendToolSuccess(id, rc)

	case "export_context":
		repoID, err := resolveRepoID()
		if err != nil {
			s.sendToolError(id, fmt.Sprintf("failed to resolve repository ID: %v", err))
			return
		}
		rc, err := s.service.Generator.Generate(ctx, repoID)
		if err != nil {
			s.sendToolError(id, fmt.Sprintf("failed to generate context: %v", err))
			return
		}

		if len(rc.Decisions) == 0 && len(rc.Intents) == 0 && len(rc.Facts) == 0 && len(rc.Events) == 0 {
			result := CallToolResult{
				Content: []ContentBlock{
					{
						Type: "text",
						Text: "No repository context available.",
					},
				},
				IsError: false,
			}
			s.sendResult(id, result)
			return
		}

		markdown, err := s.service.Renderer.Render(rc)
		if err != nil {
			s.sendToolError(id, fmt.Sprintf("failed to render context: %v", err))
			return
		}

		result := CallToolResult{
			Content: []ContentBlock{
				{
					Type: "text",
					Text: markdown,
				},
			},
			IsError: false,
		}
		s.sendResult(id, result)

	default:
		s.sendToolError(id, fmt.Sprintf("unknown tool name: %q", params.Name))
	}
}

func (s *Server) sendToolError(id *json.RawMessage, errMsg string) {
	result := CallToolResult{
		Content: []ContentBlock{
			{
				Type: "text",
				Text: errMsg,
			},
		},
		IsError: true,
	}
	s.sendResult(id, result)
}

func (s *Server) sendToolSuccess(id *json.RawMessage, data interface{}) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		s.sendToolError(id, fmt.Sprintf("failed to marshal tool output: %v", err))
		return
	}
	result := CallToolResult{
		Content: []ContentBlock{
			{
				Type: "text",
				Text: string(jsonData),
			},
		},
		IsError: false,
	}
	s.sendResult(id, result)
}

// getInputSchema returns the input schema definition for a given tool.
func getInputSchema(toolName string) InputSchema {
	schema := InputSchema{
		Type:       "object",
		Properties: make(map[string]interface{}),
	}

	switch toolName {
	case "get_decision", "trace_decision", "explain_decision":
		schema.Properties["decision_id"] = map[string]interface{}{
			"type":        "string",
			"description": "The unique identifier of the decision memory",
		}
		schema.Required = []string{"decision_id"}

	case "get_event", "trace_event", "explain_event":
		schema.Properties["event_id"] = map[string]interface{}{
			"type":        "string",
			"description": "The unique identifier of the event memory",
		}
		schema.Required = []string{"event_id"}

	case "get_intent":
		schema.Properties["intent_id"] = map[string]interface{}{
			"type":        "string",
			"description": "The unique identifier of the intent memory",
		}
		schema.Required = []string{"intent_id"}

	case "get_fact":
		schema.Properties["fact_id"] = map[string]interface{}{
			"type":        "string",
			"description": "The unique identifier of the fact memory",
		}
		schema.Required = []string{"fact_id"}

	case "list_decisions", "list_events", "list_intents", "list_facts", "generate_context", "export_context":
		schema.Properties["repository_id"] = map[string]interface{}{
			"type":        "string",
			"description": "Optional repository filter",
		}
	}

	return schema
}

func (s *Server) sendResult(id *json.RawMessage, result interface{}) {
	resp := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	}
	s.writeResponse(resp)
}

func (s *Server) sendError(id *json.RawMessage, code int, msg string, data interface{}) {
	resp := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error: &JSONRPCError{
			Code:    code,
			Message: msg,
			Data:    data,
		},
	}
	s.writeResponse(resp)
}

func (s *Server) writeResponse(resp interface{}) {
	data, err := json.Marshal(resp)
	if err != nil {
		return
	}
	_, _ = s.stdout.Write(append(data, '\n'))
}
