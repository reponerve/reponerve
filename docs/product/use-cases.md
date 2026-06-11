# RepoNerve Use Cases

Version: 1.0

Status: Draft

Authors: RepoNerve Contributors

Last Updated: 2026-06-11

---

# Purpose

This document defines the primary use cases RepoNerve is designed to solve.

Use cases represent real-world workflows that users and AI systems perform while interacting with software repositories.

These use cases drive:

* Product design
* CLI design
* API design
* MCP tool design
* Memory extraction requirements
* Context engine requirements

---

# Use Case Categories

RepoNerve supports three major categories:

1. Repository Understanding
2. Repository Memory
3. Repository Context

---

# Category 1: Repository Understanding

Repository understanding focuses on helping users learn how software systems work.

---

## UC-001: New Developer Onboarding

### Actor

Individual Developer

### Goal

Understand an unfamiliar repository.

### Scenario

A developer joins a project for the first time.

The repository contains:

* Thousands of files
* Multiple services
* Years of history

The developer needs to become productive quickly.

---

### Current Workflow

The developer:

* Reads documentation
* Reviews code
* Searches pull requests
* Asks teammates

This process is slow.

---

### RepoNerve Workflow

```bash
reponerve explain repository
```

RepoNerve provides:

* Repository overview
* Major components
* Architecture summary
* Ownership information
* Important decisions

---

### Success Criteria

The developer gains repository understanding significantly faster.

---

## UC-002: Explain Component

### Actor

Developer

### Goal

Understand a specific component.

### Scenario

A developer encounters:

```text
services/auth
```

and wants to understand:

* Purpose
* History
* Ownership
* Dependencies

---

### RepoNerve Workflow

```bash
reponerve explain services/auth
```

---

### Expected Output

* Component purpose
* Historical evolution
* Related decisions
* Ownership information
* Related ADRs

---

# Category 2: Repository Memory

Repository memory focuses on historical reasoning and intent.

---

## UC-003: Why Does This Exist?

### Actor

Developer

### Goal

Understand why a component was created.

### Scenario

A developer finds:

```text
Redis Cache Layer
```

and asks:

Why does this exist?

---

### RepoNerve Workflow

```bash
reponerve ask "Why was Redis introduced?"
```

---

### Expected Output

* Decision
* Reason
* Alternatives considered
* Related PRs
* Related ADRs
* Historical context

---

## UC-004: Historical Decision Lookup

### Actor

Staff Engineer

### Goal

Understand architectural decisions.

### Scenario

A team wants to understand why a technology was selected.

---

### RepoNerve Workflow

```bash
reponerve ask "Why was Kafka selected?"
```

---

### Expected Output

* Decision record
* Alternatives
* Tradeoffs
* Sources

---

## UC-005: Ownership Discovery

### Actor

Developer

### Goal

Identify system ownership.

### Scenario

A service requires changes.

The developer needs to know who owns it.

---

### RepoNerve Workflow

```bash
reponerve ask "Who owns billing?"
```

---

### Expected Output

* Team ownership
* Historical ownership
* Related contacts
* Source evidence

---

## UC-006: Incident Investigation

### Actor

Staff Engineer

### Goal

Understand repository changes related to an incident.

### Scenario

An incident occurs.

The team wants to understand:

* Related changes
* Related decisions
* Historical context

---

### RepoNerve Workflow

```bash
reponerve ask "What changed after incident INC-42?"
```

---

### Expected Output

* Related commits
* Related pull requests
* Related decisions
* Impacted components

---

# Category 3: Repository Context

Repository context focuses on development acceleration.

---

## UC-007: Generate Context Pack

### Actor

Developer

### Goal

Gather relevant context before implementation.

### Scenario

The developer receives a task:

```text
Add MFA support.
```

---

### RepoNerve Workflow

```bash
reponerve context "Add MFA support"
```

---

### Expected Output

* Relevant services
* Relevant files
* Existing patterns
* Related decisions
* Related ADRs
* Similar implementations

---

### Success Criteria

The developer spends less time searching.

---

## UC-008: AI Agent Context Retrieval

### Actor

AI Coding Agent

### Goal

Obtain repository-specific context.

### Scenario

An AI agent receives a development task.

Before generating code, it requests repository memory.

---

### RepoNerve Workflow

```text
get_context_pack
```

---

### Expected Output

* Relevant repository memory
* Relevant patterns
* Relevant decisions
* Related implementations

---

### Success Criteria

The AI consumes less context and produces more accurate outputs.

---

## UC-009: Similar Change Discovery

### Actor

Developer

### Goal

Find similar work completed in the past.

### Scenario

The developer wants to implement a feature.

A similar feature may already exist.

---

### RepoNerve Workflow

```bash
reponerve ask "Show similar MFA implementations"
```

---

### Expected Output

* Similar pull requests
* Similar commits
* Related components
* Relevant patterns

---

# MCP Use Cases

These use cases become important in future releases.

---

## UC-010: Repository Memory Skill

### Actor

AI Agent

### Goal

Access repository memory.

---

### MCP Tool

```text
get_repository_memory
```

---

### Response

* Facts
* Events
* Decisions
* Ownership
* Intent

---

## UC-011: Explain Component Skill

### Actor

AI Agent

### Goal

Understand repository components.

---

### MCP Tool

```text
explain_component
```

---

### Response

* Purpose
* Dependencies
* Ownership
* Historical context

---

## UC-012: Context Pack Skill

### Actor

AI Agent

### Goal

Obtain task-specific repository context.

---

### MCP Tool

```text
get_context_pack
```

---

### Response

* Relevant files
* Relevant patterns
* Relevant decisions
* Related ADRs
* Similar implementations

---

# MVP Use Cases

The following use cases define MVP scope:

* UC-001: New Developer Onboarding
* UC-002: Explain Component
* UC-003: Why Does This Exist?
* UC-004: Historical Decision Lookup
* UC-007: Generate Context Pack

---

# Future Use Cases

The following use cases are post-MVP:

* UC-006: Incident Investigation
* UC-008: AI Agent Context Retrieval
* UC-009: Similar Change Discovery
* UC-010: Repository Memory Skill
* UC-011: Explain Component Skill
* UC-012: Context Pack Skill

---

# Category 4: Token-Efficient AI Development

Use cases focused on reducing LLM token consumption while preserving answer quality.

---

## UC-013: Agent Session Within Context Limits

### Actor

AI Coding Agent (Cursor, Claude Code, Copilot)

### Goal

Complete implementation tasks without exhausting context limits on repository exploration.

### RepoNerve Workflow

1. `reponerve scan` (zero LLM tokens)
2. Agent connects via MCP
3. Agent calls bounded tools (`explain`, `analyze_impact`, `trace_graph`) instead of bulk file reads

### Success Criteria

* 80%+ reduction in exploration tokens per task
* Same or better answer quality with evidence

See `docs/product/token-economics.md`.

---

## UC-014: Premium Model Cost Control

### Actor

Engineering Team

### Goal

Use premium models for implementation while avoiding repeated archaeology costs.

### Success Criteria

* Understanding delivered before LLM reasoning begins
* Measurable token savings per sprint

---

# Category 5: Greenfield Development

Use cases for repositories built from scratch.

---

## UC-015: Memory From First Commit

### Actor

Founding Developer or Agent-Assisted Greenfield Team

### Goal

Preserve architectural decisions from project inception.

### RepoNerve Workflow

```bash
reponerve init
# Write ADR-0001 before or with first code
git commit
reponerve scan
reponerve mcp
```

### Success Criteria

* No "documentation debt" accumulation
* Session N retains understanding from session 1

See `docs/product/greenfield-guide.md`.

---

## UC-016: Plan New Feature on Young Codebase

### Actor

Developer

### Goal

Get implementation guidance before writing code.

### RepoNerve Workflow

```bash
reponerve plan "Add OAuth Login"
```

### Success Criteria

* Starting points, impacted areas, and linked decisions returned with evidence

---

# Category 6: Market Differentiation

---

## UC-017: Evidence-Backed "Why" Questions

### Goal

Answer why questions with ADR + commit + code links — not LLM inference alone.

### Competitor gap

Code-graph tools answer *where*; RepoNerve answers *why* with mandatory evidence.

See `docs/product/market-positioning.md`.

---

# Guiding Principle

RepoNerve should eliminate unnecessary repository rediscovery.

If a user or AI system has already learned something once, RepoNerve should make that knowledge accessible instead of forcing it to be rediscovered.
