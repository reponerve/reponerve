# ISSUE-043 — Graph Relationship Engine

## Objective

Generate deterministic, evidence-backed derived graph relationships from existing repository knowledge.

---

# Background

The Memory Engine stores repository facts.

The Knowledge Graph introduces derived relationships that connect repository knowledge in meaningful ways.

Derived relationships are conclusions.

Derived relationships are not facts.

Every derived relationship must be explainable and evidence-backed.

---

# Philosophy

Evidence First.

Derived relationships without evidence are invalid.

Stored relationships are facts.

Derived relationships are conclusions.

The same repository state must always generate the same graph relationships.

---

# Scope

Create:

internal/graph/relationships/

Files:

* engine.go
* types.go
* engine_test.go

---

# Architecture Requirements

The Relationship Engine must consume:

* Memory Engine entities
* Existing relationship records
* Ownership Intelligence entities

The Relationship Engine must NOT:

* Re-scan repositories
* Execute Git commands
* Access SQLite directly
* Generate AI-derived relationships

Relationship generation must be deterministic.

---

# Derived Relationship Model

Create:

```go
type DerivedRelationship struct {
    Edge *model.GraphEdge

    Evidence json.RawMessage

    Explanation string
}
```

Explanation must be human-readable.

Example:

Decision A depends on Decision B because Decision A references Decision B in repository memory.

---

# Supported Derived Relationships

## Decision Relationships

DECISION_DEPENDS_ON_DECISION

Represents:

Decision A cannot be understood without Decision B.

Evidence examples:

* Explicit references
* Existing memory links
* Repository metadata

---

## Fact Relationships

FACT_SUPPORTS_FACT

Represents:

Fact A provides supporting evidence for Fact B.

Evidence must identify the supporting chain.

---

## Domain Relationships

DOMAIN_RELATES_TO_DOMAIN

Represents:

Two expertise domains are connected through repository activity.

Evidence must identify contributors and repository activity linking the domains.

---

# Evidence Requirements

Every relationship must include:

Evidence JSON

Human-readable Explanation

GraphEdge EvidenceJSON

All three are required.

Derived relationships without evidence are invalid.

---

# Duplicate Prevention

Relationship generation must be idempotent.

Repeated generation over the same repository state must produce:

* identical relationship IDs
* identical relationship counts
* identical evidence

Duplicate relationships are invalid.

---

# Deterministic Ordering

Output ordering must be stable.

Recommended:

* EdgeType ascending
* FromNodeID ascending
* ToNodeID ascending

Same repository state must always generate identical ordering.

---

# Relationship Generation Rules

Rule 1:

Generate only relationships that can be supported by repository evidence.

Rule 2:

Do not infer speculative relationships.

Rule 3:

Do not generate relationships based on AI reasoning.

Rule 4:

Derived relationships must remain reproducible.

Rule 5:

Evidence must be sufficient to explain relationship existence.

---

# Engine API

Implement:

```go
type Engine struct {
}
```

```go
func NewEngine() *Engine
```

```go
func (e *Engine) Generate(
    ctx context.Context,
    repositoryID string,
) ([]*DerivedRelationship, error)
```

---

# Explainability Requirement

Every relationship must be explainable.

Unsupported:

Decision A depends on Decision B because they seem related.

Supported:

Decision A depends on Decision B because repository memory contains explicit dependency references.

---

# Unit Tests

Cover:

* Empty repositories
* Relationship generation
* Evidence generation
* Explanation generation
* Duplicate prevention
* Deterministic ordering
* Deterministic IDs

---

# Integration Tests

Verify:

Repository Memory
↓
Relationship Engine
↓
Graph Edges

using migration-backed SQLite repositories.

Verify:

* generated relationships
* evidence persistence
* explanation correctness
* deterministic output

---

# Constraints

Do NOT implement:

* Graph traversal
* Impact analysis
* MCP tools

Only implement relationship generation.

---

# Acceptance Criteria

Derived relationships are generated deterministically.

All relationships contain evidence.

All relationships contain explanations.

Duplicate relationships are prevented.

All tests pass.
