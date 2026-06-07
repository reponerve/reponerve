# ISSUE-045 — Impact Graph Analysis

## Objective

Implement graph-aware impact analysis.

---

# Background

Knowledge Graph Intelligence should answer:

What breaks if this changes?

This issue introduces impact propagation through graph relationships.

---

# Scope

## Impact Engine

Create:

internal/graph/impact/

Implement:

* AnalyzeDecisionImpact
* AnalyzeFactImpact
* AnalyzeEventImpact
* AnalyzeContributorImpact

---

## Impact Propagation

Traverse graph dependencies and dependents.

Return:

* Affected decisions
* Affected facts
* Affected events
* Affected contributors

---

## Evidence Requirements

Every impact result must include supporting graph evidence.

---

## Deterministic Behavior

Impact results must be reproducible.

---

## Testing

Cover:

* Impact propagation
* Dependency chains
* Empty graphs
* Cyclic graphs
* Evidence generation

Include integration tests.

---

# Constraints

Do NOT implement:

* MCP tools

Only implement graph impact analysis.

---

# Acceptance Criteria

Impact analysis is graph-aware.

Impact evidence is preserved.

All tests pass.
