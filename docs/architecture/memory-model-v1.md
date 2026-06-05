# Memory Model V1

## Purpose

The Memory Model defines how RepoNerve represents repository knowledge.

Scanners collect sources.

The Memory Engine transforms sources into memories.

Memories are durable knowledge units that can be queried, linked, explained, and consumed by humans and AI systems.

---

## Principles

### Source First

Every memory must originate from one or more repository sources.

### Traceability

Every memory must be traceable back to its originating source.

### Deterministic First

Memory extraction should be deterministic whenever possible.

### AI Optional

AI may enhance memory extraction in the future but is not required for V1.

### Linkable

Memories should be designed to support relationships and graph traversal.

---

# Memory Types

## Event

Represents something that happened.

Examples:

* Redis Introduced
* Authentication Service Created
* ADR Added
* API Deprecated

Schema:

```go
type Event struct {
    ID string

    RepositoryID string

    EventType string

    Title string

    Description string

    SourceID string

    Timestamp time.Time
}
```

---

## Decision

Represents a repository decision.

Examples:

* Use Redis
* Adopt gRPC
* Move To PostgreSQL

Schema:

```go
type Decision struct {
    ID string

    RepositoryID string

    Title string

    Status string

    SourceID string
}
```

---

## Intent

Represents the reason behind a decision.

Examples:

* Reduce Latency
* Improve Reliability
* Lower Cost
* Simplify Deployment

Schema:

```go
type Intent struct {
    ID string

    RepositoryID string

    Description string

    SourceID string
}
```

---

## Fact

Represents a repository truth.

Examples:

* Auth Service Uses Redis
* Billing Uses PostgreSQL
* API Calls User Service

Schema:

```go
type Fact struct {
    ID string

    RepositoryID string

    Subject string

    Predicate string

    Object string

    SourceID string
}
```

---

## Ownership

Represents ownership information.

Examples:

* Backend Team Owns Auth Service
* Alice Owns Billing Service

Schema:

```go
type Ownership struct {
    ID string

    RepositoryID string

    Owner string

    Resource string

    SourceID string
}
```

---

## Relationship

Represents links between memories.

Examples:

Intent → Decision

Decision → Event

Fact → Ownership

Schema:

```go
type Relationship struct {
    ID string

    FromID string

    ToID string

    Type string
}
```

---

# Source Traceability

Every memory must maintain references to its source artifacts.

Example:

ADR
→ Decision

Commit
→ Event

ADR
→ Intent

This allows explanations to always provide evidence.

---

# Version

Version: 1.0

Status: Draft
