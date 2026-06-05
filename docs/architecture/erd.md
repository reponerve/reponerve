# RepoNerve Entity Relationship Diagram (ERD)

Version: 1.0

Status: Draft

Authors: RepoNerve Contributors

Last Updated: 2026-06-05

---

# Purpose

This document defines the Entity Relationship Diagram (ERD) for RepoNerve.

The ERD represents the physical storage model implemented in SQLite and serves as the reference for:

* Database design
* Repository implementation
* Query development
* Memory retrieval
* Context generation
* MCP integrations

This document complements:

* memory-model.md
* data-model.md

---

# Design Philosophy

RepoNerve is not storing source code.

RepoNerve is storing repository memory.

The database exists to preserve:

* Facts
* Events
* Decisions
* Ownership
* Intent
* Relationships
* Evidence

All entities ultimately connect back to repository artifacts.

---

# High-Level ERD

```text
┌──────────────────┐
│   repositories   │
└─────────┬────────┘
          │
          │ 1:N
          ▼
┌──────────────────┐
│     sources      │
└─────────┬────────┘
          │
          │ N:M
          ▼
┌──────────────────┐
│     evidence     │
└─────────┬────────┘
          │
          │ N:1
          ▼
┌──────────────────┐
│     memories     │
└────┬────┬────┬───┘
     │    │    │
     │    │    │
     ▼    ▼    ▼
 facts events decisions
     │
     ▼
 ownerships
     │
     ▼
 intents

          │
          ▼

┌──────────────────┐
│  relationships   │
└──────────────────┘
```

---

# Entity Overview

| Entity        | Purpose                       |
| ------------- | ----------------------------- |
| repositories  | Repository metadata           |
| sources       | Original repository artifacts |
| memories      | Base memory record            |
| facts         | Objective repository facts    |
| events        | Historical events             |
| decisions     | Architectural decisions       |
| ownerships    | Ownership information         |
| intents       | Goals and motivations         |
| relationships | Connections between memory    |
| evidence      | Traceability layer            |

---

# Repository Entity

Represents a repository indexed by RepoNerve.

---

## Table

```text
repositories
```

---

## Cardinality

```text
Repository
    │
    └───► Sources
```

One repository may contain many sources.

---

## Example

```text
Repository:
reponerve
```

Contains:

```text
Commits

PRs

ADRs

Documentation
```

---

# Source Entity

Represents original repository artifacts.

---

## Table

```text
sources
```

---

## Source Types

```text
commit

pull_request

issue

adr

documentation

source_code
```

---

## Cardinality

```text
Repository
    │
    ▼
Sources
```

One repository contains many sources.

---

## Example

```text
PR-143

ADR-12

Commit abc123
```

---

# Memory Entity

Base entity for all repository memory.

---

## Table

```text
memories
```

---

## Purpose

Provides common fields:

```text
ID

Type

Title

Summary

Confidence

Metadata
```

---

## Cardinality

```text
Memory
    │
    ├── Fact
    ├── Event
    ├── Decision
    ├── Ownership
    └── Intent
```

---

# Fact Entity

Represents objective truths.

---

## Table

```text
facts
```

---

## Example

```text
UserService USES Redis
```

---

## Relationship

```text
Memory
   │
   ▼
Fact
```

1:1

---

# Event Entity

Represents historical events.

---

## Table

```text
events
```

---

## Examples

```text
PR Merged

Commit Created

Issue Closed
```

---

## Relationship

```text
Memory
   │
   ▼
Event
```

1:1

---

# Decision Entity

Represents architectural decisions.

---

## Table

```text
decisions
```

---

## Examples

```text
Use Redis

Adopt Kafka

Split Monolith
```

---

## Relationship

```text
Memory
   │
   ▼
Decision
```

1:1

---

# Ownership Entity

Represents ownership assignments.

---

## Table

```text
ownerships
```

---

## Examples

```text
Billing Service

Owned By

Payments Team
```

---

## Relationship

```text
Memory
   │
   ▼
Ownership
```

1:1

---

# Intent Entity

Represents goals and motivations.

---

## Table

```text
intents
```

---

## Examples

```text
Reduce Latency

Improve Scalability

Meet Compliance
```

---

## Relationship

```text
Memory
   │
   ▼
Intent
```

1:1

---

# Relationship Entity

Represents graph-style connections.

---

## Table

```text
relationships
```

---

## Purpose

Connect memory objects.

---

## Examples

```text
Decision
      │
      ▼
INFLUENCED_BY
      │
      ▼
ADR
```

---

```text
Fact
      │
      ▼
RELATED_TO
      │
      ▼
Decision
```

---

## Cardinality

```text
Memory
    │
    ▼
Relationship
    ▼
Memory
```

Many-to-many

---

# Evidence Entity

Provides traceability.

---

## Table

```text
evidence
```

---

## Purpose

Connect memory to sources.

---

## Cardinality

```text
Memory
     │
     ▼
Evidence
     ▼
Source
```

Many-to-many

---

## Example

```text
Decision:
Use Redis

Evidence:
PR-143

ADR-12
```

---

# Repository Relationship Model

Example:

```text
Repository
      │
      ▼
ADR-12
      │
      ▼
Decision:
Use Redis
      │
      ▼
Intent:
Reduce Latency
      │
      ▼
PR-143
      │
      ▼
Commit abc123
      │
      ▼
Fact:
UserService Uses Redis
```

---

# Memory Graph Example

```text
Intent
      │
      ▼
Reduce Latency
      │
      ▼
Decision
      │
      ▼
Use Redis
      │
      ▼
Event
      │
      ▼
PR-143 Merged
      │
      ▼
Fact
      │
      ▼
UserService Uses Redis
```

---

# Query Traversal Example

Question:

```text
Why was Redis introduced?
```

Traversal:

```text
Redis
  │
  ▼
Fact
  │
  ▼
Decision
  │
  ▼
Intent
  │
  ▼
Evidence
```

Response:

```text
Decision

Intent

Supporting Sources
```

---

# Future Entities

Reserved for future releases.

---

## Pattern

Repository conventions.

---

## Incident

Operational incidents.

---

## Technical Debt

Debt tracking.

---

## Risk

Architectural risk tracking.

---

## Context Pack

Reusable context bundles.

---

# ERD Evolution Strategy

The MVP schema should remain intentionally small.

Future capabilities should extend the model rather than redesign it.

---

# Normalization Strategy

Target:

Third Normal Form (3NF)

Benefits:

* Reduced duplication
* Easier maintenance
* Better consistency

---

# Guiding Principle

Repository memory is the primary entity.

Everything else exists to discover, connect, retrieve, or explain repository memory.
