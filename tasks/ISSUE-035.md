# ISSUE-035 — Architectural Guidance

## Objective

Provide architectural explanations using repository memory.

---

# Background

RepoNerve already stores:

* Decisions
* Facts
* Intents
* Relationships

Agents need explanations.

---

# Examples

Why was this decision made?

What supports this decision?

Which intent drove this?

---

# Architecture

Entity
↓
Memory Graph
↓
Guidance Service
↓
Explanation

---

# Deliverables

Create:

internal/agent/guidance/

service.go

---

# Constraints

Deterministic.

No LLM dependency.

Reuse Explain Engine.

---

# Acceptance Criteria

Architectural explanations can be generated from repository memory.
