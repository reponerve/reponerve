# ISSUE-054 — Agent Session Intelligence

Status: Implemented

Milestone: v1.0

---

# Objective

Implement Agent Session Intelligence.

Agent Session Intelligence creates repository-aware sessions for AI agents.

It answers:

- What repository context has already been provided?
- What repository knowledge is currently relevant?
- What context should be carried forward?
- What intelligence packages should remain available during an agent interaction?

---

# Background

RepoNerve already provides:

- Repository Intelligence
- Agent Context Packages
- Repository Search

However, agents still operate request-by-request.

Agent Session Intelligence introduces repository-aware session packaging.

---

# Philosophy

Evidence First.

Agent Session Intelligence packages repository intelligence.

Agent Session Intelligence does not generate repository intelligence.

Agent Session Intelligence must remain deterministic.

---

# Session Authority Rule

Agent Session Intelligence consumes:

- Agent Context Packages
- Repository Search Results
- Repository Intelligence

Agent Session Intelligence must not:

- Generate discovery results
- Generate learning paths
- Generate reviewer recommendations
- Generate change plans
- Generate graph relationships
- Generate repository intelligence

Responsibilities:

Repository Intelligence
↓
Produces Intelligence

Agent Session Intelligence
↓
Maintains Session Context

Session Intelligence is a session layer.

It is not an intelligence layer.

---

# Architecture Requirements

Reuse:

- Agent Context Builder
- Repository Search
- Repository Intelligence Services

Do NOT:

- Access SQLite directly
- Execute Git commands
- Re-scan repositories
- Recompute repository intelligence

---

# Scope

Create:

internal/agent/session/

Files:

- models.go
- service.go
- service_test.go

---

# Models

Implement:

```go
type SessionArtifact struct {
    ArtifactType string `json:"artifact_type"`

    Source string `json:"source"`

    Data json.RawMessage `json:"data"`
}
```

Supported ArtifactType values:

- context_package
- search_result

---

Implement:

```go
type AgentSession struct {
    SessionID string `json:"session_id"`

    RepositoryID string `json:"repository_id"`

    Artifacts []*SessionArtifact `json:"artifacts"`
}
```

---

# Validation

Implement:

```go
func ValidateSession(
    session *AgentSession,
) error
```

Validate:

- SessionID exists
- RepositoryID exists
- Artifacts valid
- ArtifactType supported
- Source exists
- Data exists

Reject invalid sessions.

---

# Session Types

Support:

## Repository Session

Repository-wide context.

---

## Domain Session

Domain-focused context.

---

## Contributor Session

Contributor-focused context.

---

# Artifact Sources

Supported Source values:

- context
- search
- discovery
- learning
- reviewers
- changeplan

---

# Session Composition

Repository Session:

1. Repository Context Package

---

Domain Session:

1. Domain Context Package
2. Domain Search Results

---

Contributor Session:

1. Contributor Context Package
2. Contributor Search Results

---

# Session Rules

Agent Session Intelligence must:

- Preserve upstream artifacts unchanged
- Preserve evidence
- Preserve explanations
- Preserve scores
- Preserve priorities

Agent Session Intelligence must not:

- Modify intelligence outputs
- Re-rank search results
- Re-rank repository intelligence

---

# Ordering

Artifacts must remain in insertion order.

Do not sort artifacts.

Ordering must remain deterministic.

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

Inject required dependencies.

---

# APIs

Implement:

```go
func (s *Service) CreateRepositorySession(
    ctx context.Context,
    repositoryID string,
) (*AgentSession, error)
```

---

Implement:

```go
func (s *Service) CreateDomainSession(
    ctx context.Context,
    repositoryID string,
    domain string,
) (*AgentSession, error)
```

---

Implement:

```go
func (s *Service) CreateContributorSession(
    ctx context.Context,
    repositoryID string,
    contributorID string,
) (*AgentSession, error)
```

---

# Session Identity

Session IDs must be deterministic.

Recommended:

sha256(
    repositoryID +
    sessionType +
    identifier
)

Examples:

repository session:
repositoryID + "repository"

domain session:
repositoryID + domain

contributor session:
repositoryID + contributorID

---

# Unit Tests

Cover:

- Validation
- Repository sessions
- Domain sessions
- Contributor sessions
- Artifact preservation
- Ordering
- Deterministic IDs

---

# Integration Tests

Create migration-backed SQLite tests.

Verify:

Repository Intelligence
↓
Agent Context Builder
↓
Repository Search
↓
Agent Session Intelligence

Verify:

- Session creation
- Artifact preservation
- Ordering
- Determinism

---

# Constraints

Do NOT:

- Persist sessions
- Add MCP tools
- Add AI reasoning
- Add autonomous memory
- Modify upstream intelligence

Only implement deterministic session packaging.

---

# Acceptance Criteria

Agent sessions are created successfully.

Repository intelligence is reused.

Search results are reused.

No intelligence is recomputed.

Evidence is preserved.

Ordering is deterministic.

Session IDs are deterministic.

All tests pass.
