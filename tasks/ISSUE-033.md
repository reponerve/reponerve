# ISSUE-033 — Repository Q&A

## Objective

Enable deterministic repository question answering.

---

# Background

RepoNerve already stores repository knowledge.

Users and agents need natural access to that knowledge.

---

# Supported Questions

Examples:

Why was this decision made?

What facts support this decision?

What event caused this?

What depends on this?

---

# Architecture

Question
↓
Agent Service
↓
Query Engine
↓
Answer

---

# Deliverables

Create:

internal/agent/qa/

service.go

---

# Constraints

No external LLM.

No vector search.

No embeddings.

Use existing memory graph only.

---

# Acceptance Criteria

Supported repository questions can be answered deterministically.
