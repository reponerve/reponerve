# ISSUE-028 — MCP Server Core

## Objective

Implement the RepoNerve MCP Server core.

This issue introduces the MCP transport layer and tool discovery.

The implementation should use STDIO transport.

---

# Background

ISSUE-027 established:

* MCP Service Layer
* MCP Tool Registry
* MCP Tool Definitions

This issue exposes those capabilities through MCP.

---

# Goals

Create:

* MCP Server
* Tool Discovery
* STDIO Transport

No tool execution is implemented in this issue.

---

# Package Structure

Create:

internal/mcp/server/

server.go

---

# Responsibilities

The server must:

* Start via STDIO
* Register tools from the MCP Registry
* Expose tool discovery
* Handle MCP initialization

---

# Constraints

Do NOT implement:

* Tool execution
* Memory queries
* Context generation
* HTTP transport
* Authentication

Only MCP transport and discovery.

---

# Unit Tests

Cover:

* Server initialization
* Tool registration
* Tool discovery

---

# Acceptance Criteria

An MCP-compatible client can connect and discover RepoNerve tools.
