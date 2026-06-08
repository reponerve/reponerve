# ISSUE-052 — Agent Context Builder

Status: Planned

Milestone: v1.0

---

# Objective

Implement the Agent Context Builder.

The Agent Context Builder packages repository intelligence into deterministic, structured context bundles suitable for AI agents.

It answers:

* What should an agent know about this repository?
* What should an agent know about a domain?
* What should an agent know about a contributor area?

---

# Background

RepoNerve already provides:

* Memory
* Context
* Ownership
* Knowledge Graph
* Repository Intelligence

However, AI agents currently consume these capabilities individually.

Agent Context Builder creates a unified context package that can be consumed directly by agents.

---

# Philosophy

Evidence First.

Agent Context Packages are derived knowledge.

Agent Context Packages are not repository facts.

Every package must preserve:

* Evidence
* Explanations
* Scores
* Priorities

Agent Context Builder must remain deterministic.

---

# Scope

Create:

internal/agent/context/

Files:

* models.go
* service.go
* service_test.go

---

# Architectural Rule

Repository Intelligence remains authoritative.

Agent Context Builder consumes Repository Intelligence.

Agent Context Builder must not:

* Generate repository intelligence
* Re-rank repository knowledge
* Recompute impact
* Generate graph relationships

Responsibilities:

Repository Intelligence
↓
Generates Intelligence

Agent Context Builder
↓
Packages Intelligence

---

# Architecture Requirements

Reuse:

* Discovery Service
* Learning Service
* Reviewer Service
* Change Planning Service
* Context Engine

Do NOT:

* Access SQLite directly
* Execute Git commands
* Re-scan repositories

---

# Models

Implement:

```go
type ContextSection struct {
    Name string `json:"name"`

    Data json.RawMessage `json:"data"`
}
```

---

Implement:

```go
type AgentContextPackage struct {
    RepositoryID string `json:"repository_id"`

    Sections []*ContextSection `json:"sections"`
}
```

---

# Context Package Types

Support:

## Repository Context Package

Answers:

What should an agent know about the repository?

---

## Domain Context Package

Answers:

What should an agent know about a repository domain?

---

## Contributor Context Package

Answers:

What should an agent know about a contributor area?

---

# Package Composition

Repository Context Package should include:

* Repository Overview
* Discovery Results
* Learning Path
* Reviewer Recommendations

---

Domain Context Package should include:

* Domain Discovery
* Domain Learning Path
* Domain Reviewers

---

Contributor Context Package should include:

* Contributor Expertise
* Contributor Learning Path
* Contributor Change Plan

---

# Evidence Preservation

Packages must preserve:

* EvidenceJSON
* Explanations
* Scores
* Priorities
* Positions

No information loss is permitted.

---

# Ordering

Context sections should be ordered:

1. Overview
2. Discovery
3. Learning
4. Reviewers
5. Change Planning

Ordering must be deterministic.

---

# Service

Implement:

```go
type Service struct {
}
```

---

Constructor:

```go
func NewService(...) *Service
```

---

# APIs

Implement:

```go
func (s *Service) BuildRepositoryContext(
    ctx context.Context,
    repositoryID string,
) (*AgentContextPackage, error)
```

---

Implement:

```go
func (s *Service) BuildDomainContext(
    ctx context.Context,
    repositoryID string,
    domain string,
) (*AgentContextPackage, error)
```

---

Implement:

```go
func (s *Service) BuildContributorContext(
    ctx context.Context,
    repositoryID string,
    contributorID string,
) (*AgentContextPackage, error)
```

---

# Validation

Validate:

* RepositoryID exists
* Sections exist
* Section names exist
* Section payloads exist

Reject invalid packages.

---

# Unit Tests

Cover:

* Empty repositories
* Repository packages
* Domain packages
* Contributor packages
* Section ordering
* Evidence preservation
* Deterministic output
* Validation

---

# Integration Tests

Create migration-backed SQLite tests.

Verify:

Repository Intelligence
↓
Agent Context Builder

Verify:

* Package generation
* Evidence preservation
* Ordering
* Determinism

---

# Constraints

Do NOT:

* Generate intelligence
* Re-rank discovery
* Recompute impact
* Add MCP tools

Only implement Agent Context Builder.

---

# Acceptance Criteria

Agent Context Packages are generated successfully.

Repository Intelligence is reused.

Evidence is preserved.

Ordering is deterministic.

All tests pass.
