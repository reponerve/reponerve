package server

import (
	"bufio"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/reponerve/reponerve/internal/graph/model"
	"github.com/reponerve/reponerve/internal/graph/traversal"
	"github.com/reponerve/reponerve/internal/mcp"
	memorymodels "github.com/reponerve/reponerve/internal/memory/models"
	models "github.com/reponerve/reponerve/pkg/models"
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

	case "list_contributors":
		repoID, err := resolveRepoID()
		if err != nil || repoID == "" {
			s.sendToolError(id, "failed to resolve repository ID")
			return
		}
		list, err := s.service.OwnershipReader.ListContributors(ctx, repoID)
		if err != nil {
			s.sendToolError(id, fmt.Sprintf("failed to list contributors: %v", err))
			return
		}
		s.sendToolSuccess(id, list)

	case "get_contributor":
		contribID, err := getArg("contributor_id", true)
		if err != nil {
			s.sendToolError(id, err.Error())
			return
		}
		repoID, err := resolveRepoID()
		if err != nil || repoID == "" {
			s.sendToolError(id, "failed to resolve repository ID")
			return
		}
		c, err := s.service.OwnershipReader.GetContributor(ctx, repoID, contribID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				s.sendToolError(id, fmt.Sprintf("contributor with ID %q not found", contribID))
			} else {
				s.sendToolError(id, fmt.Sprintf("failed to get contributor: %v", err))
			}
			return
		}
		s.sendToolSuccess(id, c)

	case "list_expertise":
		repoID, err := resolveRepoID()
		if err != nil || repoID == "" {
			s.sendToolError(id, "failed to resolve repository ID")
			return
		}
		list, err := s.service.OwnershipReader.ListExpertise(ctx, repoID)
		if err != nil {
			s.sendToolError(id, fmt.Sprintf("failed to list expertise: %v", err))
			return
		}
		s.sendToolSuccess(id, list)

	case "trace_contributor":
		contribID, err := getArg("contributor_id", true)
		if err != nil {
			s.sendToolError(id, err.Error())
			return
		}
		repoID, err := resolveRepoID()
		if err != nil || repoID == "" {
			s.sendToolError(id, "failed to resolve repository ID")
			return
		}
		trace, err := s.service.OwnershipReader.TraceContributor(ctx, repoID, contribID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				s.sendToolError(id, fmt.Sprintf("contributor with ID %q not found", contribID))
			} else {
				s.sendToolError(id, fmt.Sprintf("failed to trace contributor: %v", err))
			}
			return
		}
		s.sendToolSuccess(id, trace)

	case "recommend_reviewers":
		repoID, err := getArg("repository_id", true)
		if err != nil {
			s.sendToolError(id, err.Error())
			return
		}
		recType, err := getArg("recommendation_type", true)
		if err != nil {
			s.sendToolError(id, err.Error())
			return
		}

		if s.service.ReviewerService == nil {
			s.sendToolError(id, "reviewer service is not configured")
			return
		}

		var report interface{}
		switch recType {
		case "repository":
			report, err = s.service.ReviewerService.RecommendRepositoryReviewers(ctx, repoID)
		case "domain":
			domain, errArg := getArg("domain", true)
			if errArg != nil {
				s.sendToolError(id, errArg.Error())
				return
			}
			report, err = s.service.ReviewerService.RecommendDomainReviewers(ctx, repoID, domain)
		case "impact":
			entID, errArg := getArg("entity_id", true)
			if errArg != nil {
				s.sendToolError(id, errArg.Error())
				return
			}
			report, err = s.service.ReviewerService.RecommendImpactReviewers(ctx, repoID, entID)
		default:
			s.sendToolError(id, fmt.Sprintf("unsupported recommendation_type %q: must be one of repository, domain, impact", recType))
			return
		}

		if err != nil {
			s.sendToolError(id, fmt.Sprintf("failed to recommend reviewers: %v", err))
			return
		}
		s.sendToolSuccess(id, report)

	case "discover_knowledge":
		repoID, err := getArg("repository_id", true)
		if err != nil {
			s.sendToolError(id, err.Error())
			return
		}

		if s.service.DiscoveryService == nil {
			s.sendToolError(id, "discovery service is not configured")
			return
		}

		report, err := s.service.DiscoveryService.Discover(ctx, repoID)
		if err != nil {
			s.sendToolError(id, fmt.Sprintf("failed to discover knowledge: %v", err))
			return
		}
		s.sendToolSuccess(id, report)

	case "generate_learning_path":
		repoID, err := getArg("repository_id", true)
		if err != nil {
			s.sendToolError(id, err.Error())
			return
		}
		pathType, err := getArg("path_type", true)
		if err != nil {
			s.sendToolError(id, err.Error())
			return
		}

		if s.service.LearningService == nil {
			s.sendToolError(id, "learning service is not configured")
			return
		}

		var path interface{}
		switch pathType {
		case "repository":
			path, err = s.service.LearningService.GenerateRepositoryPath(ctx, repoID)
		case "domain":
			domain, errArg := getArg("domain", true)
			if errArg != nil {
				s.sendToolError(id, errArg.Error())
				return
			}
			path, err = s.service.LearningService.GenerateDomainPath(ctx, repoID, domain)
		case "contributor":
			contribID, errArg := getArg("contributor_id", true)
			if errArg != nil {
				s.sendToolError(id, errArg.Error())
				return
			}
			path, err = s.service.LearningService.GenerateContributorPath(ctx, repoID, contribID)
		default:
			s.sendToolError(id, fmt.Sprintf("unsupported path_type %q: must be one of repository, domain, contributor", pathType))
			return
		}

		if err != nil {
			s.sendToolError(id, fmt.Sprintf("failed to generate learning path: %v", err))
			return
		}
		s.sendToolSuccess(id, path)

	case "generate_change_plan":
		repoID, err := getArg("repository_id", true)
		if err != nil {
			s.sendToolError(id, err.Error())
			return
		}
		entityType, err := getArg("entity_type", true)
		if err != nil {
			s.sendToolError(id, err.Error())
			return
		}
		entityID, err := getArg("entity_id", true)
		if err != nil {
			s.sendToolError(id, err.Error())
			return
		}

		if s.service.ChangePlanService == nil {
			s.sendToolError(id, "change plan service is not configured")
			return
		}

		var plan interface{}
		switch strings.ToLower(entityType) {
		case "decision":
			plan, err = s.service.ChangePlanService.GenerateDecisionPlan(ctx, repoID, entityID)
		case "fact":
			plan, err = s.service.ChangePlanService.GenerateFactPlan(ctx, repoID, entityID)
		case "event":
			plan, err = s.service.ChangePlanService.GenerateEventPlan(ctx, repoID, entityID)
		case "contributor":
			plan, err = s.service.ChangePlanService.GenerateContributorPlan(ctx, repoID, entityID)
		default:
			s.sendToolError(id, fmt.Sprintf("unsupported entity_type %q: must be one of decision, fact, event, contributor", entityType))
			return
		}

		if err != nil {
			s.sendToolError(id, fmt.Sprintf("failed to generate change plan: %v", err))
			return
		}
		s.sendToolSuccess(id, plan)

	case "trace_graph":
		nodeID, err := getArg("node_id", true)
		if err != nil {
			s.sendToolError(id, err.Error())
			return
		}
		if s.service.GraphTraversalEngine == nil {
			s.sendToolError(id, "graph traversal engine is not configured")
			return
		}
		repoID, err := resolveRepoID()
		if err != nil || repoID == "" {
			s.sendToolError(id, "failed to resolve repository ID")
			return
		}
		maxDepth := 10
		if d, _ := getArg("max_depth", false); d != "" {
			if v, err := strconv.Atoi(d); err == nil && v > 0 {
				maxDepth = v
			}
		}
		opts := traversal.TraversalOptions{
			MaxDepth:       maxDepth,
			IncludeStored:  true,
			IncludeDerived: true,
		}
		result, err := s.service.GraphTraversalEngine.TraceGraph(ctx, repoID, nodeID, opts)
		if err != nil {
			s.sendToolError(id, fmt.Sprintf("graph traversal failed: %v", err))
			return
		}
		s.sendToolSuccess(id, result)

	case "trace_path":
		startNodeID, err := getArg("start_node_id", true)
		if err != nil {
			s.sendToolError(id, err.Error())
			return
		}
		endNodeID, err := getArg("end_node_id", true)
		if err != nil {
			s.sendToolError(id, err.Error())
			return
		}
		if s.service.GraphTraversalEngine == nil {
			s.sendToolError(id, "graph traversal engine is not configured")
			return
		}
		repoID, err := resolveRepoID()
		if err != nil || repoID == "" {
			s.sendToolError(id, "failed to resolve repository ID")
			return
		}
		opts := traversal.TraversalOptions{
			MaxDepth:       10,
			IncludeStored:  true,
			IncludeDerived: true,
		}
		allPaths, err := s.service.GraphTraversalEngine.FindDependencies(ctx, repoID, startNodeID, opts)
		if err != nil {
			s.sendToolError(id, fmt.Sprintf("graph traversal failed: %v", err))
			return
		}
		// Filter to paths ending at endNodeID
		var matchedPaths []*traversal.TraversalPath
		for _, p := range allPaths.Paths {
			if len(p.Nodes) > 0 && p.Nodes[len(p.Nodes)-1].ID == endNodeID {
				matchedPaths = append(matchedPaths, p)
			}
		}
		if matchedPaths == nil {
			matchedPaths = []*traversal.TraversalPath{}
		}
		s.sendToolSuccess(id, matchedPaths)

	case "find_dependencies":
		nodeID, err := getArg("node_id", true)
		if err != nil {
			s.sendToolError(id, err.Error())
			return
		}
		if s.service.GraphTraversalEngine == nil {
			s.sendToolError(id, "graph traversal engine is not configured")
			return
		}
		repoID, err := resolveRepoID()
		if err != nil || repoID == "" {
			s.sendToolError(id, "failed to resolve repository ID")
			return
		}
		opts := traversal.TraversalOptions{
			MaxDepth:       10,
			IncludeStored:  true,
			IncludeDerived: true,
		}
		result, err := s.service.GraphTraversalEngine.FindDependencies(ctx, repoID, nodeID, opts)
		if err != nil {
			s.sendToolError(id, fmt.Sprintf("find dependencies failed: %v", err))
			return
		}
		s.sendToolSuccess(id, result)

	case "find_dependents":
		nodeID, err := getArg("node_id", true)
		if err != nil {
			s.sendToolError(id, err.Error())
			return
		}
		if s.service.GraphTraversalEngine == nil {
			s.sendToolError(id, "graph traversal engine is not configured")
			return
		}
		repoID, err := resolveRepoID()
		if err != nil || repoID == "" {
			s.sendToolError(id, "failed to resolve repository ID")
			return
		}
		opts := traversal.TraversalOptions{
			MaxDepth:       10,
			IncludeStored:  true,
			IncludeDerived: true,
		}
		result, err := s.service.GraphTraversalEngine.FindDependents(ctx, repoID, nodeID, opts)
		if err != nil {
			s.sendToolError(id, fmt.Sprintf("find dependents failed: %v", err))
			return
		}
		s.sendToolSuccess(id, result)

	case "analyze_impact":
		// node_id is the entity ID (decision ID, fact ID, etc.)
		// node_type indicates which impact analysis to run
		entityID, err := getArg("node_id", true)
		if err != nil {
			s.sendToolError(id, err.Error())
			return
		}
		nodeType, err := getArg("node_type", true)
		if err != nil {
			s.sendToolError(id, err.Error())
			return
		}
		if s.service.GraphImpactService == nil {
			s.sendToolError(id, "graph impact service is not configured")
			return
		}
		repoID, err := resolveRepoID()
		if err != nil || repoID == "" {
			s.sendToolError(id, "failed to resolve repository ID")
			return
		}
		var impactReport interface{}
		switch model.NodeType(strings.ToUpper(nodeType)) {
		case model.NodeTypeDecision:
			impactReport, err = s.service.GraphImpactService.AnalyzeDecisionImpact(ctx, repoID, entityID)
		case model.NodeTypeFact:
			impactReport, err = s.service.GraphImpactService.AnalyzeFactImpact(ctx, repoID, entityID)
		case model.NodeTypeEvent:
			impactReport, err = s.service.GraphImpactService.AnalyzeEventImpact(ctx, repoID, entityID)
		case model.NodeTypeContributor:
			impactReport, err = s.service.GraphImpactService.AnalyzeContributorImpact(ctx, repoID, entityID)
		default:
			s.sendToolError(id, fmt.Sprintf("unsupported node_type %q: must be one of DECISION, FACT, EVENT, CONTRIBUTOR", nodeType))
			return
		}
		if err != nil {
			s.sendToolError(id, fmt.Sprintf("impact analysis failed: %v", err))
			return
		}
		s.sendToolSuccess(id, impactReport)

	default:
		if s.handleDevelopmentTool(ctx, id, params.Name, getArg, resolveRepoID) {
			return
		}
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

	case "get_contributor", "trace_contributor":
		schema.Properties["contributor_id"] = map[string]interface{}{
			"type":        "string",
			"description": "The unique identifier of the contributor",
		}
		schema.Required = []string{"contributor_id"}
		schema.Properties["repository_id"] = map[string]interface{}{
			"type":        "string",
			"description": "Optional repository filter",
		}

	case "recommend_reviewers":
		schema.Properties["repository_id"] = map[string]interface{}{
			"type":        "string",
			"description": "The unique identifier of the repository",
		}
		schema.Properties["recommendation_type"] = map[string]interface{}{
			"type":        "string",
			"description": "The type of reviewer recommendation: repository, domain, or impact",
		}
		schema.Properties["domain"] = map[string]interface{}{
			"type":        "string",
			"description": "The knowledge domain to recommend reviewers for (required for domain type)",
		}
		schema.Properties["entity_id"] = map[string]interface{}{
			"type":        "string",
			"description": "The entity ID to analyze impact for (required for impact type)",
		}
		schema.Required = []string{"repository_id", "recommendation_type"}

	case "discover_knowledge":
		schema.Properties["repository_id"] = map[string]interface{}{
			"type":        "string",
			"description": "The unique identifier of the repository",
		}
		schema.Required = []string{"repository_id"}

	case "generate_learning_path":
		schema.Properties["repository_id"] = map[string]interface{}{
			"type":        "string",
			"description": "The unique identifier of the repository",
		}
		schema.Properties["path_type"] = map[string]interface{}{
			"type":        "string",
			"description": "The learning path type (repository, domain, or contributor)",
		}
		schema.Properties["domain"] = map[string]interface{}{
			"type":        "string",
			"description": "The domain name (required for domain path type)",
		}
		schema.Properties["contributor_id"] = map[string]interface{}{
			"type":        "string",
			"description": "The contributor ID (required for contributor path type)",
		}
		schema.Required = []string{"repository_id", "path_type"}

	case "generate_change_plan":
		schema.Properties["repository_id"] = map[string]interface{}{
			"type":        "string",
			"description": "The unique identifier of the repository",
		}
		schema.Properties["entity_type"] = map[string]interface{}{
			"type":        "string",
			"description": "The type of entity being changed (decision, fact, event, or contributor)",
		}
		schema.Properties["entity_id"] = map[string]interface{}{
			"type":        "string",
			"description": "The ID of the entity being changed",
		}
		schema.Required = []string{"repository_id", "entity_type", "entity_id"}

	case "list_decisions", "list_events", "list_intents", "list_facts", "generate_context", "export_context", "list_contributors", "list_expertise":
		schema.Properties["repository_id"] = map[string]interface{}{
			"type":        "string",
			"description": "Optional repository filter",
		}

	case "trace_graph", "find_dependencies", "find_dependents":
		schema.Properties["node_id"] = map[string]interface{}{
			"type":        "string",
			"description": "The graph node ID to start traversal from",
		}
		schema.Required = []string{"node_id"}
		schema.Properties["repository_id"] = map[string]interface{}{
			"type":        "string",
			"description": "Optional repository filter",
		}
		if toolName == "trace_graph" {
			schema.Properties["max_depth"] = map[string]interface{}{
				"type":        "string",
				"description": "Maximum traversal depth (default: 10)",
			}
		}

	case "trace_path":
		schema.Properties["start_node_id"] = map[string]interface{}{
			"type":        "string",
			"description": "The graph node ID to start from",
		}
		schema.Properties["end_node_id"] = map[string]interface{}{
			"type":        "string",
			"description": "The graph node ID to find paths to",
		}
		schema.Required = []string{"start_node_id", "end_node_id"}
		schema.Properties["repository_id"] = map[string]interface{}{
			"type":        "string",
			"description": "Optional repository filter",
		}

	case "analyze_impact":
		schema.Properties["node_id"] = map[string]interface{}{
			"type":        "string",
			"description": "The entity ID to analyze impact for (e.g. decision ID, fact ID)",
		}
		schema.Properties["node_type"] = map[string]interface{}{
			"type":        "string",
			"description": "The entity type: DECISION, FACT, EVENT, or CONTRIBUTOR",
		}
		schema.Required = []string{"node_id", "node_type"}
		schema.Properties["repository_id"] = map[string]interface{}{
			"type":        "string",
			"description": "Optional repository filter",
		}

	case "ask":
		schema.Properties["question"] = map[string]interface{}{
			"type":        "string",
			"description": "The repository or development question to answer",
		}
		schema.Required = []string{"question"}
		schema.Properties["repository_id"] = map[string]interface{}{
			"type":        "string",
			"description": "Optional repository filter",
		}

	case "explain", "review":
		schema.Properties["topic"] = map[string]interface{}{
			"type":        "string",
			"description": "The topic to explain or prepare a review guide for",
		}
		schema.Required = []string{"topic"}
		schema.Properties["repository_id"] = map[string]interface{}{
			"type":        "string",
			"description": "Optional repository filter",
		}

	case "explain_file":
		schema.Properties["file_path"] = map[string]interface{}{
			"type":        "string",
			"description": "Repository-relative file path to explain",
		}
		schema.Required = []string{"file_path"}
		schema.Properties["repository_id"] = map[string]interface{}{
			"type":        "string",
			"description": "Optional repository filter",
		}

	case "explain_function", "explain_struct", "explain_interface", "explain_type":
		schema.Properties["symbol"] = map[string]interface{}{
			"type":        "string",
			"description": "The code symbol name to explain",
		}
		schema.Required = []string{"symbol"}
		schema.Properties["package_path"] = map[string]interface{}{
			"type":        "string",
			"description": "Optional Go package path to disambiguate short symbol names",
		}
		schema.Properties["repository_id"] = map[string]interface{}{
			"type":        "string",
			"description": "Optional repository filter",
		}

	case "plan":
		schema.Properties["task"] = map[string]interface{}{
			"type":        "string",
			"description": "The development task to plan",
		}
		schema.Required = []string{"task"}
		schema.Properties["repository_id"] = map[string]interface{}{
			"type":        "string",
			"description": "Optional repository filter",
		}

	case "analyze_topic_impact":
		schema.Properties["subject"] = map[string]interface{}{
			"type":        "string",
			"description": "The topic, symbol, or area to analyze impact for",
		}
		schema.Required = []string{"subject"}
		schema.Properties["repository_id"] = map[string]interface{}{
			"type":        "string",
			"description": "Optional repository filter",
		}

	case "onboard":
		schema.Properties["assignment"] = map[string]interface{}{
			"type":        "string",
			"description": "Optional first assignment or pasted task to plan",
		}
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
		fmt.Fprintf(os.Stderr, "reponerve mcp: marshal response: %v\n", err)
		return
	}
	if _, err := s.stdout.Write(append(data, '\n')); err != nil {
		fmt.Fprintf(os.Stderr, "reponerve mcp: write response: %v\n", err)
	}
}
