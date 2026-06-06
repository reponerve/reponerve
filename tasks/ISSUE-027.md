# ISSUE-027 — MCP Architecture & Foundations

## Objective

Create the foundational MCP service layer for RepoNerve.

This issue establishes the boundary between:

Repository Intelligence
↓
MCP Transport

No MCP server implementation is included in this issue.

---

# Background

RepoNerve currently provides:

* Memory Engine
* Query Engine
* Context Engine

These capabilities must be exposed through MCP without duplicating business logic.

The MCP layer should act as a thin integration layer on top of existing engines.

---

# Goals

Create:

* MCP Service Layer
* MCP Tool Registry
* MCP Tool Definitions

These components will later be consumed by the MCP Server.

---

# Package Structure

Create:

```text
internal/mcp/

models.go

service.go

registry.go
```

---

# MCP Models

Create:

```go
type ToolDefinition struct {
    Name        string
    Description string
}
```

Purpose:

Represent MCP-exposed capabilities.

---

# Tool Registry

Create:

```go
type Registry struct {
}
```

Responsibilities:

* Register tools
* Prevent duplicate registrations
* List tools
* Lookup tools

---

# MCP Service

Create:

```go
type Service struct {
}
```

The service aggregates existing RepoNerve capabilities.

Dependencies may include:

* DecisionReader
* IntentReader
* FactReader
* EventReader
* RelationshipReader
* Context Generator
* Context Renderer

The service itself should not expose MCP transport logic.

---

# Initial Tool Definitions

Register:

```text
list_decisions

get_decision

trace_decision

generate_context
```

Only definitions.

No execution logic.

---

# Constraints

Do NOT implement:

* MCP server
* JSON-RPC
* STDIO transport
* HTTP transport
* Tool execution
* Agent integrations
* Authentication
* Ownership extraction

Only architecture foundations.

---

# Unit Tests

Cover:

* Tool registration
* Duplicate registration prevention
* Tool lookup
* Tool listing
* Service construction

---

# Integration Tests

Verify:

MCP Registry
↓
MCP Service
↓
RepoNerve Dependencies

without transport layers.

---

# Acceptance Criteria

The MCP foundation layer exists and compiles.

Tool definitions can be registered and discovered.

No transport or protocol implementation exists.

All tests pass.
