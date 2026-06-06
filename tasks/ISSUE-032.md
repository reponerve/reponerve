# ISSUE-032 — Repository Onboarding

## Objective

Generate a structured repository onboarding package.

---

# Background

RepoNerve already understands:

* Decisions
* Intents
* Facts
* Events
* Relationships
* Repository Context

New contributors and AI agents need a fast way to understand a repository.

---

# Goal

Create an onboarding service that produces:

* Repository Summary
* Key Decisions
* Key Intents
* Key Facts
* Recent Events

---

# Architecture

Repository
↓
Context Engine
↓
Onboarding Service
↓
Onboarding Package

---

# Deliverables

Create:

internal/agent/onboarding/

service.go

models.go

---

# Output

```go
type OnboardingPackage struct {
    RepositoryID string

    Summary string

    Decisions []*models.Decision

    Intents []*models.Intent

    Facts []*models.Fact

    Events []*models.Event
}
```

---

# Constraints

No AI.

No LLMs.

No embeddings.

Deterministic only.

---

# Acceptance Criteria

A repository onboarding package can be generated from repository memory.
