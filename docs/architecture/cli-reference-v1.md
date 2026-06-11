# CLI Reference V1

## Purpose

This document defines the public CLI surface of RepoNerve.

The CLI is the primary interface for interacting with:

* Repository ingestion
* Repository memory
* Repository knowledge graph

This document acts as the source of truth for command structure and naming.

---

# Design Principles

## Human First

Commands should be easy to discover and remember.

Example:

```bash
reponerve memory list decisions
```

is preferred over deeply nested alternatives.

---

## Stable API

CLI commands are considered a public interface.

Breaking changes should be avoided.

---

## Consistent Verbs

RepoNerve uses the following verbs:

```text
init
scan
list
get
trace
explain
```

---

# Root Commands

## Initialize Workspace

```bash
reponerve init
```

Creates a RepoNerve workspace.

---

## Scan Repository

```bash
reponerve scan
```

Discovers repository sources and updates repository memory.

---

# Memory Commands

## Memory Root

```bash
reponerve memory
```

Parent command for repository memory operations.

---

# List Commands

## Events

```bash
reponerve memory list events
```

List repository events.

---

## Decisions

```bash
reponerve memory list decisions
```

List repository decisions.

---

## Intents

```bash
reponerve memory list intents
```

List repository intents.

---

## Facts

```bash
reponerve memory list facts
```

List repository facts.

---

## Relationships

```bash
reponerve memory list relationships
```

List repository relationships.

---

# Lookup Commands

## Event

```bash
reponerve memory get event <id>
```

Retrieve a single event.

---

## Decision

```bash
reponerve memory get decision <id>
```

Retrieve a single decision.

---

## Intent

```bash
reponerve memory get intent <id>
```

Retrieve a single intent.

---

## Fact

```bash
reponerve memory get fact <id>
```

Retrieve a single fact.

---

# Trace Commands

## Decision Trace

```bash
reponerve memory trace decision <id>
```

Traverse all relationships connected to a decision.

---

## Event Trace

```bash
reponerve memory trace event <id>
```

Traverse all relationships connected to an event.

---

## Intent Trace

```bash
reponerve memory trace intent <id>
```

Traverse all relationships connected to an intent.

---

# Explain Commands

## Decision Explanation

```bash
reponerve memory explain decision <id>
```

Generate a deterministic explanation of a decision.

---

## Event Explanation

```bash
reponerve memory explain event <id>
```

Generate a deterministic explanation of an event.

---

# Common Flags

## Repository Filter

Supported by list commands.

```bash
--repository <repository-id>
```

Example:

```bash
reponerve memory list decisions \
  --repository repo_123
```

---

# Output Philosophy

## List Commands

Tabular output.

Example:

```text
ID        STATUS      TITLE
------------------------------------
decision1 Accepted    Use Redis
decision2 Proposed    Adopt gRPC
```

---

## Get Commands

Detailed object view.

Example:

```text
Decision

ID:
decision1

Title:
Use Redis

Status:
Accepted
```

---

## Trace Commands

Tree-style output.

Example:

```text
Intent
└── Decision
    └── Event
```

---

## Explain Commands

Human-readable narrative output.

Example:

```text
Decision:
Use Redis

Reason:
Reduce Database Latency

Resulting Events:
- Introduce Redis Cache
```

---

# Search Command

```bash
reponerve search <query>
```

Deterministic repository knowledge search using FTS5 `memory_search` (rebuilt on each `scan`) plus keyword matching across decisions, facts, events, contributors, expertise, and relationships.

Supported query prefixes:

```text
type:decision
type:fact
type:event
type:contributor
type:expertise
type:relationship
domain:<name>
```

---

# Impact Command

```bash
reponerve impact <NODE_TYPE> <NODE_ID>
```

Analyze graph traversal impact for repository memory entities.

Supported `NODE_TYPE` values:

```text
DECISION
FACT
EVENT
CONTRIBUTOR
```

Graph impact is canonical for structural traversal. Memory-relationship impact used by `ask` is a separate analysis path.

---

# Ask Command

```bash
reponerve ask "<question>"
```

Deterministic repository Q&A using supported question patterns (onboarding, decision/event guidance, ownership, impact).

---

# Context Commands

```bash
reponerve context generate
reponerve context export --output <path>
```

Generate or export a structured repository context briefing.

---

# Version

Version: 1.0

Status: Draft
