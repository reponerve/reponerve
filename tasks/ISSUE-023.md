# ISSUE-023 — Context Generator

## Objective

Transform repository memories into a structured Repository Context model.

The Context Generator assembles information from the Query Engine into a single repository briefing.

This context will later be consumed by:

* CLI users
* AI coding agents
* MCP integrations

---

# Background

The Memory Engine stores:

* Events
* Decisions
* Intents
* Facts
* Relationships

The Query Engine retrieves and navigates those memories.

The Context Generator combines them into a coherent repository overview.

---

# Input

The generator must consume memory data through the Query Engine readers.

Allowed:

* EventReader
* DecisionReader
* IntentReader
* FactReader
* RelationshipReader

Not Allowed:

* Direct SQLite access
* Raw SQL queries
* Scanner packages
* Extractor packages

---

# Output Model

Create:

```go
type RepositoryContext struct {
    RepositoryID string

    GeneratedAt time.Time

    Decisions []*models.Decision

    Intents []*models.Intent

    Facts []*models.Fact

    Events []*models.Event
}
```

Keep the model intentionally simple for V1.

---

# Context Generation Rules

## Decisions

Include all repository decisions.

Order:

* Most recent first

---

## Intents

Include all repository intents.

Order:

* Most recent first

---

## Facts

Include all repository facts.

Order:

* Alphabetical by subject

---

## Events

Include recent events.

Order:

* Most recent first

---

# Context Builder

Create:

```go
type Generator struct {
    ...
}
```

Responsibilities:

1. Load repository memories.
2. Build RepositoryContext.
3. Return a fully populated context object.

The generator must be deterministic.

---

# Package Structure

Recommended:

```text
internal/context/

generator.go

models.go
```

Keep context concerns isolated from memory extraction and query logic.

---

# Constraints

Do NOT implement:

* Markdown rendering
* CLI commands
* AI summarization
* LLM integrations
* Embeddings
* MCP
* Ownership extraction

Only generate the RepositoryContext model.

---

# Unit Tests

Cover:

* Empty repository
* Repository with decisions
* Repository with intents
* Repository with facts
* Repository with events
* Ordering guarantees

---

# Integration Tests

Verify:

Readers
↓
Generator
↓
RepositoryContext

using SQLite-backed readers.

---

# Acceptance Criteria

The Context Generator can produce a deterministic RepositoryContext for any scanned repository.

All tests pass.

No direct database access exists within the generator.
