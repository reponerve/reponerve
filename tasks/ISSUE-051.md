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

## Intelligence Service Authority Rule

Repository Intelligence MCP Tools must delegate to Repository Intelligence services.

MCP must not implement:

- Discovery logic
- Learning Path logic
- Reviewer Recommendation logic
- Change Planning logic

Responsibilities:

Repository Intelligence Services
↓
Generate Intelligence

MCP
↓
Expose Intelligence

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
```

---

## Additional Architecture Rules

## Intelligence Service Authority Rule

Repository Intelligence MCP Tools must delegate to Repository Intelligence services.

MCP must not implement:

* Knowledge Discovery logic
* Learning Path generation logic
* Reviewer Recommendation logic
* Change Planning logic

Responsibilities:

Repository Intelligence Services
↓
Generate Intelligence

MCP
↓
Expose Intelligence

MCP remains a thin orchestration layer.

---

## Service Reuse Requirements

The MCP layer must reuse:

* discovery.Service
* learning.Service
* reviewers.Service
* changeplan.Service

MCP must not duplicate intelligence logic.

All intelligence generation must occur inside the corresponding service.

---

## Response Preservation Rule

Repository Intelligence MCP responses must preserve:

* EvidenceJSON
* Explanation
* Score
* Priority
* Position

Responses must not flatten or discard intelligence metadata.

Repository Intelligence outputs should be returned directly from the service layer.

---

## Routing Requirements

Tool:

discover_knowledge

Delegates to:

discovery.Service

---

Tool:

generate_learning_path

Delegates to:

learning.Service

---

Tool:

recommend_reviewers

Delegates to:

reviewers.Service

---

Tool:

generate_change_plan

Delegates to:

changeplan.Service

---

## Validation Requirements

Validate:

* repository_id
* contributor_id
* domain
* entity_type
* entity_id
* path_type
* recommendation_type

Return structured MCP errors for invalid requests.

Do not panic.

---

## Testing Requirements

Add verification for:

* Tool registration
* Tool discovery
* Schema validation
* Parameter validation
* Routing correctness
* Evidence preservation
* Explanation preservation
* Deterministic ordering

All Repository Intelligence tools must have execution coverage.

---

## Acceptance Criteria Addendum

Repository Intelligence MCP Tools are considered complete when:

* All tools are registered
* All tools execute successfully
* Service delegation is verified
* Evidence is preserved
* Explanations are preserved
* Ordering remains deterministic
* No intelligence logic exists inside MCP
* All tests pass
