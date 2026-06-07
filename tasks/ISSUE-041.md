# ISSUE-041 — Ownership MCP Tools

## Objective

Expose ownership intelligence through MCP.

---

# Background

Ownership data should be available to AI coding agents through the existing MCP server.

---

# Scope

## MCP Tools

Implement:

* list_contributors
* get_contributor
* list_expertise
* trace_contributor
* recommend_reviewers

---

## Reviewer Recommendations

Recommendations must be evidence-based.

Example:

Contributor: Alice

Score: 0.92

Evidence:

* 87 commits
* 4 decisions
* Recent activity

---

## Tool Discovery

Expose tools through:

* tools/list

with JSON schemas.

---

## Tool Execution

Expose tools through:

* tools/call

using existing MCP architecture.

---

## Testing

Cover:

* Discovery
* Execution
* Error handling
* Evidence generation

Include integration tests.

---

# Explainability Requirement

All reviewer recommendations must include evidence.

Unsupported:

* "Best engineer"
* Subjective rankings

Supported:

* Activity-based recommendations
* Expertise-based recommendations

---

# Acceptance Criteria

Ownership intelligence is available through MCP.

All recommendations are explainable.

All tests pass.
