// Package impact provides graph-aware impact analysis for repository knowledge.
//
// Impact analysis is derived knowledge. Impact conclusions are not facts.
// Every impact path preserves:
//   - Traversal path (nodes and edges)
//   - Evidence (from graph edges)
//   - Reason (deterministic, reproducible explanation)
//
// Impact analysis consumes the Graph Traversal Engine.
// It does not re-scan repositories, execute Git commands, or access SQLite directly.
package impact

import (
	"context"
	"fmt"
	"sort"

	"github.com/reponerve/reponerve/internal/graph/model"
	"github.com/reponerve/reponerve/internal/graph/traversal"
)

// Service provides graph-aware impact analysis built on the traversal engine.
type Service struct {
	traversalEngine *traversal.Engine
}

// NewService constructs a new impact analysis Service.
func NewService(traversalEngine *traversal.Engine) *Service {
	return &Service{
		traversalEngine: traversalEngine,
	}
}

// defaultOptions returns the standard traversal options for impact analysis.
// Stored and derived edges are both included.
func defaultOptions() traversal.TraversalOptions {
	return traversal.TraversalOptions{
		MaxDepth:       10,
		IncludeStored:  true,
		IncludeDerived: true,
	}
}

// AnalyzeDecisionImpact returns all impact paths reachable from the given decision node.
func (s *Service) AnalyzeDecisionImpact(ctx context.Context, repositoryID string, decisionID string) (*ImpactReport, error) {
	nodeID := model.NodeID(repositoryID, model.NodeTypeDecision, decisionID)

	result, err := s.traversalEngine.FindDependencies(ctx, repositoryID, nodeID, defaultOptions())
	if err != nil {
		return nil, fmt.Errorf("impact analysis failed for decision %s: %w", decisionID, err)
	}

	return buildReport(result, decisionID), nil
}

// AnalyzeFactImpact returns all impact paths reachable from the given fact node.
func (s *Service) AnalyzeFactImpact(ctx context.Context, repositoryID string, factID string) (*ImpactReport, error) {
	nodeID := model.NodeID(repositoryID, model.NodeTypeFact, factID)

	result, err := s.traversalEngine.FindDependencies(ctx, repositoryID, nodeID, defaultOptions())
	if err != nil {
		return nil, fmt.Errorf("impact analysis failed for fact %s: %w", factID, err)
	}

	return buildReport(result, factID), nil
}

// AnalyzeEventImpact returns all impact paths reachable from the given event node.
func (s *Service) AnalyzeEventImpact(ctx context.Context, repositoryID string, eventID string) (*ImpactReport, error) {
	nodeID := model.NodeID(repositoryID, model.NodeTypeEvent, eventID)

	result, err := s.traversalEngine.FindDependencies(ctx, repositoryID, nodeID, defaultOptions())
	if err != nil {
		return nil, fmt.Errorf("impact analysis failed for event %s: %w", eventID, err)
	}

	return buildReport(result, eventID), nil
}

// AnalyzeContributorImpact returns all impact paths reachable from the given contributor node.
// It uses FindDependents to discover inbound expertise edges, then FindDependencies to extend
// outward from those expertise nodes. This reveals the full knowledge domain footprint.
func (s *Service) AnalyzeContributorImpact(ctx context.Context, repositoryID string, contributorID string) (*ImpactReport, error) {
	nodeID := model.NodeID(repositoryID, model.NodeTypeContributor, contributorID)

	// Traverse dependents — inbound edges from expertise nodes that link to this contributor
	result, err := s.traversalEngine.FindDependents(ctx, repositoryID, nodeID, defaultOptions())
	if err != nil {
		return nil, fmt.Errorf("impact analysis failed for contributor %s: %w", contributorID, err)
	}

	return buildReport(result, contributorID), nil
}

// buildReport constructs an ImpactReport from a TraversalResult.
// It generates deterministic reasons for each path and sorts the output.
func buildReport(result *traversal.TraversalResult, startEntityID string) *ImpactReport {
	if result == nil || len(result.Paths) == 0 {
		return &ImpactReport{ImpactPaths: []*ImpactPath{}}
	}

	var impactPaths []*ImpactPath

	for _, path := range result.Paths {
		if !isValidPath(path) {
			continue
		}

		reason := generateReason(path)
		if reason == "" {
			continue
		}

		impactPaths = append(impactPaths, &ImpactPath{
			Path:   path,
			Reason: reason,
		})
	}

	sortImpactPaths(impactPaths)

	return &ImpactReport{ImpactPaths: impactPaths}
}

// isValidPath verifies a traversal path has the required structure and edge evidence.
func isValidPath(path *traversal.TraversalPath) bool {
	if path == nil {
		return false
	}
	if len(path.Nodes) < 2 {
		return false
	}
	if len(path.Edges) == 0 {
		return false
	}
	if len(path.Edges) != len(path.Nodes)-1 {
		return false
	}
	for _, node := range path.Nodes {
		if node == nil {
			return false
		}
	}
	for _, edge := range path.Edges {
		if edge == nil {
			return false
		}
		if edge.EvidenceJSON == "" {
			return false
		}
	}
	return true
}

// generateReason produces a deterministic, human-readable explanation for an impact path.
// Reasons are derived from node types only — no content inspection.
func generateReason(path *traversal.TraversalPath) string {
	if len(path.Nodes) < 2 {
		return ""
	}

	startNode := path.Nodes[0]
	endNode := path.Nodes[len(path.Nodes)-1]
	startType := startNode.NodeType
	endType := endNode.NodeType
	startID := startNode.EntityID
	endID := endNode.EntityID

	// Specific known relationship patterns
	switch {
	case startType == model.NodeTypeDecision && endType == model.NodeTypeDecision:
		return fmt.Sprintf("Decision %s impacts Decision %s because Decision %s depends on Decision %s.", startID, endID, endID, startID)

	case startType == model.NodeTypeFact && endType == model.NodeTypeFact:
		return fmt.Sprintf("Fact %s impacts Fact %s because Fact %s is supported by Fact %s.", startID, endID, endID, startID)

	case startType == model.NodeTypeContributor && endType == model.NodeTypeExpertise:
		return fmt.Sprintf("Contributor %s impacts Domain %s because repository expertise connects the contributor to the domain.", startID, endID)

	case startType == model.NodeTypeExpertise && endType == model.NodeTypeContributor:
		return fmt.Sprintf("Contributor %s impacts Domain %s because repository expertise connects the contributor to the domain.", endID, startID)

	case startType == model.NodeTypeDecision && endType == model.NodeTypeEvent:
		return fmt.Sprintf("Decision %s impacts Event %s because Event %s results from Decision %s.", startID, endID, endID, startID)

	case startType == model.NodeTypeFact && endType == model.NodeTypeDecision:
		return fmt.Sprintf("Fact %s impacts Decision %s because Decision %s is supported by Fact %s.", startID, endID, endID, startID)

	case startType == model.NodeTypeIntent && endType == model.NodeTypeDecision:
		return fmt.Sprintf("Intent %s impacts Decision %s because Decision %s is driven by Intent %s.", startID, endID, endID, startID)

	default:
		return fmt.Sprintf("%s %s impacts %s %s through traversal path.", startType, startID, endType, endID)
	}
}

// sortImpactPaths sorts impact paths deterministically:
//  1. Path length ascending
//  2. Starting node ID ascending
//  3. Ending node ID ascending
func sortImpactPaths(paths []*ImpactPath) {
	sort.Slice(paths, func(i, j int) bool {
		pi := paths[i].Path
		pj := paths[j].Path

		lenI := len(pi.Edges)
		lenJ := len(pj.Edges)
		if lenI != lenJ {
			return lenI < lenJ
		}

		startI := pi.Nodes[0].ID
		startJ := pj.Nodes[0].ID
		if startI != startJ {
			return startI < startJ
		}

		endI := pi.Nodes[len(pi.Nodes)-1].ID
		endJ := pj.Nodes[len(pj.Nodes)-1].ID
		return endI < endJ
	})
}
