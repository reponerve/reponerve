# ISSUE-030 — Context MCP Tools

## Objective

Expose Context Engine capabilities through MCP.

---

# Background

RepoNerve can already generate:

RepositoryContext

through:

reponerve context generate

This issue exposes that capability through MCP.

---

# MCP Tools

Implement:

generate_context

export_context

---

# Architecture

MCP Tool
↓
MCP Service
↓
Context Engine

---

# Output

RepositoryContext

or

Rendered Markdown Context

depending on tool.

---

# Constraints

Do NOT:

* Reimplement context generation
* Reimplement rendering

Reuse existing Context Engine components.

---

# Unit Tests

Cover:

* Context generation
* Empty repositories
* Rendering

---

# Acceptance Criteria

Repository context is available through MCP.
