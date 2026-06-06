# ISSUE-034 — Impact Analysis

## Objective

Determine the impact of repository changes.

---

# Background

Agents frequently need to understand change consequences.

---

# Example Questions

If decision X changes:

* Which facts are affected?
* Which events are related?
* Which intents are involved?

---

# Architecture

Entity
↓
Relationship Graph
↓
Impact Analyzer
↓
Impact Report

---

# Deliverables

Create:

internal/agent/impact/

service.go

models.go

---

# Constraints

Use relationship graph only.

No AI.

No embeddings.

---

# Acceptance Criteria

Impact reports can be generated for repository entities.
