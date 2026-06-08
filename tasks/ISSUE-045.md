# ISSUE-045 — Impact Graph Analysis

## Objective

Implement graph-aware impact analysis using Knowledge Graph traversal paths.

---

# Background

Knowledge Graph Intelligence should answer:

What is affected if this changes?

Impact analysis builds on graph traversal and relationship evidence.

Impact conclusions must be explainable.

Impact conclusions must remain evidence-based.

---

# Philosophy

Evidence First.

Impact conclusions are derived knowledge.

Impact conclusions are not facts.

Every impact conclusion must include:

* Supporting path
* Supporting evidence
* Human-readable reasoning

---

# Scope

Create:

internal/graph/impact/

Files:

* types.go
* service.go
* service_test.go

---

# Architecture Requirements

Reuse:

* Knowledge Graph Model
* Graph Relationship Engine
* Graph Traversal Engine

Do NOT:

* Re-scan repositories
* Execute Git commands
* Access SQLite directly
* Generate new graph relationships

Impact analysis consumes graph knowledge.

It does not create graph knowledge.

---

# Impact Models

Implement:

```go id="dzpm0q"
type ImpactPath struct {
    Path *traversal.TraversalPath

    Reason string
}
```

Reason must explain why the path represents impact.

Example:

Decision A impacts Event B because Event B is reachable through a dependency chain originating from Decision A.

---

Implement:

```go id="d8m4n0"
type ImpactReport struct {
    ImpactPaths []*ImpactPath
}
```

Impact reports are collections of explainable impact paths.

---

# Impact Service

Implement:

```go id="u9h3r4"
type Service struct {
    traversalEngine *traversal.Engine
}
```

Constructor:

```go id="x3jpt7"
func NewService(
    traversalEngine *traversal.Engine,
) *Service
```

---

# Impact APIs

Implement:

```go id="mn7t2i"
func (s *Service) AnalyzeDecisionImpact(
    ctx context.Context,
    repositoryID string,
    decisionID string,
) (*ImpactReport, error)
```

---

Implement:

```go id="zvphs5"
func (s *Service) AnalyzeFactImpact(
    ctx context.Context,
    repositoryID string,
    factID string,
) (*ImpactReport, error)
```

---

Implement:

```go id="8m72wa"
func (s *Service) AnalyzeEventImpact(
    ctx context.Context,
    repositoryID string,
    eventID string,
) (*ImpactReport, error)
```

---

Implement:

```go id="2u1knw"
func (s *Service) AnalyzeContributorImpact(
    ctx context.Context,
    repositoryID string,
    contributorID string,
) (*ImpactReport, error)
```

---

# Impact Rules

Impact analysis must:

* Traverse dependencies
* Traverse dependents
* Preserve graph evidence
* Preserve relationship chains

Impact analysis must not:

* Invent relationships
* Use AI reasoning
* Use heuristics without evidence

---

# Impact Reasoning

Every ImpactPath must contain a deterministic reason.

Examples:

Decision Impact:

```text id="l3h9g4"
Decision A impacts Decision B because Decision B depends on Decision A.
```

Fact Impact:

```text id="7c7kec"
Fact A impacts Fact B because Fact B is supported by Fact A.
```

Contributor Impact:

```text id="4svj6z"
Contributor A impacts Domain B because repository expertise connects the contributor to the domain.
```

Reason strings must be deterministic.

---

# Evidence Preservation

Every ImpactPath must preserve:

* Nodes
* Edges
* Edge evidence
* Traversal path

Impact analysis must never discard graph evidence.

---

# Deterministic Ordering

Impact paths must be returned in stable order.

Recommended:

1. Path length ascending
2. Start node ID ascending
3. End node ID ascending

The same repository state must produce identical results.

---

# Validation

Validate:

* Non-nil paths
* Non-empty reasons
* Valid traversal paths
* Preserved evidence

Reject incomplete impact paths.

---

# Unit Tests

Cover:

* Empty repositories
* Decision impact
* Fact impact
* Event impact
* Contributor impact
* Multi-hop impacts
* Deterministic ordering
* Reason generation

---

# Integration Tests

Verify:

Repository Memory
↓
Relationships
↓
Traversal
↓
Impact Analysis

using migration-backed SQLite repositories.

Verify:

* Impact chains
* Evidence preservation
* Stable ordering
* Reason generation

---

# Constraints

Do NOT implement:

* MCP tools
* Conflict detection
* AI reasoning

Only implement impact analysis.

---

# Acceptance Criteria

Impact analysis reuses traversal.

Impact reports preserve graph evidence.

Impact reports contain deterministic reasons.

Impact ordering is reproducible.

All tests pass.
