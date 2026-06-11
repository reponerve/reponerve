# Agent-Native Repository Intelligence

Version: v1.1

Status: Draft

Updated: 2026-06-09

Related: `tasks/ARCH-001.md`, `docs/architecture/architecture-overview.md`

---

# Overview

RepoNerve is a software understanding platform built around knowledge preservation. Its purpose is to ensure that software understanding survives beyond individual contributors and remains accessible to both humans and AI systems.

Repository Intelligence and Code Intelligence are independent authorities. Repository-Code Linking connects them. Feature Understanding resolves feature-level understanding across both. Development Experience is the product surface. Software Understanding is the outcome.

---

# Architectural Pillars

```text
Knowledge Preservation          (foundation subsystem)
    ↓
Repository Intelligence         (why)
    +
Code Intelligence               (how)
    ↓
Repository-Code Linking           (cross-authority)
    ↓
Feature Understanding            (what — feature-level)
    ↓
Development Experience          (product surface)
    ↓
Software Understanding          (outcome)
```

The **Understanding Engine** retrieves and assembles context across all sources. It evolved from the Query Engine as the platform grew beyond repository memory.

---

# System Architecture

```text
Repository
        │
        ▼
┌─────────────────────┐
│ Repository Scanner  │
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│ Ingestion Pipeline  │
└───────┬─────┬───────┘
        │     │
        ▼     ▼
┌─────────────────┐   ┌─────────────────┐
│ Repository Int. │   │ Code Int.       │
└────────┬────────┘   └────────┬────────┘
         │                     │
         └─────────┬───────────┘
                   ▼
        ┌───────────────────┐
        │ Repository-Code   │
        │ Linking           │
        └─────────┬─────────┘
                  ▼
        ┌───────────────────┐
        │ Feature           │
        │ Intelligence      │
        └─────────┬─────────┘
                  ▼
        ┌───────────────────┐
        │ Development Exp.  │
        └─────────┬─────────┘
                  ▼
        ┌───────────────────┐
        │ Software          │
        │ Understanding     │
        └─────────┬─────────┘
                  │
                  ▼
              MCP / CLI
                  │
                  ▼
           AI Agents
```

All pillars rest on the **Knowledge Preservation Layer**.

---

# Layer Responsibilities

## Knowledge Preservation Layer

Foundation subsystem. Stores all durable software knowledge:

* Memory — Decisions, Facts, Events, Relationships
* Ownership — Expertise, domain ownership, reviewers
* Context — Repository context packages
* Code entities — Modules, packages, files, symbols, endpoints
* Repository-Code links — Cross-authority references

---

## Code Intelligence

Authoritative source for code understanding.

Answers:

* How does it work?
* Which files, packages, and symbols are involved?
* What is the call graph?
* Which symbols depend on this symbol?

Does not create repository memory entities.

Status: ISSUE-057.

---

## Repository Intelligence

Authoritative source for repository knowledge.

Answers:

* Why does this exist?
* Who owns it?
* What architectural decisions led to it?
* Who should review changes?
* What areas are impacted?

Does not parse or index code structure.

Status: Complete.

---

## Repository-Code Linking

Cross-authority subsystem connecting repository entities to code entities.

Example:

```text
Decision: ADR-004 OAuth
    → oauth.go
    → AuthService (struct)
    → LoginHandler (function)
```

Required for Development Experience. Without links, explain output cannot combine repository and code context.

Link types: `DECISION_REFERENCES_CODE`, `FACT_REFERENCES_CODE`, `EVENT_REFERENCES_CODE`.

Status: ISSUE-057.

---

## Feature Understanding

Emerging capability — not a separate authority.

Resolves feature-level understanding:

```text
Feature → Code → Ownership → Decisions → Impact
```

Orchestrated by Development Experience using Repository Search, Repository-Code links, Code Intelligence, and Repository Intelligence.

Examples: Authentication, Billing, Metadata Management.

Status: ISSUE-057 (v1.0).

---

## Development Experience

Orchestrates all intelligence sources into development-facing workflows.

Produces guidance through CLI and MCP.

Does not duplicate any intelligence authority.

Status: ISSUE-057.

---

## Understanding Engine

Retrieval and context assembly layer spanning:

* Repository Intelligence retrieval (Query Engine implementation)
* Code Intelligence retrieval
* Repository-Code link traversal
* Development context assembly
* Evidence collection

---

# Agent Consumption Model

```text
Agent Request
    ↓
MCP / CLI (Interface Layer)
    ↓
Development Experience
    ↓
Understanding Engine
    ↓
┌──────────────┬──────────────┬──────────────────┐
│ Repository   │ Code         │ Repository-Code  │
│ Intelligence │ Intelligence │ Links            │
└──────────────┴──────────────┴──────────────────┘
    ↓
Knowledge Preservation Layer
    ↓
Evidence-Backed Context Package
    ↓
Agent
```

Agents receive Software Understanding — not raw retrieval. Context packages combine code context, repository context, and cross-authority links with evidence.

---

# Existing Stack

The following already exists:

```text
Repository
    ↓
Ingestion
    ↓
Memory
    ↓
Context
    ↓
Ownership
    ↓
Knowledge Graph
    ↓
Repository Intelligence
    ↓
Query Engine (Understanding Engine — repository path)
    ↓
Agent Intelligence (packaging)
    ↓
MCP
```

Status: Repository Intelligence — Complete.

---

# v1.0 Extension

```text
Repository
    ↓
Ingestion
    ↓
Memory + Code Indexing
    ↓
Repository Intelligence + Code Intelligence
    ↓
Repository-Code Linking
    ↓
Feature Understanding (via Development Experience)
    ↓
Development Experience
    ↓
Understanding Engine (repository + code + links)
    ↓
MCP / CLI
    ↓
AI Agents
```

Status: Code Intelligence, Repository-Code Linking, Feature Understanding, Development Experience — ISSUE-057 (In Progress).

---

# Architectural Rules

## Dual Authority

* Code Intelligence → code understanding
* Repository Intelligence → repository knowledge
* Development Experience → orchestration
* Repository-Code Linking → cross-authority references

## No Duplication

Development Experience must reuse all existing Repository Intelligence services.

Code Intelligence must not duplicate repository memory extraction.

## Evidence Preservation

All outputs must preserve evidence and provenance.

## Determinism

The same repository state must produce identical outputs.

---

# Dependency Direction

```text
Knowledge Preservation Layer
    ↓
Readers / Code Stores / Link Store
    ↓
Repository Intelligence + Code Intelligence
    ↓
Repository-Code Linking
    ↓
Understanding Engine
    ↓
Development Experience
    ↓
MCP / CLI
```

Dependency direction must remain one-way.

---

# Explain Output Contract

Development Experience combines three context layers in explain output:

**Code Context** (from Code Intelligence)

* Files, packages, types, functions, endpoints
* Call graph, symbol dependencies

**Repository Context** (from Repository Intelligence)

* Decisions, facts, events
* Ownership, expertise, reviewers
* Impact, change plans

**Repository-Code Links** (from Repository-Code Linking)

* Decision → code entity references
* Fact → code entity references
* Event → code entity references

---

# Competitive Positioning

Many tools provide:

```text
Code Graph → Code Retrieval → LLM Context
```

Examples: GitNexus, Cortex, Code-Nexus, Codebase-Memory, Sourcegraph Cody.

RepoNerve provides:

```text
Knowledge Preservation
    +
Code Intelligence
    +
Repository Intelligence
    +
Repository-Code Linking
    +
Feature Understanding
    ↓
Software Understanding
```

This enables richer understanding with fewer tokens and less repository exploration.

## Differentiation

| Question type | Code-graph tools | RepoNerve |
| --- | --- | --- |
| Where is this symbol? | Strong | Strong (after ISSUE-057) |
| Why does this exist? | Weak | Strong (ADRs, decisions, links) |
| Who owns it? | Weak | Strong (ownership intelligence) |
| What breaks if I change it? | Graph impact | Graph + repository impact + decisions |
| Is this answer evidence-backed? | Varies | Mandatory |

## Token Economics

RepoNerve reduces agent token cost by:

1. Deterministic `scan` (zero LLM tokens for indexing)
2. Bounded MCP tool responses (structured, not raw files)
3. Pre-indexed memory (no per-session re-exploration)
4. Context packs with relevance ranking (ISSUE-060, v1.0)

See `docs/product/token-economics.md`.

## Composability

RepoNerve composes with — does not replace:

* **RTK** — shell output compression
* **LLM coding agents** — Cursor, Claude Code, Copilot (implementation)
* **Graphify-style discovery** — communities, surprises (ISSUE-061, v1.0)

Full competitive analysis: `docs/product/market-positioning.md`. Iteration plan: `docs/roadmap/v1.0-iteration-plan.md`.

---

# Release State

| Pillar | Status |
| --- | --- |
| Knowledge Preservation | Core Platform Capability |
| Repository Intelligence | ✅ Complete |
| Code Intelligence | ❌ ISSUE-057 |
| Repository-Code Linking | ❌ ISSUE-057 |
| Feature Understanding | ❌ ISSUE-057 |
| Development Experience | ❌ ISSUE-057 |
| Software Understanding | Blocked |

v1.0 release is blocked until ISSUE-057 is complete.

---

# Summary

RepoNerve is a software understanding platform — not a repository memory tool. Knowledge Preservation is the foundation. Repository Intelligence and Code Intelligence are independent authorities connected by Repository-Code Linking. Feature Understanding and Development Experience deliver understanding to humans and AI agents with determinism, explainability, and evidence-based reasoning.
