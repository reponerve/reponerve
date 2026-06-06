# Context Engine V1

## Purpose

The Context Engine converts repository memory into structured context consumable by AI coding agents.

The Memory Engine stores repository knowledge.

The Query Engine retrieves repository knowledge.

The Context Engine assembles repository knowledge into concise, high-signal context.

---

# Goals

Enable AI agents to quickly understand:

* Repository purpose
* Important decisions
* Repository intent
* Architectural facts
* Recent events

without scanning the entire repository.

---

# Architecture

Repository
↓
Memory Engine
↓
Query Engine
↓
Context Engine
↓
Agent Context

---

# Context Sections

## Repository Summary

High-level repository overview.

---

## Key Decisions

Most important decisions.

Examples:

* Use Redis Cache
* Adopt gRPC

---

## Key Intents

Repository motivations.

Examples:

* Reduce Latency
* Improve Reliability

---

## Key Facts

Examples:

* Auth Service USES Redis
* API Gateway DEPENDS_ON Auth Service

---

## Recent Events

Examples:

* Introduce Redis Cache
* Refactor Authentication Flow

---

# Output Format

Markdown.

Example:

# Repository Context

## Key Decisions

...

## Key Intents

...

## Key Facts

...

## Recent Events

...

---

# Constraints

V1 is deterministic.

No AI.

No summarization models.

No embeddings.

No vector databases.

---

# Version

Version: 1.0

Status: Draft
