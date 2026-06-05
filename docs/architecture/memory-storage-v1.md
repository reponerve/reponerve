# Memory Storage V1

## Purpose

This document defines how repository memories are persisted in RepoNerve.

The Memory Engine transforms repository sources into structured memories.

Memory storage is responsible for persisting those memories while preserving traceability back to the original repository artifacts.

---

# Design Principles

## Source of Truth

Repository sources remain the canonical source of truth.

Examples:

* Git commits
* ADRs
* Documentation
* Pull Requests (future)
* Issues (future)

Memories are derived from sources and can be regenerated at any time.

---

## Deterministic Storage

Memory storage should not depend on:

* LLMs
* Embeddings
* Vector databases
* Graph databases

V1 uses SQLite only.

---

## Traceability

Every memory record must maintain a reference to the source that produced it.

Example:

ADR
→ Decision

Commit
→ Event

ADR
→ Intent

---

## Regeneratable

All memories should be disposable and reproducible.

The system must be able to:

1. Delete all memories
2. Re-run extraction
3. Produce equivalent memory records

---

# Storage Architecture

```text
repositories
     ↓
sources
     ↓
events
decisions
intents
facts
ownerships
     ↓
relationships
```

---

# Tables

## events

Represents repository events.

Examples:

* Redis Introduced
* Authentication Service Created
* API Deprecated

Schema:

```sql
CREATE TABLE events (
    id TEXT PRIMARY KEY,

    repository_id TEXT NOT NULL,

    event_type TEXT NOT NULL,

    title TEXT NOT NULL,

    description TEXT,

    source_id TEXT NOT NULL,

    timestamp DATETIME NOT NULL,

    created_at DATETIME NOT NULL
);
```

---

## decisions

Represents repository decisions.

Examples:

* Use Redis
* Adopt gRPC
* Migrate To PostgreSQL

Schema:

```sql
CREATE TABLE decisions (
    id TEXT PRIMARY KEY,

    repository_id TEXT NOT NULL,

    title TEXT NOT NULL,

    status TEXT NOT NULL,

    source_id TEXT NOT NULL,

    created_at DATETIME NOT NULL
);
```

---

## intents

Represents repository intent.

Examples:

* Reduce Latency
* Improve Reliability
* Lower Cost

Schema:

```sql
CREATE TABLE intents (
    id TEXT PRIMARY KEY,

    repository_id TEXT NOT NULL,

    description TEXT NOT NULL,

    source_id TEXT NOT NULL,

    created_at DATETIME NOT NULL
);
```

---

## facts

Represents repository facts.

Examples:

* Auth Service Uses Redis
* Billing Uses PostgreSQL

Schema:

```sql
CREATE TABLE facts (
    id TEXT PRIMARY KEY,

    repository_id TEXT NOT NULL,

    subject TEXT NOT NULL,

    predicate TEXT NOT NULL,

    object TEXT NOT NULL,

    source_id TEXT NOT NULL,

    created_at DATETIME NOT NULL
);
```

---

## ownerships

Represents ownership information.

Examples:

* Backend Team Owns Auth Service
* Platform Team Owns Deployment System

Schema:

```sql
CREATE TABLE ownerships (
    id TEXT PRIMARY KEY,

    repository_id TEXT NOT NULL,

    owner TEXT NOT NULL,

    resource TEXT NOT NULL,

    source_id TEXT NOT NULL,

    created_at DATETIME NOT NULL
);
```

---

## relationships

Represents links between memories.

Examples:

Intent
→ Decision

Decision
→ Event

Fact
→ Ownership

Schema:

```sql
CREATE TABLE relationships (
    id TEXT PRIMARY KEY,

    from_id TEXT NOT NULL,

    to_id TEXT NOT NULL,

    relationship_type TEXT NOT NULL,

    created_at DATETIME NOT NULL
);
```

---

# Relationship Types

Supported V1 relationship types:

```text
INTENT_DRIVES_DECISION

DECISION_RESULTS_IN_EVENT

FACT_OWNED_BY

EVENT_SUPPORTS_DECISION

SOURCE_GENERATED_MEMORY
```

Additional relationship types may be introduced in future releases.

---

# Indexing Strategy

Create indexes on:

```sql
CREATE INDEX idx_events_repository_id
ON events(repository_id);

CREATE INDEX idx_decisions_repository_id
ON decisions(repository_id);

CREATE INDEX idx_intents_repository_id
ON intents(repository_id);

CREATE INDEX idx_facts_repository_id
ON facts(repository_id);

CREATE INDEX idx_ownerships_repository_id
ON ownerships(repository_id);

CREATE INDEX idx_relationships_from_id
ON relationships(from_id);

CREATE INDEX idx_relationships_to_id
ON relationships(to_id);
```

---

# Deletion Strategy

Sources are immutable.

Memories are replaceable.

During future rescans:

1. Sources are updated.
2. Memories may be regenerated.
3. Relationships may be rebuilt.

Memory records should never become the system of record.

---

# Future Evolution

Potential future enhancements:

* Memory versioning
* Memory confidence scores
* Semantic memory search
* Embedding storage
* Knowledge graph visualization
* Distributed memory storage
* MCP memory APIs

These features are explicitly out of scope for V1.

---

# Version

Version: 1.0

Status: Draft
