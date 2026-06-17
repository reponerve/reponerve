package server

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/reponerve/reponerve/internal/agent/sessionmemory"
	"github.com/reponerve/reponerve/internal/graph/communities"
	graphdiscovery "github.com/reponerve/reponerve/internal/graph/discovery"
	"github.com/reponerve/reponerve/internal/graph/traversal"
)

func isGraphTool(name string) bool {
	switch name {
	case "discover_surprises", "suggest_questions", "query_graph":
		return true
	default:
		return false
	}
}

func isSessionTool(name string) bool {
	switch name {
	case "remember", "forget":
		return true
	default:
		return false
	}
}

func (s *Server) handleGraphTool(
	ctx context.Context,
	id *json.RawMessage,
	toolName string,
	getArg func(string, bool) (string, error),
	resolveRepoID func() (string, error),
) bool {
	if !isGraphTool(toolName) {
		return false
	}
	if s.service.GraphTraversalEngine == nil {
		s.sendToolError(id, "graph traversal engine is not configured")
		return true
	}

	repoID, err := resolveRepoID()
	if err != nil || repoID == "" {
		s.sendToolError(id, "failed to resolve repository ID")
		return true
	}

	engine := s.service.GraphTraversalEngine
	snapshot, err := engine.LoadGraphSnapshot(ctx, repoID, traversal.TraversalOptions{
		IncludeStored:  true,
		IncludeDerived: true,
	})
	if err != nil {
		s.sendToolError(id, fmt.Sprintf("failed to load graph: %v", err))
		return true
	}
	communityResult := communities.Detect(repoID, snapshot.Nodes, snapshot.Edges)
	report, err := graphdiscovery.Analyze(repoID, snapshot.Nodes, snapshot.Edges, communityResult)
	if err != nil {
		s.sendToolError(id, fmt.Sprintf("graph analysis failed: %v", err))
		return true
	}

	switch toolName {
	case "discover_surprises":
		s.sendToolSuccess(id, report)
	case "suggest_questions":
		s.sendToolSuccess(id, map[string]any{
			"repository_id": repoID,
			"questions":     report.SuggestedQuestions,
		})
	case "query_graph":
		startNode, err := getArg("start_node_id", true)
		if err != nil {
			s.sendToolError(id, err.Error())
			return true
		}
		budget := 500
		if budgetStr, _ := getArg("token_budget", false); strings.TrimSpace(budgetStr) != "" {
			if n, err := strconv.Atoi(strings.TrimSpace(budgetStr)); err == nil {
				budget = n
			}
		}
		result, err := engine.TraverseWithBudget(ctx, repoID, startNode, traversal.BudgetTraversalOptions{
			TraversalOptions: traversal.TraversalOptions{IncludeStored: true, IncludeDerived: true},
			TokenBudget:      budget,
		})
		if err != nil {
			s.sendToolError(id, fmt.Sprintf("query_graph failed: %v", err))
			return true
		}
		s.sendToolSuccess(id, result)
	}
	return true
}

func (s *Server) handleSessionTool(
	ctx context.Context,
	id *json.RawMessage,
	toolName string,
	getArg func(string, bool) (string, error),
	resolveRepoID func() (string, error),
) bool {
	if !isSessionTool(toolName) {
		return false
	}
	if s.service.SessionMemoryService == nil {
		s.sendToolError(id, "session memory service is not configured")
		return true
	}
	repoID, err := resolveRepoID()
	if err != nil || repoID == "" {
		s.sendToolError(id, "failed to resolve repository ID")
		return true
	}
	svc := s.service.SessionMemoryService

	switch toolName {
	case "remember":
		subject, _ := getArg("subject", false)
		content, err := getArg("content", true)
		if err != nil {
			s.sendToolError(id, err.Error())
			return true
		}
		if strings.TrimSpace(subject) == "" {
			subject = "session"
		}
		fact, err := svc.Remember(ctx, sessionmemory.RememberRequest{
			RepositoryID: repoID,
			Subject:      subject,
			Content:      content,
		})
		if err != nil {
			s.sendToolError(id, fmt.Sprintf("remember failed: %v", err))
			return true
		}
		s.sendToolSuccess(id, fact)
	case "forget":
		factID, err := getArg("fact_id", true)
		if err != nil {
			s.sendToolError(id, err.Error())
			return true
		}
		if err := svc.Forget(ctx, repoID, factID); err != nil {
			s.sendToolError(id, fmt.Sprintf("forget failed: %v", err))
			return true
		}
		s.sendToolSuccess(id, map[string]string{"status": "forgotten", "fact_id": factID})
	}
	return true
}
