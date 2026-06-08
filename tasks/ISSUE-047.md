# ISSUE-047 — Knowledge Discovery Engine

Status: Implemented

Milestone: v0.9.0-alpha

---

# Objective

Implement the Knowledge Discovery Engine.

The Knowledge Discovery Engine helps humans and AI systems discover important repository knowledge.

It answers:

- What should I read?
- What repository knowledge matters most?
- Where does repository knowledge live?

---

# Background

RepoNerve currently stores:

- Decisions
- Facts
- Events
- Intents
- Contributors
- Expertise
- Knowledge Graph relationships

Repository Intelligence requires a mechanism for surfacing important repository knowledge.

Knowledge Discovery provides that capability.

---

# Philosophy

Evidence First.

Discovery results are recommendations.

Discovery results are not facts.

Every discovery result must include:

- Evidence
- Explanation

Knowledge Discovery must remain deterministic.

---

# Scope

Create:

internal/intelligence/discovery/

Files:

- models.go
- service.go
- service_test.go

---

# Architecture Requirements

Reuse:

- Memory Engine
- Ownership Intelligence
- Knowledge Graph Intelligence
- Context Engine

Do NOT:

- Access SQLite directly
- Execute Git commands
- Re-scan repositories

Knowledge Discovery consumes repository knowledge.

It does not create repository knowledge.

---

# Discovery Models

Implement:

```go
type DiscoveryItem struct {
    EntityType string

    EntityID string

    Score float64

    EvidenceJSON string

    Explanation string
}
