# RepoNerve Context Engine

Version: 1.0

Status: Draft

Authors: RepoNerve Contributors

Last Updated: 2026-06-05

---

# Purpose

This document defines the Context Engine architecture within RepoNerve.

The Context Engine transforms repository memory into task-specific context that can be consumed by:

* Developers
* AI coding assistants
* MCP clients
* Automation systems
* Future RepoNerve integrations

The Context Engine is responsible for ensuring that consumers receive only the information required to solve a problem.

---

# Problem Statement

Modern development workflows repeatedly rediscover repository knowledge.

Example:

A developer receives a task:

```text id="3u9i4g"
Add MFA support.
```

To complete the task, they must discover:

* Relevant services
* Related APIs
* Existing patterns
* Historical decisions
* Ownership information
* Previous implementations

AI coding systems face the same problem.

Most AI systems spend significant resources understanding repositories before generating solutions.

This results in:

* Excessive token usage
* Slower execution
* Reduced accuracy
* Duplicate repository exploration

---

# Vision

The Context Engine should eliminate unnecessary repository rediscovery.

Instead of searching repositories repeatedly, users and AI systems should consume repository memory that has already been extracted and validated.

---

# Core Principle

Memory is the source.

Context is the output.

The Context Engine never creates new knowledge.

The Context Engine assembles relevant knowledge from repository memory.

---

# High-Level Architecture

```text id="d3tukr"
User Request
      │
      ▼
Intent Analyzer
      │
      ▼
Memory Retrieval
      │
      ▼
Relationship Expansion
      │
      ▼
Context Ranking
      │
      ▼
Context Assembly
      │
      ▼
Context Pack
```

---

# Definitions

---

## Memory

Repository knowledge stored by RepoNerve.

Examples:

* Facts
* Decisions
* Events
* Intent
* Ownership

---

## Context

A curated subset of memory relevant to a task.

---

## Context Pack

A structured bundle of repository knowledge delivered to a consumer.

---

# Context Engine Responsibilities

The Context Engine must:

* Identify relevant repository memory
* Retrieve supporting evidence
* Expand relationships
* Remove irrelevant information
* Minimize context size
* Maximize context quality

---

# Inputs

---

## Developer Requests

Examples:

```text id="0ohy5n"
Add MFA support

Explain authentication

Refactor billing service
```

---

## AI Agent Requests

Examples:

```text id="4ql6gd"
Generate implementation plan

Find similar implementation

Explain repository patterns
```

---

## MCP Requests

Examples:

```text id="l4xygo"
get_context_pack

get_repository_memory

explain_component
```

---

# Outputs

The Context Engine produces:

```text id="34oj0x"
Context Pack
```

---

Example:

```text id="vq8sl8"
Relevant Services

Relevant Files

Relevant Decisions

Relevant ADRs

Relevant Patterns

Relevant Ownership

Related Implementations
```

---

# Context Pack Structure

## Metadata

```json id="e2b6s7"
{
  "task": "Add MFA support",
  "generated_at": "...",
  "repository": "..."
}
```

---

## Relevant Components

```json id="g87drs"
[
  "AuthService",
  "UserService"
]
```

---

## Relevant Decisions

```json id="a4r9r6"
[
  {
    "decision": "Use JWT",
    "reason": "Stateless authentication"
  }
]
```

---

## Relevant Sources

```json id="z6xjqu"
[
  "ADR-12",
  "PR-143"
]
```

---

# Context Generation Pipeline

---

## Stage 1

Intent Detection

---

Input:

```text id="rntkkw"
Add MFA support
```

---

Output:

```text id="10j2t4"
Authentication

Security

User Login
```

---

Purpose:

Identify repository domains.

---

# Stage 2

Memory Retrieval

---

Retrieve:

* Decisions
* Facts
* Events
* Ownership
* Intent

---

Example:

```text id="hl4g4v"
Authentication Decisions

Authentication ADRs

Authentication Services
```

---

# Stage 3

Relationship Expansion

---

Expand related memories.

---

Example:

```text id="k4vknq"
Auth Service
      │
      ▼
Uses JWT
      │
      ▼
Related ADR
      │
      ▼
Related PR
```

---

Purpose:

Gather complete context.

---

# Stage 4

Ranking

---

Not all memory is equally important.

Rank based on:

* Relevance
* Recency
* Confidence
* Evidence quality

---

Example:

```text id="lyqzkn"
ADR-12

Score: 98
```

---

```text id="wql7tr"
Old Discussion

Score: 12
```

---

# Stage 5

Assembly

---

Create final context pack.

---

Output:

```text id="cnvytq"
Context Pack
```

---

# Context Ranking Strategy

The ranking engine should prioritize:

---

## Direct Relevance

Highest priority.

---

Example:

```text id="4qj2kx"
Authentication ADR
```

for:

```text id="8k6cv8"
Add MFA support
```

---

## Historical Decisions

Very high priority.

---

## Existing Patterns

High priority.

---

## Ownership

Medium priority.

---

## General Repository Facts

Lower priority.

---

# Context Types

---

## Decision Context

Examples:

* Technology selection
* Architecture decisions
* Tradeoffs

---

## Historical Context

Examples:

* Related pull requests
* Related incidents
* Related discussions

---

## Ownership Context

Examples:

* Teams
* Maintainers
* Contributors

---

## Pattern Context

Future release.

Examples:

* Authentication patterns
* Logging patterns
* API conventions

---

# AI Token Optimization Strategy

This is one of RepoNerve's primary goals.

---

# Current Workflow

```text id="8od9w8"
Task
      │
      ▼
AI Searches Repository
      │
      ▼
Large Context
      │
      ▼
High Token Usage
```

---

# RepoNerve Workflow

```text id="t61txf"
Task
      │
      ▼
RepoNerve
      │
      ▼
Context Pack
      │
      ▼
AI
```

---

Benefits:

* Lower token usage
* Faster execution
* Better repository understanding
* More consistent implementations

---

# MCP Integration

The Context Engine is the primary MCP capability.

---

## Tool: get_context_pack

Input:

```json id="8n4a2r"
{
  "task": "Add MFA support"
}
```

---

Output:

```json id="dfghzk"
{
  "context": [...]
}
```

---

## Tool: explain_component

Input:

```json id="6zovsj"
{
  "component": "AuthService"
}
```

---

Output:

```json id="p59r6e"
{
  "purpose": "...",
  "history": "...",
  "decisions": [...]
}
```

---

## Tool: find_related_decisions

Input:

```json id="4i4s1r"
{
  "topic": "authentication"
}
```

---

Output:

Relevant decisions and evidence.

---

# Context Quality Metrics

The Context Engine should optimize for:

---

## Precision

Relevant information returned.

---

## Compression

Minimal unnecessary information.

---

## Explainability

Every context item includes evidence.

---

## Reusability

Context should be reusable across consumers.

---

# Future Context Packs

Future versions may support:

---

## Development Context Pack

Implementation-focused.

---

## Review Context Pack

Code review-focused.

---

## Architecture Context Pack

Architecture-focused.

---

## Incident Context Pack

Operational troubleshooting.

---

## Refactoring Context Pack

Technical debt and modernization.

---

# Success Criteria

The Context Engine succeeds when:

* Developers spend less time searching.
* AI systems consume fewer tokens.
* Context quality improves implementation accuracy.
* Repository memory becomes immediately actionable.

---

# Guiding Principle

Do not send consumers the repository.

Send consumers the knowledge they actually need.
