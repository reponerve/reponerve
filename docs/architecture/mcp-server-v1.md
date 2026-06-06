# MCP Server V1

## Purpose

Expose RepoNerve repository intelligence through the Model Context Protocol (MCP).

This allows AI coding agents to query repository memory, relationships, and context directly.

Supported clients may include:

* Claude Code
* Cursor
* Windsurf
* Cline
* Roo
* Codex
* OpenAI Agents
* Future MCP-compatible clients

---

# Goals

Provide a standardized interface for accessing:

* Repository Memory
* Repository Relationships
* Repository Context

without direct database access.

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
MCP Server
↓
AI Agents

---

# Design Principles

## Read-Only V1

The MCP Server must not mutate repository memory.

Allowed:

* Read
* Query
* Context generation

Not Allowed:

* Writes
* Deletes
* Updates

---

## Reuse Existing Engines

The MCP layer must consume:

* Query Engine
* Context Engine

and must not duplicate logic.

---

## Thin Transport Layer

Business logic remains in:

* Memory Engine
* Query Engine
* Context Engine

MCP is only a transport mechanism.

---

# MCP Resources

## Repository Context

Returns:

RepositoryContext

Equivalent to:

reponerve context generate

---

## Decisions

Returns repository decisions.

Equivalent to:

reponerve memory list decisions

---

## Intents

Returns repository intents.

Equivalent to:

reponerve memory list intents

---

## Facts

Returns repository facts.

Equivalent to:

reponerve memory list facts

---

## Events

Returns repository events.

Equivalent to:

reponerve memory list events

---

# MCP Tools

## get_decision

Input:

decision_id

Output:

Decision

---

## get_event

Input:

event_id

Output:

Event

---

## get_intent

Input:

intent_id

Output:

Intent

---

## get_fact

Input:

fact_id

Output:

Fact

---

## trace_decision

Input:

decision_id

Output:

Decision Trace

---

## trace_event

Input:

event_id

Output:

Event Trace

---

## explain_decision

Input:

decision_id

Output:

Decision Explanation

---

## explain_event

Input:

event_id

Output:

Event Explanation

---

# Security

V1 assumes local execution.

Authentication and authorization are out of scope.

---

# Constraints

Do NOT implement:

* Memory mutations
* AI summarization
* Embeddings
* Vector search
* Ownership extraction

These remain future roadmap items.

---

# Version

Version: 1.0

Status: Draft
