# ISSUE-051 — Repository Intelligence MCP Tools

Status: Planned

Milestone: v0.9.0-alpha

---

# Objective

Expose Repository Intelligence capabilities through the MCP Server.

Repository Intelligence MCP Tools allow AI systems and external clients to consume:

- Knowledge Discovery
- Learning Paths
- Reviewer Recommendations
- Change Planning

through a consistent MCP interface.

---

# Background

Repository Intelligence introduces:

- Knowledge Discovery Engine
- Repository Learning Paths
- Reviewer Recommendation Engine
- Change Planning Engine

These capabilities must be accessible through MCP.

The MCP layer remains a thin orchestration layer.

Business logic belongs in Repository Intelligence services.

---

# Philosophy

Evidence First.

Repository Intelligence recommendations are derived knowledge.

Repository Intelligence recommendations are not facts.

Every MCP response must preserve:

- Evidence
- Explanations
- Ordering

MCP must never discard repository intelligence context.

---

# Scope

Expose Repository Intelligence through MCP.

Reuse:

- Knowledge Discovery Engine
- Learning Path Engine
- Reviewer Recommendation Engine
- Change Planning Engine

Do not duplicate intelligence logic.

---

# Architecture Requirements

Dependency Direction:

Storage
↓
Readers
↓
Memory
↓
Ownership
↓
Knowledge Graph
↓
Repository Intelligence
↓
MCP

MCP must not:

- Access SQLite directly
- Execute Git commands
- Re-scan repositories
- Re-implement intelligence logic

MCP delegates to services.

---

# MCP Tools

Register:

- discover_knowledge
- generate_learning_path
- recommend_reviewers
- generate_change_plan

---

# discover_knowledge

Arguments:

```json
{
  "repository_id": "string"
}