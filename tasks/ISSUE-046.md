# ISSUE-046 — Knowledge Graph MCP Tools

## Objective

Expose Knowledge Graph Intelligence through MCP.

---

# Background

Graph Intelligence should be available to AI coding agents.

---

# Scope

## MCP Tools

Implement:

* trace_graph
* find_dependencies
* find_dependents
* analyze_impact
* find_conflicts

---

## Query Delegation

MCP tools must reuse:

* Graph Traversal Engine
* Graph Impact Engine

No duplicate graph logic.

---

## Evidence Requirements

All graph outputs must include evidence.

Graph conclusions without evidence are invalid.

---

## Tool Discovery

Expose schemas through:

* tools/list

---

## Tool Execution

Expose execution through:

* tools/call

following existing MCP architecture.

---

## Testing

Cover:

* Tool registration
* Tool discovery
* Tool execution
* Evidence preservation
* Error handling

Include integration tests.

---

# Constraints

Do NOT implement:

* AI-generated graph reasoning
* Embeddings
* Vector search

Only expose graph intelligence through MCP.

---

# Acceptance Criteria

Knowledge Graph Intelligence is available through MCP.

Evidence remains visible in all outputs.

All tests pass.
