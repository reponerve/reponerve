# RepoNerve Data Model

Version: 1.0

Status: Draft

Authors: RepoNerve Contributors

Last Updated: 2026-06-05

---

# Purpose

This document defines the physical data model used by RepoNerve.

The data model is responsible for storing:

* Repository metadata
* Repository memory
* Sources
* Relationships
* Evidence
* Search indexes

The MVP implementation uses SQLite as the primary datastore.

---

# Design Goals

The data model must be:

* Local-first
* Portable
* Searchable
* Explainable
* Migration-friendly
* Graph-compatible
* AI-friendly

---

# Database Technology

## Primary Database

SQLite

Reasons:

* Embedded
* No server required
* Cross-platform
* Reliable
* Excellent tooling

---

## Search Engine

SQLite FTS5

Reasons:

* Native integration
* Lightweight
* Fast text search
* No external dependencies

---

# Database Overview

```text
┌───────────────────┐
│   repositories    │
└─────────┬─────────┘
          │
          ▼
┌───────────────────┐
│      sources      │
└─────────┬─────────┘
          │
          ▼
┌───────────────────┐
│     memories      │
└─────────┬─────────┘
          │
          ▼
┌───────────────────┐
│   relationships   │
└─────────┬─────────┘
          │
          ▼
┌───────────────────┐
│     evidence      │
└───────────────────┘
```

---

# Repository Table

Stores repository metadata.

---

## repositories

```sql
CREATE TABLE repositories (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    path TEXT NOT NULL,
    default_branch TEXT,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL
);
```

---

## Example

```text
id: repo_001
name: reponerve
path: /workspace/reponerve
default_branch: main
```

---

# Sources Table

Stores evidence sources.

Every memory object must originate from one or more sources.

---

## sources

```sql
CREATE TABLE sources (
    id TEXT PRIMARY KEY,
    repository_id TEXT NOT NULL,
    source_type TEXT NOT NULL,
    reference TEXT NOT NULL,
    title TEXT,
    author TEXT,
    timestamp DATETIME,
    metadata_json TEXT,
    created_at DATETIME NOT NULL,

    FOREIGN KEY (repository_id)
        REFERENCES repositories(id)
);
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

# Memory Table

Stores all memory entities.

Uses a polymorphic design.

---

## memories

```sql
CREATE TABLE memories (
    id TEXT PRIMARY KEY,
    repository_id TEXT NOT NULL,
    memory_type TEXT NOT NULL,
    title TEXT NOT NULL,
    summary TEXT,
    confidence TEXT NOT NULL,
    metadata_json TEXT,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,

    FOREIGN KEY (repository_id)
        REFERENCES repositories(id)
);
```

---

## Memory Types

```text
fact
event
decision
ownership
intent
relationship
pattern
```

Pattern memory is reserved for future releases.

---

# Fact Table

Stores fact-specific data.

---

## facts

```sql
CREATE TABLE facts (
    memory_id TEXT PRIMARY KEY,
    subject TEXT NOT NULL,
    predicate TEXT NOT NULL,
    object TEXT NOT NULL,

    FOREIGN KEY (memory_id)
        REFERENCES memories(id)
);
```

---

## Example

```text
subject: UserService
predicate: USES
object: Redis
```

---

# Event Table

Stores event-specific information.

---

## events

```sql
CREATE TABLE events (
    memory_id TEXT PRIMARY KEY,
    event_type TEXT NOT NULL,
    event_timestamp DATETIME,
    actor TEXT,
    resource TEXT,

    FOREIGN KEY (memory_id)
        REFERENCES memories(id)
);
```

---

## Example

```text
event_type: PullRequestMerged
actor: john
resource: PR-143
```

---

# Decision Table

Stores decision memory.

This is one of RepoNerve's most valuable datasets.

---

## decisions

```sql
CREATE TABLE decisions (
    memory_id TEXT PRIMARY KEY,
    reason TEXT,
    alternatives TEXT,
    tradeoffs TEXT,
    decision_maker TEXT,
    outcome TEXT,

    FOREIGN KEY (memory_id)
        REFERENCES memories(id)
);
```

---

## Example

```text
Decision:
Use Redis

Reason:
Reduce latency

Alternatives:
Memcached

Tradeoffs:
Additional infrastructure complexity
```

---

# Ownership Table

Stores ownership information.

---

## ownerships

```sql
CREATE TABLE ownerships (
    memory_id TEXT PRIMARY KEY,
    resource TEXT NOT NULL,
    owner_type TEXT NOT NULL,
    owner TEXT NOT NULL,
    start_date DATETIME,
    end_date DATETIME,

    FOREIGN KEY (memory_id)
        REFERENCES memories(id)
);
```

---

# Intent Table

Stores goals and motivations.

---

## intents

```sql
CREATE TABLE intents (
    memory_id TEXT PRIMARY KEY,
    goal TEXT NOT NULL,
    description TEXT,
    outcome TEXT,

    FOREIGN KEY (memory_id)
        REFERENCES memories(id)
);
```

---

# Relationship Table

Stores graph-style relationships.

---

## relationships

```sql
CREATE TABLE relationships (
    id TEXT PRIMARY KEY,
    source_memory_id TEXT NOT NULL,
    relation TEXT NOT NULL,
    target_memory_id TEXT NOT NULL,
    confidence TEXT,

    FOREIGN KEY (source_memory_id)
        REFERENCES memories(id),

    FOREIGN KEY (target_memory_id)
        REFERENCES memories(id)
);
```

---

## Example

```text
Decision
    INFLUENCED_BY
ADR
```

---

# Evidence Table

Connects memory to sources.

This table powers explainability.

---

## evidence

```sql
CREATE TABLE evidence (
    id TEXT PRIMARY KEY,
    memory_id TEXT NOT NULL,
    source_id TEXT NOT NULL,
    confidence TEXT,
    explanation TEXT,

    FOREIGN KEY (memory_id)
        REFERENCES memories(id),

    FOREIGN KEY (source_id)
        REFERENCES sources(id)
);
```

---

## Example

```text
Decision:
Use Redis

Evidence:
PR-143

Confidence:
Explicit
```

---

# Search Index

Full-text search table.

---

## memory_search

```sql
CREATE VIRTUAL TABLE memory_search
USING fts5 (
    memory_id,
    title,
    summary,
    content
);
```

---

# Query Strategy

All user queries follow:

```text
Question
     │
     ▼
FTS Search
     │
     ▼
Memory Retrieval
     │
     ▼
Relationship Expansion
     │
     ▼
Evidence Collection
     │
     ▼
Response Generation
```

---

# Repository Workspace

SQLite database location:

```text
.reponerve/
└── memory.db
```

---

# Future Tables

Reserved for future releases.

---

## patterns

Stores repository conventions.

---

## incidents

Stores operational incidents.

---

## risks

Stores architectural risks.

---

## technical_debt

Stores known debt records.

---

## context_packs

Stores reusable AI context bundles.

---

# Indexing Strategy

Required indexes:

```sql
CREATE INDEX idx_memories_type
ON memories(memory_type);

CREATE INDEX idx_sources_type
ON sources(source_type);

CREATE INDEX idx_evidence_memory
ON evidence(memory_id);

CREATE INDEX idx_relationships_source
ON relationships(source_memory_id);

CREATE INDEX idx_relationships_target
ON relationships(target_memory_id);
```

---

# Migration Strategy

Database schema changes must be versioned.

Migration directory:

```text
internal/storage/migrations/
```

Example:

```text
000001_initial_schema.sql
000002_add_patterns.sql
000003_add_context_packs.sql
```

---

# Scalability Philosophy

The MVP optimizes for:

```text
Single Repository
Single Developer
Local Machine
```

Future scalability concerns should not complicate the MVP architecture.

---

# Success Criteria

The data model succeeds when:

* Repository memory is durable.
* Memory remains queryable.
* Evidence remains traceable.
* Relationships remain navigable.
* Future memory types can be added safely.

---

# Guiding Principle

The database is not a code index.

The database is a repository memory store.
