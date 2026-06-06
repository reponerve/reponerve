# ISSUE-036 — Context Compression

## Objective

Produce compact context packages for AI agents.

---

# Background

Repository context can become large.

Agents operate under token limits.

---

# Goal

Generate smaller context representations.

---

# Architecture

Repository Context
↓
Compression Service
↓
Compressed Context

---

# Deliverables

Create:

internal/agent/compression/

service.go

models.go

---

# Constraints

Deterministic.

No summarization models.

No embeddings.

No vector databases.

---

# Acceptance Criteria

Compressed context packages can be generated from repository context.
