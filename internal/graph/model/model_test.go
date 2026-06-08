package model

import (
	"testing"
)

func TestNodeCreation(t *testing.T) {
	repoID := "repo_abc"
	nodeType := NodeTypeDecision
	entityID := "dec_123"

	node := NewNode(repoID, nodeType, entityID)

	if node.RepositoryID != repoID {
		t.Errorf("expected RepositoryID %q, got %q", repoID, node.RepositoryID)
	}
	if node.NodeType != nodeType {
		t.Errorf("expected NodeType %q, got %q", nodeType, node.NodeType)
	}
	if node.EntityID != entityID {
		t.Errorf("expected EntityID %q, got %q", entityID, node.EntityID)
	}

	expectedID := NodeID(repoID, nodeType, entityID)
	if node.ID != expectedID {
		t.Errorf("expected ID %q, got %q", expectedID, node.ID)
	}
}

func TestEdgeCreation(t *testing.T) {
	repoID := "repo_abc"
	fromID := "nod_from"
	toID := "nod_to"
	edgeType := EdgeType("DECISION_DEPENDS_ON_DECISION")
	category := CategoryDerived
	evidence := `{"reason":"direct reference"}`

	edge := NewEdge(repoID, fromID, toID, edgeType, category, evidence)

	if edge.RepositoryID != repoID {
		t.Errorf("expected RepositoryID %q, got %q", repoID, edge.RepositoryID)
	}
	if edge.FromNodeID != fromID {
		t.Errorf("expected FromNodeID %q, got %q", fromID, edge.FromNodeID)
	}
	if edge.ToNodeID != toID {
		t.Errorf("expected ToNodeID %q, got %q", toID, edge.ToNodeID)
	}
	if edge.EdgeType != edgeType {
		t.Errorf("expected EdgeType %q, got %q", edgeType, edge.EdgeType)
	}
	if edge.Category != category {
		t.Errorf("expected Category %q, got %q", category, edge.Category)
	}
	if edge.EvidenceJSON != evidence {
		t.Errorf("expected EvidenceJSON %q, got %q", evidence, edge.EvidenceJSON)
	}

	expectedID := EdgeID(repoID, fromID, toID, edgeType)
	if edge.ID != expectedID {
		t.Errorf("expected ID %q, got %q", expectedID, edge.ID)
	}
}

func TestValidateNode(t *testing.T) {
	t.Run("nil node", func(t *testing.T) {
		err := ValidateNode(nil)
		if err == nil || err.Error() != "node is nil" {
			t.Errorf("expected 'node is nil' error, got %v", err)
		}
	})

	t.Run("valid node", func(t *testing.T) {
		node := NewNode("repo_1", NodeTypeFact, "fact_1")
		if err := ValidateNode(node); err != nil {
			t.Errorf("unexpected error for valid node: %v", err)
		}
	})

	t.Run("missing ID", func(t *testing.T) {
		node := &GraphNode{
			RepositoryID: "repo_1",
			NodeType:     NodeTypeFact,
			EntityID:     "fact_1",
		}
		err := ValidateNode(node)
		if err == nil || err.Error() != "missing node ID" {
			t.Errorf("expected 'missing node ID' error, got %v", err)
		}
	})

	t.Run("missing repository ID", func(t *testing.T) {
		node := &GraphNode{
			ID:       "nod_1",
			NodeType: NodeTypeFact,
			EntityID: "fact_1",
		}
		err := ValidateNode(node)
		if err == nil || err.Error() != "missing repository ID" {
			t.Errorf("expected 'missing repository ID' error, got %v", err)
		}
	})

	t.Run("missing node type", func(t *testing.T) {
		node := &GraphNode{
			ID:           "nod_1",
			RepositoryID: "repo_1",
			EntityID:     "fact_1",
		}
		err := ValidateNode(node)
		if err == nil || err.Error() != "missing node type" {
			t.Errorf("expected 'missing node type' error, got %v", err)
		}
	})

	t.Run("missing entity ID", func(t *testing.T) {
		node := &GraphNode{
			ID:           "nod_1",
			RepositoryID: "repo_1",
			NodeType:     NodeTypeFact,
		}
		err := ValidateNode(node)
		if err == nil || err.Error() != "missing entity ID" {
			t.Errorf("expected 'missing entity ID' error, got %v", err)
		}
	})
}

func TestValidateEdge(t *testing.T) {
	t.Run("nil edge", func(t *testing.T) {
		err := ValidateEdge(nil)
		if err == nil || err.Error() != "edge is nil" {
			t.Errorf("expected 'edge is nil' error, got %v", err)
		}
	})

	t.Run("valid edge", func(t *testing.T) {
		edge := NewEdge("repo_1", "nod_1", "nod_2", "TEST_TYPE", CategoryStored, `{"a": 1}`)
		if err := ValidateEdge(edge); err != nil {
			t.Errorf("unexpected error for valid edge: %v", err)
		}
	})

	t.Run("missing ID", func(t *testing.T) {
		edge := &GraphEdge{
			RepositoryID: "repo_1",
			FromNodeID:   "nod_1",
			ToNodeID:     "nod_2",
			EdgeType:     "TEST_TYPE",
			Category:     CategoryStored,
			EvidenceJSON: `{"a": 1}`,
		}
		err := ValidateEdge(edge)
		if err == nil || err.Error() != "missing edge ID" {
			t.Errorf("expected 'missing edge ID' error, got %v", err)
		}
	})

	t.Run("missing repository ID", func(t *testing.T) {
		edge := &GraphEdge{
			ID:           "edg_1",
			FromNodeID:   "nod_1",
			ToNodeID:     "nod_2",
			EdgeType:     "TEST_TYPE",
			Category:     CategoryStored,
			EvidenceJSON: `{"a": 1}`,
		}
		err := ValidateEdge(edge)
		if err == nil || err.Error() != "missing repository ID" {
			t.Errorf("expected 'missing repository ID' error, got %v", err)
		}
	})

	t.Run("missing from node ID", func(t *testing.T) {
		edge := &GraphEdge{
			ID:           "edg_1",
			RepositoryID: "repo_1",
			ToNodeID:     "nod_2",
			EdgeType:     "TEST_TYPE",
			Category:     CategoryStored,
			EvidenceJSON: `{"a": 1}`,
		}
		err := ValidateEdge(edge)
		if err == nil || err.Error() != "missing from node ID" {
			t.Errorf("expected 'missing from node ID' error, got %v", err)
		}
	})

	t.Run("missing to node ID", func(t *testing.T) {
		edge := &GraphEdge{
			ID:           "edg_1",
			RepositoryID: "repo_1",
			FromNodeID:   "nod_1",
			EdgeType:     "TEST_TYPE",
			Category:     CategoryStored,
			EvidenceJSON: `{"a": 1}`,
		}
		err := ValidateEdge(edge)
		if err == nil || err.Error() != "missing to node ID" {
			t.Errorf("expected 'missing to node ID' error, got %v", err)
		}
	})

	t.Run("missing edge type", func(t *testing.T) {
		edge := &GraphEdge{
			ID:           "edg_1",
			RepositoryID: "repo_1",
			FromNodeID:   "nod_1",
			ToNodeID:     "nod_2",
			Category:     CategoryStored,
			EvidenceJSON: `{"a": 1}`,
		}
		err := ValidateEdge(edge)
		if err == nil || err.Error() != "missing edge type" {
			t.Errorf("expected 'missing edge type' error, got %v", err)
		}
	})

	t.Run("missing category", func(t *testing.T) {
		edge := &GraphEdge{
			ID:           "edg_1",
			RepositoryID: "repo_1",
			FromNodeID:   "nod_1",
			ToNodeID:     "nod_2",
			EdgeType:     "TEST_TYPE",
			EvidenceJSON: `{"a": 1}`,
		}
		err := ValidateEdge(edge)
		if err == nil || err.Error() != "missing relationship category" {
			t.Errorf("expected 'missing relationship category' error, got %v", err)
		}
	})

	t.Run("invalid category", func(t *testing.T) {
		edge := &GraphEdge{
			ID:           "edg_1",
			RepositoryID: "repo_1",
			FromNodeID:   "nod_1",
			ToNodeID:     "nod_2",
			EdgeType:     "TEST_TYPE",
			Category:     RelationshipCategory("INVALID"),
			EvidenceJSON: `{"a": 1}`,
		}
		err := ValidateEdge(edge)
		if err == nil || err.Error() != `invalid relationship category: "INVALID"` {
			t.Errorf("expected invalid category error, got %v", err)
		}
	})

	t.Run("missing evidence", func(t *testing.T) {
		edge := &GraphEdge{
			ID:           "edg_1",
			RepositoryID: "repo_1",
			FromNodeID:   "nod_1",
			ToNodeID:     "nod_2",
			EdgeType:     "TEST_TYPE",
			Category:     CategoryStored,
		}
		err := ValidateEdge(edge)
		if err == nil || err.Error() != "missing evidence" {
			t.Errorf("expected 'missing evidence' error, got %v", err)
		}
	})

	t.Run("invalid evidence JSON syntax", func(t *testing.T) {
		edge := &GraphEdge{
			ID:           "edg_1",
			RepositoryID: "repo_1",
			FromNodeID:   "nod_1",
			ToNodeID:     "nod_2",
			EdgeType:     "TEST_TYPE",
			Category:     CategoryStored,
			EvidenceJSON: `{"invalid": `,
		}
		err := ValidateEdge(edge)
		if err == nil || err.Error() != "evidence must be valid JSON" {
			t.Errorf("expected 'evidence must be valid JSON' error, got %v", err)
		}
	})
}

func TestDeterministicIDsAndCollisionPrevention(t *testing.T) {
	t.Run("node IDs are deterministic", func(t *testing.T) {
		id1 := NodeID("repo_1", NodeTypeDecision, "123")
		id2 := NodeID("repo_1", NodeTypeDecision, "123")
		if id1 != id2 {
			t.Errorf("expected node IDs to be identical, got %q and %q", id1, id2)
		}
	})

	t.Run("node type changes generate different node IDs", func(t *testing.T) {
		decID := NodeID("repo_1", NodeTypeDecision, "123")
		factID := NodeID("repo_1", NodeTypeFact, "123")
		if decID == factID {
			t.Errorf("expected different node IDs for DECISION and FACT but got identical %q", decID)
		}
	})

	t.Run("edge IDs are deterministic", func(t *testing.T) {
		id1 := EdgeID("repo_1", "nod_1", "nod_2", "TEST_TYPE")
		id2 := EdgeID("repo_1", "nod_1", "nod_2", "TEST_TYPE")
		if id1 != id2 {
			t.Errorf("expected edge IDs to be identical, got %q and %q", id1, id2)
		}
	})

	t.Run("edge type changes generate different edge IDs", func(t *testing.T) {
		id1 := EdgeID("repo_1", "nod_1", "nod_2", "TYPE_A")
		id2 := EdgeID("repo_1", "nod_1", "nod_2", "TYPE_B")
		if id1 == id2 {
			t.Errorf("expected different edge IDs for TYPE_A and TYPE_B but got identical %q", id1)
		}
	})

	t.Run("node IDs start with nod_ prefix", func(t *testing.T) {
		id := NodeID("repo_1", NodeTypeDecision, "123")
		if id[:4] != "nod_" {
			t.Errorf("expected node ID to start with 'nod_', got %q", id)
		}
	})

	t.Run("edge IDs start with edg_ prefix", func(t *testing.T) {
		id := EdgeID("repo_1", "nod_1", "nod_2", "TEST_TYPE")
		if id[:4] != "edg_" {
			t.Errorf("expected edge ID to start with 'edg_', got %q", id)
		}
	})
}
