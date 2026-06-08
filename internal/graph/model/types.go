// Package model defines foundational Knowledge Graph abstractions.
//
// It provides graph nodes and graph edges but does not perform:
//
// - Relationship generation
// - Graph traversal
// - Impact analysis
// - Persistence
//
// These capabilities are implemented in later milestones.
package model

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
)

// NodeType represents the type of entity wrapped by a graph node.
type NodeType string

const (
	NodeTypeIntent       NodeType = "INTENT"
	NodeTypeDecision     NodeType = "DECISION"
	NodeTypeFact         NodeType = "FACT"
	NodeTypeEvent        NodeType = "EVENT"
	NodeTypeContributor  NodeType = "CONTRIBUTOR"
	NodeTypeExpertise    NodeType = "EXPERTISE"
)

// RelationshipCategory indicates whether a connection is explicitly stored (fact) or derived (conclusion).
type RelationshipCategory string

const (
	CategoryStored RelationshipCategory = "STORED"
	CategoryDerived RelationshipCategory = "DERIVED"
)

// EdgeType represents the connection type between two graph nodes.
type EdgeType string

// GraphNode wraps an existing repository entity.
//
// Graph nodes reference canonical repository entities.
// They do not duplicate repository entity structures.
//
// Examples:
//
// NodeType=DECISION, EntityID=<decision-id>
//
// NodeType=FACT, EntityID=<fact-id>
//
// The Memory Engine remains the source of truth.
type GraphNode struct {
	ID           string   `json:"id"`
	RepositoryID string   `json:"repository_id"`
	NodeType     NodeType `json:"node_type"`
	EntityID     string   `json:"entity_id"`
}

// GraphEdge represents a connection between two graph nodes.
type GraphEdge struct {
	ID           string               `json:"id"`
	RepositoryID string               `json:"repository_id"`
	FromNodeID   string               `json:"from_node_id"`
	ToNodeID     string               `json:"to_node_id"`
	EdgeType     EdgeType             `json:"edge_type"`
	Category     RelationshipCategory `json:"category"`
	// EvidenceJSON stores evidence supporting the edge.
	// ISSUE-042 validates JSON syntax only.
	// Evidence meaning is validated by future graph relationship engines.
	EvidenceJSON string               `json:"evidence_json"`
}

// NodeID computes a deterministic ID for a GraphNode.
func NodeID(repositoryID string, nodeType NodeType, entityID string) string {
	h := sha256.Sum256([]byte(fmt.Sprintf("%s:%s:%s", repositoryID, nodeType, entityID)))
	return "nod_" + hex.EncodeToString(h[:])
}

// EdgeID computes a deterministic ID for a GraphEdge.
func EdgeID(repositoryID string, fromNodeID string, toNodeID string, edgeType EdgeType) string {
	h := sha256.Sum256([]byte(fmt.Sprintf("%s:%s:%s:%s", repositoryID, fromNodeID, toNodeID, edgeType)))
	return "edg_" + hex.EncodeToString(h[:])
}

// NewNode constructs a new GraphNode with a deterministic ID.
func NewNode(repositoryID string, nodeType NodeType, entityID string) *GraphNode {
	return &GraphNode{
		ID:           NodeID(repositoryID, nodeType, entityID),
		RepositoryID: repositoryID,
		NodeType:     nodeType,
		EntityID:     entityID,
	}
}

// NewEdge constructs a new GraphEdge with a deterministic ID.
func NewEdge(
	repositoryID string,
	fromNodeID string,
	toNodeID string,
	edgeType EdgeType,
	category RelationshipCategory,
	evidenceJSON string,
) *GraphEdge {
	return &GraphEdge{
		ID:           EdgeID(repositoryID, fromNodeID, toNodeID, edgeType),
		RepositoryID: repositoryID,
		FromNodeID:   fromNodeID,
		ToNodeID:     toNodeID,
		EdgeType:     edgeType,
		Category:     category,
		EvidenceJSON: evidenceJSON,
	}
}

// ValidateNode validates the node's properties.
// It verifies that IDs, node types, and entity references are non-empty.
func ValidateNode(node *GraphNode) error {
	if node == nil {
		return fmt.Errorf("node is nil")
	}
	if node.ID == "" {
		return fmt.Errorf("missing node ID")
	}
	if node.RepositoryID == "" {
		return fmt.Errorf("missing repository ID")
	}
	if node.NodeType == "" {
		return fmt.Errorf("missing node type")
	}
	if node.EntityID == "" {
		return fmt.Errorf("missing entity ID")
	}
	return nil
}

// ValidateEdge validates the edge's properties.
// It verifies that IDs, nodes, categories, and evidence are non-empty, and that category and evidence formats are valid.
func ValidateEdge(edge *GraphEdge) error {
	if edge == nil {
		return fmt.Errorf("edge is nil")
	}
	if edge.ID == "" {
		return fmt.Errorf("missing edge ID")
	}
	if edge.RepositoryID == "" {
		return fmt.Errorf("missing repository ID")
	}
	if edge.FromNodeID == "" {
		return fmt.Errorf("missing from node ID")
	}
	if edge.ToNodeID == "" {
		return fmt.Errorf("missing to node ID")
	}
	if edge.EdgeType == "" {
		return fmt.Errorf("missing edge type")
	}
	switch edge.Category {
	case CategoryStored, CategoryDerived:
		// Valid
	case "":
		return fmt.Errorf("missing relationship category")
	default:
		return fmt.Errorf("invalid relationship category: %q", edge.Category)
	}
	if edge.EvidenceJSON == "" {
		return fmt.Errorf("missing evidence")
	}
	if !json.Valid([]byte(edge.EvidenceJSON)) {
		return fmt.Errorf("evidence must be valid JSON")
	}
	return nil
}
