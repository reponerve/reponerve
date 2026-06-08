# ISSUE-046 — Knowledge Graph MCP Tools

## Objective

Expose Knowledge Graph Intelligence through the MCP Server.

---

# Background

Knowledge Graph Intelligence provides:

* Graph Relationships
* Graph Traversal
* Impact Analysis

These capabilities must be available to AI agents through MCP.

The MCP layer must remain a thin orchestration layer.

Business logic belongs in graph services.

---

# Philosophy

Evidence First.

Graph conclusions are derived knowledge.

Graph conclusions are not facts.

Every MCP graph response must preserve:

* Paths
* Relationships
* Evidence
* Reasoning

---

# Scope

Expose Knowledge Graph capabilities through MCP.

Reuse:

* Graph Relationship Engine
* Graph Traversal Engine
* Impact Analysis Service

Do not duplicate graph logic.

---

# Architecture Requirements

Dependency direction:

Storage
↓
Readers
↓
Graph Services
↓
MCP

MCP tools must not:

* Access SQLite directly
* Generate graph relationships
* Perform traversal internally
* Perform impact analysis internally

MCP delegates to graph services.

---

# MCP Tools

Register:

* trace_graph
* trace_path
* find_dependencies
* find_dependents
* analyze_impact

---

# trace_graph

Arguments:

```json id="7z5k5u"
{
  "node_id": "string",
  "max_depth": 10
}
```

Returns:

Graph traversal results originating from the specified node.

---

# trace_path

Arguments:

```json id="yvlc30"
{
  "start_node_id": "string",
  "end_node_id": "string"
}
```

Returns:

The graph paths connecting the start and end nodes.

Paths must preserve:

* Nodes
* Edges
* Evidence

---

# find_dependencies

Arguments:

```json id="x65jv4"
{
  "node_id": "string"
}
```

Returns:

Outbound dependency paths.

---

# find_dependents

Arguments:

```json id="3xvaf4"
{
  "node_id": "string"
}
```

Returns:

Inbound dependency paths.

---

# analyze_impact

Arguments:

```json id="wuzxmn"
{
  "node_id": "string",
  "node_type": "string"
}
```

Supported types:

* DECISION
* FACT
* EVENT
* CONTRIBUTOR

Returns:

ImpactReport

including:

* ImpactPaths
* Reasons
* Evidence

---

# Response Requirements

Graph responses must preserve:

* Node ordering
* Edge ordering
* Edge evidence
* Impact reasons

MCP must not discard graph context.

---

# Path Serialization

TraversalPath must be serialized completely.

Example:

```json id="qzvcgo"
{
  "nodes": [...],
  "edges": [...]
}
```

Do not flatten paths into node lists.

---

# Evidence Preservation

Every graph response must include:

* Graph edge evidence
* Traversal paths
* Impact reasons

Evidence-free graph responses are invalid.

---

# Tool Discovery

Expose schemas through:

tools/list

Update MCP registry counts and discovery tests.

---

# Tool Execution

Implement routing through:

tools/call

Reuse existing MCP architecture.

---

# Error Handling

Handle:

* Missing node IDs
* Unknown nodes
* Unsupported node types
* Empty graphs
* Missing paths

Return structured MCP tool errors.

Do not panic.

---

# Testing

Update:

* mcp_test.go
* server_test.go

Create graph MCP integration tests.

Verify:

* Tool registration
* Tool discovery
* Tool execution
* Path serialization
* Evidence preservation
* Impact analysis
* Error handling

---

# Constraints

Do NOT:

* Add AI reasoning
* Add embeddings
* Add vector search
* Add graph mutation

Only expose graph capabilities through MCP.

---

# Acceptance Criteria

Knowledge Graph Intelligence is available through MCP.

Paths remain intact.

Evidence remains visible.

Impact analysis remains explainable.

All tests pass.
