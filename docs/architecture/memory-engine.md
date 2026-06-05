# Memory Engine Architecture

## Purpose

The Memory Engine converts repository sources into repository memories.

Sources represent raw repository artifacts.

Memories represent structured repository knowledge.

---

# Architecture

```text
Repository
    ↓
Scanners
    ↓
Sources
    ↓
Memory Engine
    ↓
Memories
    ↓
Relationships
```

---

# Components

## Event Extractor

Input:

* Git Commits

Output:

* Events

Example:

Commit:

feat(cache): introduce redis

Produces:

Event:

Redis Introduced

---

## Decision Extractor

Input:

* ADRs

Output:

* Decisions

Example:

ADR:

Use Redis

Produces:

Decision:

Use Redis

---

## Intent Extractor

Input:

* ADRs
* Commit Messages

Output:

* Intents

Example:

Reduce Database Latency

Produces:

Intent:

Reduce Latency

---

## Fact Extractor

Input:

* ADRs
* Documentation

Output:

* Facts

Example:

Auth Service Uses Redis

Produces:

Fact:

Subject: Auth Service

Predicate: Uses

Object: Redis

---

## Ownership Extractor

Input:

* ADRs
* Documentation
* CODEOWNERS (future)

Output:

* Ownership Records

---

## Memory Linker

Input:

* Events
* Decisions
* Intents
* Facts
* Ownership

Output:

* Relationships

Example:

Intent
→ Decision

Decision
→ Event

---

# Storage

Memories will be persisted separately from sources.

Sources remain immutable.

Memories can be regenerated from sources.

---

# Deterministic First

V1 extraction will be entirely deterministic.

No AI.

No embeddings.

No vector databases.

---

# Future

Future versions may support:

* LLM-assisted extraction
* Semantic memory search
* Knowledge graph visualization
* Context generation
* MCP memory tools
