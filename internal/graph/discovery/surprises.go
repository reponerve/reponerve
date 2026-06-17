package discovery

import (
	"fmt"
	"sort"

	"github.com/reponerve/reponerve/internal/graph/communities"
	"github.com/reponerve/reponerve/internal/graph/model"
)

const defaultGodNodeMinDegree = 2

// Analyze discovers god nodes and cross-community edges from a graph snapshot.
func Analyze(
	repositoryID string,
	nodes []*model.GraphNode,
	edges []*model.GraphEdge,
	communityResult *communities.DetectionResult,
) (*GraphDiscoveryReport, error) {
	if repositoryID == "" {
		return nil, fmt.Errorf("repository ID is required")
	}
	if communityResult == nil {
		communityResult = communities.Detect(repositoryID, nodes, edges)
	}

	degree := make(map[string]int)
	for _, edge := range edges {
		if edge == nil {
			continue
		}
		degree[edge.FromNodeID]++
		degree[edge.ToNodeID]++
	}

	nodeByID := make(map[string]*model.GraphNode, len(nodes))
	for _, n := range nodes {
		if n != nil {
			nodeByID[n.ID] = n
		}
	}

	godNodes := make([]GodNode, 0)
	for nodeID, d := range degree {
		if d >= defaultGodNodeMinDegree {
			godNodes = append(godNodes, GodNode{NodeID: nodeID, Degree: d})
		}
	}
	sort.Slice(godNodes, func(i, j int) bool {
		if godNodes[i].Degree != godNodes[j].Degree {
			return godNodes[i].Degree > godNodes[j].Degree
		}
		return godNodes[i].NodeID < godNodes[j].NodeID
	})

	nodeCommunity := communities.NodeCommunityMap(communityResult)
	surprises := make([]SurprisingConnection, 0)
	for _, edge := range edges {
		if edge == nil {
			continue
		}
		fromNode := nodeByID[edge.FromNodeID]
		toNode := nodeByID[edge.ToNodeID]
		if fromNode == nil || toNode == nil {
			continue
		}
		fromCom := nodeCommunity[edge.FromNodeID]
		toCom := nodeCommunity[edge.ToNodeID]
		crossCommunity := fromCom != "" && toCom != "" && fromCom != toCom
		crossType := fromNode.NodeType != toNode.NodeType
		if !crossCommunity && !crossType {
			continue
		}
		surprises = append(surprises, SurprisingConnection{
			FromNodeID:    edge.FromNodeID,
			ToNodeID:      edge.ToNodeID,
			EdgeType:      string(edge.EdgeType),
			FromCommunity: fromCom,
			ToCommunity:   toCom,
		})
	}
	sort.Slice(surprises, func(i, j int) bool {
		if surprises[i].FromNodeID != surprises[j].FromNodeID {
			return surprises[i].FromNodeID < surprises[j].FromNodeID
		}
		return surprises[i].ToNodeID < surprises[j].ToNodeID
	})

	questions := suggestQuestions(godNodes, surprises)

	return &GraphDiscoveryReport{
		RepositoryID:          repositoryID,
		GodNodes:              godNodes,
		SurprisingConnections: surprises,
		SuggestedQuestions:    questions,
	}, nil
}

func suggestQuestions(godNodes []GodNode, surprises []SurprisingConnection) []string {
	questions := make([]string, 0, 5)
	if len(godNodes) > 0 {
		questions = append(questions, fmt.Sprintf("Why is %s a hub in the knowledge graph?", godNodes[0].NodeID))
	}
	if len(surprises) > 0 {
		s := surprises[0]
		questions = append(questions, fmt.Sprintf("What links %s and %s across communities?", s.FromNodeID, s.ToNodeID))
	}
	if len(questions) == 0 {
		questions = append(questions, "What are the main architectural decisions in this repository?")
	}
	return questions
}
