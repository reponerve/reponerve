# Query Engine V1

## Purpose

The Query Engine provides a read layer over the RepoNerve Memory Graph.

The Memory Engine is responsible for producing memories.

The Query Engine is responsible for retrieving, navigating, and explaining memories.

---

# Goals

Enable users and AI agents to answer:

* What happened?
* What decisions were made?
* Why was a decision made?
* What facts support a decision?
* Which events resulted from a decision?
* Trace a memory through the graph.

---

# Architecture

```text
Repository
    ↓
Sources
    ↓
Memory Engine
    ↓
Events
Decisions
Intents
Facts
Relationships
    ↓
Query Engine
```

---

# Query Types

## List Queries

Return collections of memories.

Examples:

reponerve memory list events

reponerve memory list decisions

reponerve memory list intents

reponerve memory list facts

---

## Lookup Queries

Return a single memory.

Examples:

reponerve memory get decision <id>

reponerve memory get event <id>

---

## Trace Queries

Traverse relationships.

Examples:

reponerve memory trace decision <id>

reponerve memory trace event <id>

---

## Explain Queries

Generate a human-readable explanation.

Examples:

reponerve memory explain decision <id>

reponerve memory explain event <id>

---

# Query Layer

The query layer must be read-only.

It must never modify:

* Sources
* Memories
* Relationships

---

# Query Stores

Create dedicated read interfaces.

Examples:

EventReader

DecisionReader

IntentReader

FactReader

RelationshipReader

---

# Future

V1 intentionally excludes:

* Natural language queries
* LLM integration
* Semantic search
* Embeddings
* Vector databases

These belong in future releases.
