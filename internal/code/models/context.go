package models

// EvidenceItem is a traceable evidence reference in code intelligence output.
type EvidenceItem struct {
	Source string `json:"source"`
	Detail string `json:"detail"`
}

// CallGraphEdge is one directed call relationship in a call graph.
type CallGraphEdge struct {
	FromEntityID string
	ToEntityID   string
	Relationship *CodeRelationship
}

// CallGraph is a deterministic call graph rooted at one symbol entity.
type CallGraph struct {
	RootEntityID string
	Edges        []*CallGraphEdge
}

// SymbolDependencyReport lists outbound symbol dependencies with evidence.
type SymbolDependencyReport struct {
	RootEntity   *CodeEntity
	Dependencies []*CodeRelationship
}

// CodeExplanationContext is the authoritative code context for a file or symbol.
type CodeExplanationContext struct {
	Subject      string
	Modules      []*CodeEntity
	Packages     []*CodeEntity
	Files        []*CodeEntity
	Structs      []*CodeEntity
	Interfaces   []*CodeEntity
	TypeAliases  []*CodeEntity
	Functions    []*CodeEntity
	Methods      []*CodeEntity
	Endpoints    []*CodeEntity
	CallGraph    *CallGraph
	Dependencies []*CodeRelationship
	Evidence     []EvidenceItem
}
